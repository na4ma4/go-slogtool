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

func NewSlogManagerInContext(
	ctx context.Context,
	debug bool,
	opts ...interface{},
) (context.Context, LogManager, *slog.Logger) {
	opts = append([]interface{}{WithSource(false)}, opts...)

	if debug {
		opts = append(opts,
			WithDefaultLevel(slog.LevelDebug),
			WithCustomHandler(func(_ string, _ io.Writer, opts *slog.HandlerOptions) slog.Handler {
				return prettylog.NewHandler(os.Stdout, opts)
			}),
		)
	}

	var logmgr LogManager = NewSlogManager(opts...)
	coreLogger := logmgr.Named("Core")

	ctx = context.WithValue(ctx, contextKeyLogMgr, logmgr)

	return ctx, logmgr, coreLogger
}

func LogManagerFromContext(ctx context.Context) LogManager {
	if v, ok := ctx.Value(contextKeyLogMgr).(LogManager); ok && v != nil {
		return v
	}

	logmgr := NewSlogManager()
	logmgr.Named("Core").WarnContext(
		ctx,
		"attempted to retrieve log manager from context, but no logmanager is in context, returning a new log manager",
	)
	return logmgr
}
