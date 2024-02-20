package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/khepin/queuebee/queries"
)

func joinQueue(app *App, params QueueActionParams) error {
	err := app.QueueRepo.UpsertQueue(context.Background(), queries.UpsertQueueParams{
		QueueID:     params.QueueID(),
		ChannelID:   params.ChannelID,
		TeamID:      params.TeamID,
		ChannelName: params.ChannelName,
		TeamDomain:  params.TeamDomain,
		Config: queries.Config{
			InitialAckTimeout:     initialAckTimeout,
			SubsequentAckTimeouts: []int64{subsequentAckTimeout},
		},
	})
	if err != nil {
		return fmt.Errorf("could not create queue: %w", err)
	}

	queue, err := app.QueueRepo.FindQueue(context.Background(), params.QueueID())
	if err != nil {
		return fmt.Errorf("could not find queue: %w", err)
	}
	if queue.GetFirstUser() == "" {
		queue.RequireAckBefore = time.Now().Add(time.Duration(queue.Config.InitialAckTimeout) * time.Minute).Unix()
	}
	err = queue.AddUser(params.UserID)
	if err != nil && errors.As(err, new(queries.UserAlreadyInQueueError)) {
		return nil
	}
	err = app.QueueRepo.SaveQueue(context.Background(), queue.ToSaveParams())
	if err != nil {
		return fmt.Errorf("could not save queue: %w", err)
	}

	sendQueueMessage(app.Slack.Client, queue)
	if queue.GetFirstUser() == params.UserID {
		sendAckTimeoutMessage(app.Slack.Client, queue)
		planAckReminder(app, queue)
	}

	return nil
}
