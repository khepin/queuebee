package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
	"github.com/tidwall/gjson"
)

func (app *App) HandleInteractivity(c echo.Context) error {
	params, err := app.Slack.ParseInteractivityRequest(c.Request())
	if err != nil {
		app.Logger.Error("Could not parse interactivity command", "error", err)
		return c.JSON(500, "Could not parse interactivity command")
	}
	err = app.HandleAction(params)
	if err != nil {
		app.Logger.Error("Could not handle action", "error", err)
		return c.JSON(500, err)
	}

	return c.String(204, "")
}

func (app *App) HandleSlashCommand(c echo.Context) error {
	params, err := app.Slack.ParseCommandRequest(c.Request())
	if err != nil {
		app.Logger.Error("Could not parse command", "error", err)
		return c.JSON(500, "Could not parse command")
	}

	err = app.HandleAction(params)
	if err != nil {
		app.Logger.Error("Could not handle action", "error", err)
		return c.JSON(500, err)
	}

	return c.JSON(204, "")
}

func (app *App) HandleAction(params QueueActionParams) error {
	switch params.Action {
	case "join":
		return joinQueue(app, params)
	case "leave":
		return leaveQueue(app, params)
	case "skip":
		return skipTurn(app, params)
	case "list":
		return listQueue(app, params)
	case "ack":
		return ackTurn(app, params)
	case "config":
		return configQueue(app, params)
	case "save_config":
		return saveConfig(app, params)
	case "help":
		return help(app, params)
	default:
		return fmt.Errorf("unknown command %s", params.Action)
	}
}

type QueueActionParams struct {
	Action      string
	UserID      string
	ChannelID   string
	TeamID      string
	ChannelName string
	TeamDomain  string
	TriggerID   string
	ActionValue string
}

func SlurpSlashCommand(slashCmd slack.SlashCommand) QueueActionParams {
	commandText := strings.TrimSpace(slashCmd.Text)
	action := strings.Split(commandText, " ")
	return QueueActionParams{
		Action:      action[0],
		UserID:      slashCmd.UserID,
		ChannelID:   slashCmd.ChannelID,
		TeamID:      slashCmd.TeamID,
		ChannelName: slashCmd.ChannelName,
		TeamDomain:  slashCmd.TeamDomain,
		TriggerID:   slashCmd.TriggerID,
	}
}

func SlurpInteractiveAction(payload string) QueueActionParams {
	if gjson.Get(payload, "type").String() == "view_submission" {
		meta := gjson.Get(payload, "view.private_metadata").String()
		return QueueActionParams{
			ChannelID:   gjson.Get(meta, "channel_id").String(),
			ChannelName: gjson.Get(meta, "channel_name").String(),
			Action:      gjson.Get(payload, "view.callback_id").String(),
			UserID:      gjson.Get(payload, "user.id").String(),
			TeamID:      gjson.Get(payload, "team.id").String(),
			TeamDomain:  gjson.Get(payload, "team.domain").String(),
			TriggerID:   gjson.Get(payload, "trigger_id").String(),
			ActionValue: gjson.Get(payload, "view.state.values").String(),
		}
	}
	return QueueActionParams{
		Action:      readUntilColon(gjson.Get(payload, "actions.0.action_id").String()),
		UserID:      gjson.Get(payload, "user.id").String(),
		ChannelID:   gjson.Get(payload, "channel.id").String(),
		TeamID:      gjson.Get(payload, "team.id").String(),
		ChannelName: gjson.Get(payload, "channel.name").String(),
		TeamDomain:  gjson.Get(payload, "team.domain").String(),
		TriggerID:   gjson.Get(payload, "trigger_id").String(),
		ActionValue: gjson.Get(payload, "actions.0.value").String(),
	}
}

func (qb QueueActionParams) QueueID() string {
	return qb.TeamID + ":" + qb.ChannelID
}

func (s *Slack) ParseInteractivityRequest(r *http.Request) (QueueActionParams, error) {
	verifier, err := slack.NewSecretsVerifier(r.Header, s.SigningSecret)
	if err != nil {
		return QueueActionParams{}, fmt.Errorf("could not create verifier: %v", err)
	}

	r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
	err = r.ParseForm()
	if err != nil {
		return QueueActionParams{}, fmt.Errorf("could not parse interactivity command: %v", err)
	}
	if err := verifier.Ensure(); err != nil {
		return QueueActionParams{}, fmt.Errorf("could not verify interactivity request: %v", err)
	}

	payload := r.FormValue("payload")

	return SlurpInteractiveAction(payload), nil
}

func (s *Slack) ParseCommandRequest(r *http.Request) (QueueActionParams, error) {
	verifier, err := slack.NewSecretsVerifier(r.Header, s.SigningSecret)
	if err != nil {
		return QueueActionParams{}, fmt.Errorf("could not create verifier: %v", err)
	}

	r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
	slashCmd, err := slack.SlashCommandParse(r)
	if err != nil {
		return QueueActionParams{}, fmt.Errorf("could not parse command: %v", err)
	}
	if err := verifier.Ensure(); err != nil {
		return QueueActionParams{}, fmt.Errorf("could not verify request: %v", err)
	}

	return SlurpSlashCommand(slashCmd), nil
}

func readUntilColon(s string) string {
	return strings.Split(s, ":")[0]
}
