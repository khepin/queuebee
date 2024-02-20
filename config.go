package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
	"github.com/tidwall/gjson"
)

func configQueue(app *App, params QueueActionParams) error {
	q, err := app.QueueRepo.FindQueue(context.Background(), params.QueueID())
	if err != nil {
		return fmt.Errorf("could not find queue: %w", err)
	}
	res, err := app.Slack.Client.OpenView(params.TriggerID, slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           slackText("QueueBee config"),
		Close:           slackText("Close"),
		Submit:          slackText("Save"),
		CallbackID:      "save_config",
		PrivateMetadata: fmt.Sprintf(`{"queue_id": "%s", "channel_id": "%s", "channel_name": "%s"}`, params.QueueID(), params.ChannelID, params.ChannelName),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slackInputWithInitialValue(
					"initial_ack_timeout",
					"Initial Ack Timeout",
					"When you become the first in the queue, this is how long you have to acknowledge that you are indeed taking your turn. After that time, we'll move on to the next person in the queue.",
					fmt.Sprintf("%d", q.Config.InitialAckTimeout),
				),
				slackDivider(),
				slackInputWithInitialValue(
					"subsequent_ack_timeout",
					"Subsequent Ack Timeout",
					"Once you are in the queue, this is how many minutes you have as the first in the queue. After that time, you will be pinged to acknowledge that you are still there. If you don't, we'll move on to the next person in the queue. \nMultiple Values: you can provide a list of comma separated values eg:  `10,20,30,40,50,60`",
					q.Config.SubsequentAckTimeoutsStr(),
				),
			},
		},
	})

	if err != nil {
		app.Logger.Error("Could not open config modal", "error", err, "res", res)
	}

	return err
}

func saveConfig(app *App, params QueueActionParams) error {
	initialAckTimeoutStr := gjson.Get(params.ActionValue, "initial_ack_timeout.initial_ack_timeout.value").String()
	initialAckTimeout, err := strconv.Atoi(initialAckTimeoutStr)
	if err != nil {
		return fmt.Errorf("initial ack timeout must be a number: %w", err)
	}
	subsequentAckTimeoutStr := gjson.Get(params.ActionValue, "subsequent_ack_timeout.subsequent_ack_timeout.value").String()
	timeoutStrs := strings.Split(subsequentAckTimeoutStr, ",")
	timeouts := []int64{}
	for _, timeoutStr := range timeoutStrs {
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return fmt.Errorf("subsequent ack timeout must be a number: %w", err)
		}
		timeouts = append(timeouts, int64(timeout))
	}

	q, err := app.QueueRepo.FindQueue(context.Background(), params.QueueID())
	if err != nil {
		return fmt.Errorf("could not find queue: %w", err)
	}
	q.Config.InitialAckTimeout = int64(initialAckTimeout)
	q.Config.SubsequentAckTimeouts = timeouts

	err = app.QueueRepo.SaveQueue(context.Background(), q.ToSaveParams())
	if err != nil {
		return fmt.Errorf("could not save queue: %w", err)
	}

	return nil
}

func slackInputWithInitialValue(actionID, label, hint, initialValue string) slack.InputBlock {
	input := slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", label, false, false), actionID)
	input.InitialValue = initialValue
	return *slack.NewInputBlock(
		actionID,
		slack.NewTextBlockObject("plain_text", label, false, false),
		slack.NewTextBlockObject("plain_text", hint, false, false),
		input,
	)
}

func slackText(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject("plain_text", text, false, false)
}

func slackDivider() slack.DividerBlock {
	return slack.DividerBlock{
		Type: "divider",
	}
}

func slackMarkdown(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject("mrkdwn", text, false, false)
}

func slackMarkdownSection(text string) slack.SectionBlock {
	return slack.SectionBlock{
		Type: "section",
		Text: slackMarkdown(text),
	}
}
