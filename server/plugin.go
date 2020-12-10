package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	df "github.com/leboncoin/dialogflow-go-webhook"
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

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var dfr df.Request
	if err := json.NewDecoder(r.Body).Decode(&dfr); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	fmt.Printf("%+v\n", dfr)
	// Filter on action, using a switch for example

	// Retrieve the params of the request

	if err := dfr.GetParams(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Retrieve a specific context
	if err := dfr.GetContext("my-awesome-context", &p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Do things with the context you just retrieved
	dff := &df.Fulfillment{
		FulfillmentMessages: df.Messages{
			df.ForGoogle(df.SingleSimpleResponse("hello", "hello")),
			{RichMessage: df.Text{Text: []string{"hello"}}},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dff)

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

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	parts := strings.Fields(args.Command)
	trigger := strings.TrimPrefix(parts[0], "/")
	if trigger == "assistant" {
		if parts[1] == "connect" {
			if len(parts) != 4 {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "Syntax: /assistant connect you@gmail.com mm_personal_token",
				}, nil
			}

			email := parts[2]
			token := parts[3]
			v, _ := json.Marshal(map[string]string{"email": email, "token": token})
			p.API.KVSetWithOptions(args.UserId, v, model.PluginKVSetOptions{})
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Connected!",
			}, nil
		} else if parts[1] == "disconnect" {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Disconnected!",
			}, nil
		} else {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Only connect/disconnect commands are supported!",
			}, nil
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
