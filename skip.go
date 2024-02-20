package main

import (
	"context"
	"fmt"
	"time"
)

func skipTurn(app *App, params QueueActionParams) error {
	queue, err := app.QueueRepo.FindQueue(context.Background(), params.QueueID())
	if err != nil {
		return fmt.Errorf("could not find queue: %w", err)
	}

	if queue.GetFirstUser() == params.UserID && len(queue.GetUsersInLine()) == 0 {
		return leaveQueue(app, params)
	}

	sendAckRequiredMessage := false
	if queue.GetFirstUser() == params.UserID {
		sendAckRequiredMessage = true
		queue.AckedBy = ""
		queue.RequireAckBefore = time.Now().Add(time.Duration(queue.Config.InitialAckTimeout) * time.Minute).Unix()
	}

	skipped, err := queue.SkipTurn(params.UserID)
	if err != nil {
		return fmt.Errorf("could not remove user from queue: %w", err)
	}
	if !skipped {
		return nil
	}
	err = app.QueueRepo.SaveQueue(context.Background(), queue.ToSaveParams())
	if err != nil {
		return fmt.Errorf("could not save queue: %w", err)
	}

	sendQueueMessage(app.Slack.Client, queue)
	if sendAckRequiredMessage {
		sendAckTimeoutMessage(app.Slack.Client, queue)
		planAckReminder(app, queue)
	}

	return nil
}
