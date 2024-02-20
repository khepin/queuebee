package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

type ColorHandler struct {
	slog.Handler
}

func (h *ColorHandler) Handle(ctx context.Context, record slog.Record) error {
	defer color.Unset()
	switch record.Level {
	case slog.LevelDebug:
		color.Set(color.FgBlue)
	case slog.LevelInfo:
		color.Set(color.FgGreen)
	case slog.LevelWarn:
		color.Set(color.FgYellow)
	case slog.LevelError:
		color.Set(color.FgRed)
	}
	return h.Handler.Handle(ctx, record)
}

func NewSlogger() *slog.Logger {
	var handler slog.Handler
	if isatty.IsTerminal(os.Stdout.Fd()) {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})
		handler = &ColorHandler{handler}
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})
	}

	return slog.New(handler)
}
