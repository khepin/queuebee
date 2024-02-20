package main

import (
	"strings"

	"github.com/slack-go/slack"
)

func help(app *App, params QueueActionParams) error {
	res, err := app.Slack.Client.OpenView(params.TriggerID, slack.ModalViewRequest{
		Type:  slack.VTModal,
		Title: slackText("QueueBee help"),
		Close: slackText("Close"),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slackMarkdownSection(strings.ReplaceAll(`QueueBee is a simple queue management system for Slack.
It allows you to create a queue and have people join it. Once in the queue, people can take their turn and leave the queue.
The queue can be configured to automatically skip people who don't take their turn in time to ensure the queue keeps moving.

*Available commands:*
- ''/qbee join'': Join the queue
- ''/qbee leave'': Leave the queue
- ''/qbee skip'': Skip your turn
- ''/qbee list'': List the current queue
- ''/qbee ack'': Acknowledge your turn
- ''/qbee config'': Configure the queue
- ''/qbee help'': Show this help message`, "''", "`")),
			},
		},
	})

	if err != nil {
		app.Logger.Error("Could not open config modal", "error", err, "res", res)
	}

	return err
}
