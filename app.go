package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/khepin/liteq"
	"github.com/khepin/queuebee/queries"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type App struct {
	Slack     *Slack
	QueueRepo *queries.Queries
	JobQueue  *liteq.JobQueue
	Logger    *slog.Logger
}

type Slack struct {
	Secret        string
	Client        *slack.Client
	SocketClient  *socketmode.Client
	SigningSecret string
}

func NewApp() *App {
	slackSecret := env("SLACK_SECRET", "")
	slackSigningSecret := env("SLACK_SIGNING_SECRET", "")

	slackClient := slack.New(
		slackSecret,
		slack.OptionAppLevelToken(env("SLACK_APP_TOKEN", "")),
	)
	liteqDb, err := sql.Open("sqlite3", env("LITEQ_DB", "liteq.db"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	liteq.Setup(liteqDb)
	jqueue := liteq.New(liteqDb)

	var socketClient *socketmode.Client
	if env("SOCKETMODE", "false") == "true" {
		socketClient = socketmode.New(slackClient)
	}

	app := &App{
		Slack: &Slack{
			Secret:        slackSecret,
			Client:        slackClient,
			SocketClient:  socketClient,
			SigningSecret: slackSigningSecret,
		},
		JobQueue: jqueue,
	}

	app.Logger = NewSlogger()

	sqlcdb, err := sql.Open("sqlite3", env("QUEUEBEE_DB", "queuebee.db"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app.QueueRepo = queries.New(sqlcdb)

	return app
}
