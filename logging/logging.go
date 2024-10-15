package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/namlh/vulcanLabsOA/config"
	"github.com/namlh/vulcanLabsOA/consts/ctxkey"
)

type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID, ok := ctx.Value(ctxkey.RequestID{}).(int64); ok {
		r.AddAttrs(slog.Int64("request_id", requestID))
	}

	return h.Handler.Handle(ctx, r)
}

func NewLogger(cfg *config.Logger, env string) (*slog.Logger, error) {
	opts := slog.HandlerOptions{
		Level: cfg.Level,
	}

	var out io.Writer = os.Stdout
	if cfg.Filepath != "" {
		var err error
		out, err = os.OpenFile(cfg.Filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("open file %s: %w", cfg.Filepath, err)
		}
	}

	var handler slog.Handler = slog.NewJSONHandler(out, &opts)
	if env != "prod" {
		opts.ReplaceAttr = func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				a.Value = slog.StringValue(time.Now().Format(time.DateTime))
			}

			return a
		}
		handler = slog.NewTextHandler(out, &opts)
	}
	handler = contextHandler{handler}

	logger := slog.New(handler)

	return logger, nil
}
