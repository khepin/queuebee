package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/khepin/queuebee/queries"
	"github.com/slack-go/slack"
)

func listQueue(app *App, params QueueActionParams) error {
	queue, err := app.QueueRepo.FindQueue(context.Background(), params.QueueID())
	if err != nil {
		return fmt.Errorf("could not find queue: %w", err)
	}
	sendQueueMessage(app.Slack.Client, queue)
	return nil
}

func sendQueueMessage(slackClient *slack.Client, queue *queries.Queue) {
	leaveButton := slack.NewButtonBlockElement("leave", "leave", slack.NewTextBlockObject("plain_text", "Leave", false, false))
	leaveButton.Style = "danger"
	blockset := []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", currentQueue(queue), false, false), nil, nil),
		slack.NewDividerBlock(),
		slack.NewActionBlock(
			"qbee-actions",
			slack.NewButtonBlockElement("join", "join", slack.NewTextBlockObject("plain_text", "Join", false, false)),
			slack.NewButtonBlockElement("skip", "skip", slack.NewTextBlockObject("plain_text", "Skip My Turn", false, false)),
			leaveButton,
		),
	}

	go retry(func() error {
		_, _, err := slackClient.PostMessage(queue.ChannelID, slack.MsgOptionBlocks(blockset...))
		return err
	})
}

var currentQueueTpl *template.Template

func rankInQueue(i int) string {
	switch i {
	case 0:
		return "2️⃣"
	case 1:
		return "3️⃣"
	case 2:
		return "4️⃣"
	case 3:
		return "5️⃣"
	case 4:
		return "6️⃣"
	case 5:
		return "7️⃣"
	case 6:
		return "8️⃣"
	case 7:
		return "9️⃣"
	default:
		return "- "
	}
}

func init() {
	currentQueueTpl = template.Must(template.New("current-queue").Funcs(template.FuncMap{
		"rank": rankInQueue,
	}).Parse(strings.TrimSpace(`
{{if ne .UID ""}}🐝 Current queue list 🐝
-{{else}}🐝 The queue is empty 🐝{{end}}
{{if ne .UID ""}}1️⃣  <@{{.UID}}>{{end}}
{{range $i, $user := .Queue}}
{{rank $i}}  <@{{.}}>{{end}}`)))
}

func currentQueue(q *queries.Queue) string {
	b := &bytes.Buffer{}
	currentQueueTpl.Execute(b, map[string]interface{}{
		"UID":   q.GetFirstUser(),
		"Queue": q.GetUsersInLine(),
	})
	return b.String()
}
