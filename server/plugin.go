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

	email := "email"
	idB, err := p.API.KVGet(email)
	if idB == nil || err != nil {
		p.API.LogError("Cannot get user by email", "email", email)

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	intent := dfr.Intent.Name
	var response *OutgoingResponse
	if *intent == "change_status" {
		var nErr error
		response, nErr = p.handleStatusChange(*dfr.Intent.Params.Status.Resolved, string(idB))
		if nErr != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	} else if *intent == "send_dm" {
		var nErr error
		response, nErr = p.handleSendDM(string(idB), *dfr.Intent.Params.OtherUser.Resolved, *dfr.Intent.Params.Message.Resolved)
		if nErr != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	} else {
		response = getResponseWithText("Sorry, don't know what to do!")
	}

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
func (p *Plugin) emailFromUserId(userId string) string {
	keys, _ := p.API.KVList(0, 100)
	for _, key := range keys {
		if v, _ := p.API.KVGet(key); v != nil && string(v) == userId {
			return key
		}
	}
	return ""
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
			if len(parts) != 3 {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "Syntax: /assistant connect you@gmail.com",
				}, nil
			}

			email := parts[2]
			p.API.KVCompareAndSet(email, []byte(args.UserId), []byte(args.UserId))
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Connected!",
			}, nil
		} else if parts[1] == "disconnect" {
			p.API.KVDelete(p.emailFromUserId(args.UserId))

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
