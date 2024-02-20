package main

import (
	"context"
	"os"
	"time"

	"github.com/khepin/liteq"
	"github.com/labstack/echo/v4"
	slogecho "github.com/samber/slog-echo"
)

const initialAckTimeout = 1
const subsequentAckTimeout = 10

func main() {
	e := echo.New()

	app := NewApp()
	e.Use(slogecho.New(app.Logger))

	consume(app)

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
