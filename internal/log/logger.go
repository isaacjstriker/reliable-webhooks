package log

import (
	"log/slog"
	"os"
)

func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}