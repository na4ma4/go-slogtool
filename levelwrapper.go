package slogtool

import (
	"context"
	"log/slog"
	"sync"
)

type SlogHandlerWrapper struct {
	lock    sync.RWMutex
	handler slog.Handler
	level   *slog.LevelVar
}

func NewSlogHandlerWrapper(handler slog.Handler, lvl slog.Leveler) *SlogHandlerWrapper {
	level := new(slog.LevelVar)
	level.Set(lvl.Level())
	return &SlogHandlerWrapper{
		handler: handler,
		level:   level,
	}
}

// Level return a reference to *slog.LevelVar so the level can be changed on request.
func (h *SlogHandlerWrapper) SetLevel(lvl slog.Level) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.level.Set(lvl)
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// It is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
// If called from a Logger method, the first argument is the context
// passed to that method, or context.Background() if nil was passed
// or the method does not take a context.
// The context is passed so Enabled can use its values
// to make a decision.
func (h *SlogHandlerWrapper) Enabled(_ context.Context, in slog.Level) bool {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return in > h.level.Level()
}

// Handle handles the Record.
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// It is present solely to provide Handlers access to the context's values.
// Canceling the context should not affect record processing.
// (Among other things, log messages may be necessary to debug a
// cancellation-related problem.)
//
// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - If r.PC is zero, ignore it.
//   - Attr's values should be resolved.
//   - If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - If a group's key is empty, inline the group's Attrs.
//   - If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
func (h *SlogHandlerWrapper) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(
		ctx,
		r,
	)
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (h *SlogHandlerWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewSlogHandlerWrapper(
		h.handler.WithAttrs(attrs),
		h.level.Level(),
	)
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, should be qualified by the sequence of group names.
//
// How this qualification happens is up to the Handler, so long as
// this Handler's attribute keys differ from those of another Handler
// with a different sequence of group names.
//
// A Handler should treat WithGroup as starting a Group of Attrs that ends
// at the end of the log event. That is,
//
//	logger.WithGroup("s").LogAttrs(ctx, level, msg, slog.Int("a", 1), slog.Int("b", 2))
//
// should behave like
//
//	logger.LogAttrs(ctx, level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
//
// If the name is empty, WithGroup returns the receiver.
func (h *SlogHandlerWrapper) WithGroup(name string) slog.Handler {
	return NewSlogHandlerWrapper(
		h.handler.WithGroup(name),
		h.level.Level(),
	)
}
