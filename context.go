package slogtool

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/na4ma4/go-slogtool/prettylog"
)

type contextKey string

const (
	contextKeyLogMgr contextKey = "logmanager"
)

// NewSlogManagerInContext creates a new slog manager and a core logger, and stores the slog manager in the provided context.
// It returns the new context, the slog manager, the core logger, and any error that occurred during the creation of the slog manager.
func NewSlogManagerInContext(
	ctx context.Context,
	debug bool,
	opts ...any,
) (context.Context, LogManager, *slog.Logger, error) {
	opts = append([]any{WithSource(false)}, opts...)

	if debug {
		opts = append([]any{
			WithDefaultLevel(slog.LevelDebug),
			WithCustomHandler(func(_ string, _ io.Writer, opts *slog.HandlerOptions) slog.Handler {
				return prettylog.NewHandler(os.Stdout, opts)
			}),
		}, opts...)
	}

	var logmgr LogManager
	{
		var err error
		logmgr, err = NewSlogManager(opts...)
		if err != nil {
			return ctx, nil, nil, err
		}
	}
	coreLogger := logmgr.Named("Core")

	ctx = context.WithValue(ctx, contextKeyLogMgr, logmgr)

	return ctx, logmgr, coreLogger, nil
}

// LogManagerFromContext retrieves the slog manager from the provided context, if it exists. If it does not exist, it creates a new slog manager, logs a warning to the core logger, and returns the new slog manager.
func LogManagerFromContext(ctx context.Context) LogManager {
	if v, ok := ctx.Value(contextKeyLogMgr).(LogManager); ok && v != nil {
		return v
	}

	logmgr, _ := NewSlogManager()
	logmgr.Named("Core").WarnContext(
		ctx,
		"attempted to retrieve log manager from context, but no logmanager is in context, returning a new log manager",
	)
	return logmgr
}
