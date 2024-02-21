package main

import (
	"context"
	"os"
	"time"

	"github.com/khepin/liteq"
	"github.com/labstack/echo/v4"
	slogecho "github.com/samber/slog-echo"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const initialAckTimeout = 1
const subsequentAckTimeout = 10

func main() {
	app := NewApp()
	consume(app)

	if env("SOCKETMODE", "false") == "false" {
		httpApp(app)
	} else {
		socketApp(app)
	}
}

func socketApp(app *App) {
	app.Logger.Info("starting queuebee in socket mode")
	sock := app.Slack.SocketClient
	go sock.Run()
	for event := range sock.Events {
		go func(event socketmode.Event) {
			var err error
			start := time.Now()
			switch event.Type {
			case socketmode.EventTypeSlashCommand:
				d := event.Data.(slack.SlashCommand)
				err = app.HandleAction(SlurpSlashCommand(d))
				sock.Ack(*event.Request)
			case socketmode.EventTypeInteractive:
				err = app.HandleAction(SlurpInteractiveAction(string(event.Request.Payload)))
				sock.Ack(*event.Request)
			}

			if err != nil {
				app.Logger.Error("socket event", "type", event.Type, "error", err, "latency_ms", time.Since(start).Milliseconds())
			} else {
				app.Logger.Info("socket event", "type", event.Type, "latency_ms", time.Since(start).Milliseconds())
			}
		}(event)
	}
}

func httpApp(app *App) {
	e := echo.New()

	e.Use(slogecho.New(app.Logger))

	e.POST("/slack/interactivity", app.HandleInteractivity)
	e.POST("/slack/commands/queuebee", app.HandleSlashCommand)

	port := env("PORT", "8002")
	e.Logger.Fatal(e.Start("localhost:" + port))
}

func env(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func retry(f func() error) {
	err := f()
	if err != nil {
		time.Sleep(1 * time.Second)
		err = f()
		if err != nil {
			time.Sleep(3 * time.Second)
			f()
		}
	}
}

func consume(app *App) {
	go func() {
		for {
			err := app.JobQueue.Consume(context.Background(), liteq.ConsumeParams{
				Queue:             "ack",
				PoolSize:          3,
				VisibilityTimeout: 20,
				Worker:            ackWorker(app),
			})
			if err != nil {
				app.Logger.Error("error consuming ack queue", err)
				time.Sleep(1 * time.Second)
			}
		}
	}()
}
