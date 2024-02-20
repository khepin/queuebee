package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/khepin/liteq"
	"github.com/khepin/queuebee/queries"
	"github.com/slack-go/slack"
)

func ackTurn(app *App, params QueueActionParams) error {
	queue, err := app.QueueRepo.FindQueue(context.Background(), params.QueueID())
	if err != nil {
		return fmt.Errorf("could not find queue: %w", err)
	}
	if queue.GetFirstUser() != params.UserID {
		return fmt.Errorf("user not up next (%s)", params.UserID)
	}

	timeout, err := strconv.Atoi(params.ActionValue)
	if err != nil {
		return fmt.Errorf("could not parse timeout: %w", err)
	}

	queue.RequireAckBefore = time.Now().Add(time.Duration(timeout)*time.Minute + 20*time.Second).Unix()
	queue.AckedBy = params.UserID
	err = app.QueueRepo.SaveQueue(context.Background(), queue.ToSaveParams())
	if err != nil {
		return fmt.Errorf("could not save queue: %w", err)
	}
	sendAcknowledgedMessage(app.Slack.Client, queue)
	planAckReminder(app, queue)
	return nil
}

func sendAcknowledgedMessage(slackClient *slack.Client, queue *queries.Queue) {
	message := fmt.Sprintf(
		"🐝ACK🐝 you have the queue until <!date^%d^{time}|time parse error> about %dm\nI will ping you to extend before the end of that time.",
		queue.RequireAckBefore,
		time.Until(time.Unix(queue.RequireAckBefore, 0))/time.Minute,
	)
	go retry(func() error {
		_, err := slackClient.PostEphemeral(queue.ChannelID, queue.GetFirstUser(), slack.MsgOptionText(message, false))
		return err

	})
}

func planAckReminder(app *App, queue *queries.Queue) {
	timeout := time.Unix(queue.RequireAckBefore, 0)
	// By default we want to expire 5 seconds after the official timeout
	// That way we avoid shceduling another such message
	// And we also give the user more than the official timeout to ack
	at := timeout.Add(5 * time.Second)

	if time.Until(timeout) > 2*time.Minute {
		at = timeout.Add(-2*time.Minute - 5*time.Second)
	}

	if time.Until(timeout) > (time.Minute + 6*time.Second) {
		at = timeout.Add(-time.Minute - 5*time.Second)
	}

	app.JobQueue.QueueJob(context.Background(), liteq.QueueJobParams{
		Queue:             "ack",
		ExecuteAfter:      at.Unix(),
		RemainingAttempts: 3,
		Job:               serializeJob(AckReminderJob{QueueID: queue.QueueID, UserID: queue.GetFirstUser()}),
		DedupingKey:       liteq.ReplaceDuplicate(fmt.Sprintf("ack:%s:%s", queue.QueueID, queue.GetFirstUser())),
	})
}

type AckReminderJob struct {
	QueueID string
	UserID  string
}

func serializeJob[T any](job T) string {
	b, err := json.Marshal(job)
	if err != nil {
		return ""
	}
	return string(b)
}

func deserializeJob[T any](job string) (T, error) {
	var j T
	err := json.Unmarshal([]byte(job), &j)
	return j, err
}

func ackWorker(app *App) func(ctx context.Context, job *liteq.Job) error {
	return func(ctx context.Context, job *liteq.Job) error {
		ackJob, err := deserializeJob[AckReminderJob](job.Job)
		if err != nil {
			app.Logger.Error("could not deserialize ack job", "error", err)
			return err
		}
		queue, err := app.QueueRepo.FindQueue(ctx, ackJob.QueueID)
		if err != nil {
			app.Logger.Error("could not find queue for ack job", "error", err)
			return err
		}
		if queue.GetFirstUser() != ackJob.UserID {
			return nil
		}

		if queue.RequireAckBefore < time.Now().Unix() {
			// If the user has not acked their initial turn and there are others in the queue, we skip them
			if queue.AckedBy == "" && len(queue.GetUsersInLine()) > 0 {
				err = skipTurn(app, QueueActionParams{
					Action:    "skip",
					UserID:    queue.GetFirstUser(),
					ChannelID: queue.ChannelID,
					TeamID:    queue.TeamID,
				})
				if err != nil {
					app.Logger.Error("could not skip turn", "error", err)
				}
				return err
			}

			err = leaveQueue(app, QueueActionParams{
				Action:    "leave",
				UserID:    queue.GetFirstUser(),
				ChannelID: queue.ChannelID,
				TeamID:    queue.TeamID,
			})
			if err != nil {
				app.Logger.Error("could not leave queue", "error", err)
			}
			return err
		}

		if queue.RequireAckBefore > time.Now().Unix() && queue.RequireAckBefore <= time.Now().Add(2*time.Minute).Unix() {
			sendAckTimeoutMessage(app.Slack.Client, queue)
			planAckReminder(app, queue)
			return nil
		}

		planAckReminder(app, queue)

		return nil
	}
}

func sendAckTimeoutMessage(slackClient *slack.Client, queue *queries.Queue) {
	if queue.GetFirstUser() == "" {
		return
	}

	buttons := []slack.BlockElement{}

	for i, timeout := range queue.Config.SubsequentAckTimeouts {
		ackButton := slack.NewButtonBlockElement(fmt.Sprintf("ack:%d", i), fmt.Sprintf("%d", timeout), slack.NewTextBlockObject("plain_text", fmt.Sprintf("Ack (%dm)", timeout), false, false))
		ackButton.Style = "primary"

		buttons = append(buttons, ackButton)
	}

	if len(queue.Config.SubsequentAckTimeouts) == 0 {
		ackButton := slack.NewButtonBlockElement("ack:0", fmt.Sprintf("%d", subsequentAckTimeout), slack.NewTextBlockObject("plain_text", "Ack", false, false))
		ackButton.Style = "primary"
		buttons = append(buttons, ackButton)
	}

	leaveButton := slack.NewButtonBlockElement("leave", "leave", slack.NewTextBlockObject("plain_text", "Leave", false, false))
	leaveButton.Style = "danger"
	skipButton := slack.NewButtonBlockElement("skip", "skip", slack.NewTextBlockObject("plain_text", "Skip My Turn", false, false))

	buttons = append(buttons, skipButton, leaveButton)
	blockset := []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s> You have until <!date^%d^{time}|time parse error> to acknowledge this message before we move on to the next in line", queue.GetFirstUser(), queue.RequireAckBefore), false, false), nil, nil),
		slack.NewDividerBlock(),
		slack.NewActionBlock(
			"qbee-actions",
			buttons...,
		),
	}

	go retry(func() error {
		time.Sleep(1 * time.Second)
		_, err := slackClient.PostEphemeral(queue.ChannelID, queue.GetFirstUser(), slack.MsgOptionBlocks(blockset...))
		return err
	})
}
