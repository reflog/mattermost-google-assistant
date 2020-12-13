package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func getResponseWithText(s string) *OutgoingResponse {
	return &OutgoingResponse{
		Prompt: &gPrompt{
			Override: false,
			LastSimple: &gSimple{
				Speech: &s,
				Text:   s,
			},
		},
	}
}

func (p *Plugin) handleSendDM(myUid, targetUsername, message string) (*OutgoingResponse, error) {
	ou, err := p.API.GetUserByUsername(targetUsername)
	if err != nil {
		p.API.LogError("Cannot get other user", "err", err.Error())
		return nil, err
	}
	dc, err := p.API.GetDirectChannel(myUid, ou.Id)
	if err != nil {
		p.API.LogError("Cannot create dm channel", "err", err.Error())
		return nil, err
	}
	_, err = p.API.CreatePost(&model.Post{
		ChannelId: dc.Id,
		UserId:    myUid,
		Message:   message,
	})
	if err != nil {
		p.API.LogError("Cannot create post", "err", err.Error())
		return nil, err
	}
	return getResponseWithText("Message sent!"), nil
}

func (p *Plugin) handleStatusChange(newStatus, uid string) (*OutgoingResponse, error) {
	oldStatus, err := p.API.GetUserStatus(uid)
	if err != nil {
		p.API.LogError("Cannot get status decode", "err", err.Error())
		return nil, err
	}
	_, err = p.API.UpdateUserStatus(uid, newStatus)
	if err != nil {
		p.API.LogError("Cannot update status", "err", err.Error())
		return nil, err
	}
	return getResponseWithText(fmt.Sprintf("Changing status from %s to %s", oldStatus.Status, newStatus)), nil
}
func (p *Plugin) handleReadMessages(uid string) (*OutgoingResponse, error) {
	teamUnreads, err := p.API.GetTeamsUnreadForUser(uid)
	if err != nil {
		p.API.LogError("Cannot get unread", "err", err.Error())
		return nil, err
	}

	messages := []string{}
	dms := make(map[string]bool)
	for _, teamUnread := range teamUnreads {
		cms, err := p.API.GetChannelMembersForUser(teamUnread.TeamId, uid, 0, 100)
		if err != nil {
			p.API.LogError("Cannot get members", "err", err.Error())
			return nil, err
		}
		for _, cm := range cms {
			if cm.MentionCount > 0 {
				c, _ := p.API.GetChannel(cm.ChannelId)
				if c.Type == model.CHANNEL_DIRECT {
					oid := c.GetOtherUserIdForDM(uid)
					ou, _ := p.API.GetUser(oid)
					pl, _ := p.API.GetPostsForChannel(cm.ChannelId, 0, 100)
					pl.SortByCreateAt()
					p := pl.Posts[pl.Order[0]]
					dms[fmt.Sprintf("'%s' wrote '%s'.", ou.Username, p.Message)] = true
				}
			}

		}
	}
	if len(dms) == 0 {
		messages = append(messages, "You have no unread DMs")
	} else {
		messages = append([]string{"Here are your messages:"})
		for m := range dms {
			messages = append(messages, m)
		}
	}
	return getResponseWithText(strings.Join(messages, "\n")), nil
}
func (p *Plugin) handleGetStatus(uid string) (*OutgoingResponse, error) {
	oldStatus, err := p.API.GetUserStatus(uid)
	if err != nil {
		p.API.LogError("Cannot get status", "err", err.Error())
		return nil, err
	}
	teamUnreads, err := p.API.GetTeamsUnreadForUser(uid)
	if err != nil {
		p.API.LogError("Cannot get unread", "err", err.Error())
		return nil, err
	}
	teams, err := p.API.GetTeamsForUser(uid)
	if err != nil {
		p.API.LogError("Cannot get teams", "err", err.Error())
		return nil, err
	}
	teamById := func(id string) *model.Team {
		for _, team := range teams {
			if team.Id == id {
				return team
			}
		}
		return nil
	}
	messages := []string{fmt.Sprintf("Your current status is '%s'.", oldStatus.Status)}
	for _, teamUnread := range teamUnreads {
		team := teamById(teamUnread.TeamId)
		if team == nil {
			continue
		}
		messages = append(messages, fmt.Sprintf("In team '%s' you have %d unread messages and had %d mentions.", team.DisplayName, teamUnread.MsgCount, teamUnread.MentionCount))
	}
	return getResponseWithText(strings.Join(messages, "\n")), nil
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var dfr IncomingRequest
	if err := json.NewDecoder(r.Body).Decode(&dfr); err != nil {
		p.API.LogError("Cannot decode", "err", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	p.API.LogInfo("user", "t", dfr.Intent.Params.Username)
	if dfr.Intent.Params.Username == nil || *dfr.Intent.Params.Username.Resolved == "" {

	}

	var response *OutgoingResponse
	validateUser := func() string {
		if dfr.User.Params.UserName == nil || *dfr.User.Params.UserName == "" {
			response = getResponseWithText("Sorry, you didn't set your mattermost username!")
			return ""
		}
		idB, err := p.API.KVGet(*dfr.User.Params.UserName)
		if err != nil || idB == nil {
			response = getResponseWithText("Sorry, you didn't enable google assistant integration!")
			return ""
		}
		u, _ := p.API.GetUserByUsername(*dfr.User.Params.UserName)

		return u.Id
	}

	handler := *dfr.Handler.Name
	// handler = "set_username"
	switch handler {
	case "get_status":
		{
			userId := validateUser()
			if userId == "" {
				break
			}
			var nErr error
			response, nErr = p.handleGetStatus(userId)
			if nErr != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
	case "read_direct_messages":
		{
			userId := validateUser()
			if userId == "" {
				break
			}
			var nErr error
			response, nErr = p.handleReadMessages(userId)
			if nErr != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
	case "change_status":
		{
			userId := validateUser()
			if userId == "" {
				break
			}
			var nErr error
			response, nErr = p.handleStatusChange(*dfr.Intent.Params.Status.Resolved, userId)
			if nErr != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
	case "set_username":
		{
			response = &OutgoingResponse{
				User: &gUser{
					Params: gUserParams{
						UserName: dfr.Intent.Params.Username.Resolved, //model.NewString("sysadmin"),
					},
				},
				Prompt: &gPrompt{},
			}
		}
	case "send_message":
		{
			userId := validateUser()
			if userId == "" {
				break
			}
			var nErr error
			response, nErr = p.handleSendDM(userId, dfr.Scene.Slots.Username.Value, *dfr.Intent.Params.Message.Resolved)
			if nErr != nil {
				response = getResponseWithText("Sorry, can't find that user!")
			}
		}
	default:
		{
			response = getResponseWithText("Sorry, don't know what to do!")
		}
	}
	suggestions := []gSuggestions{
		{Title: "Change status to away"},
		{Title: "Status Report"},
		{Title: "Read messages"},
		{Title: "Write message"},
	}
	response.Prompt.Suggestions = &suggestions
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func getAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData("assistant", "", "Enables or disables assistant intgeration.")
	command.AddStaticListArgument("", true, []model.AutocompleteListItem{
		{
			Item:     "connect",
			HelpText: "Connect Google Assistant account",
		}, {
			Item:     "disconnect",
			HelpText: "Disconnect Google Assistant account",
		},
	})

	return command
}

func (p *Plugin) returnHelp() (*model.CommandResponse, *model.AppError) {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Only connect/disconnect commands are supported!",
	}, nil
}
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	parts := strings.Fields(args.Command)
	trigger := strings.TrimPrefix(parts[0], "/")
	if trigger == "assistant" {
		if len(parts) < 2 {
			return p.returnHelp()
		}
		if parts[1] == "connect" {
			u, _ := p.API.GetUser(args.UserId)
			p.API.KVSet(u.Username, []byte("true"))
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Connected!",
			}, nil
		} else if parts[1] == "disconnect" {
			u, _ := p.API.GetUser(args.UserId)
			p.API.KVDelete(u.Username)

			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Disconnected!",
			}, nil
		} else {
			return p.returnHelp()
		}

	}
	return &model.CommandResponse{}, nil
}

func (p *Plugin) OnActivate() error {
	p.API.RegisterCommand(&model.Command{
		Trigger:          "assistant",
		AutoComplete:     true,
		AutoCompleteHint: "(connect|disconnect)",
		AutoCompleteDesc: "Google Assistant for Mattermost",
		AutocompleteData: getAutocompleteData(),
	})
	return nil
}
