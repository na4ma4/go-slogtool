package slogtool

import (
	"context"
	"io"
	"log/slog"
)

func TestLogger(w io.Writer) (*slog.LevelVar, *slog.Logger) {
	if w == nil {
		return new(slog.LevelVar), slog.New(SlogNullHandler{})
	}

	lvl := new(slog.LevelVar)
	return lvl, slog.New(
		slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: lvl,
		}),
	)
}

type SlogNullHandler struct{}

// Enabled reports whether the handler handles records at the given level.
func (SlogNullHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

// Handle handles the Record.
func (SlogNullHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (SlogNullHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return SlogNullHandler{}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (SlogNullHandler) WithGroup(_ string) slog.Handler {
	return SlogNullHandler{}
}
