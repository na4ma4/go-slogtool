package prettylog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

const (
	timeFormat = "[15:04:05.000]"
)

type Handler struct {
	w io.Writer
	h slog.Handler
	r func([]string, slog.Attr) slog.Attr
	b *bytes.Buffer
	m *sync.Mutex
}

func NewHandler(w io.Writer, opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	b := &bytes.Buffer{}
	return &Handler{
		w: w,
		b: b,
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		r: opts.ReplaceAttr,
		m: &sync.Mutex{},
	}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		w: h.w,
		h: h.h.WithAttrs(attrs),
		b: h.b,
		r: h.r,
		m: h.m,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		w: h.w,
		h: h.h.WithGroup(name),
		b: h.b,
		r: h.r,
		m: h.m,
	}
}

func (h *Handler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()
	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func (h *Handler) colourizeLevel(level string, r *slog.Record) string {
	switch {
	case r.Level <= slog.LevelDebug:
		return colorize(lightGray, level)
	case r.Level <= slog.LevelInfo:
		return colorize(cyan, level)
	case r.Level < slog.LevelWarn:
		return colorize(lightBlue, level)
	case r.Level < slog.LevelError:
		return colorize(lightYellow, level)
	case r.Level <= slog.LevelError+1:
		return colorize(lightRed, level)
	case r.Level > slog.LevelError+1:
		return colorize(lightMagenta, level)
	}

	return level
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	var level string
	levelAttr := slog.Attr{
		Key:   slog.LevelKey,
		Value: slog.AnyValue(r.Level),
	}
	if h.r != nil {
		levelAttr = h.r([]string{}, levelAttr)
	}

	if !levelAttr.Equal(slog.Attr{}) {
		level = levelAttr.Value.String() + ":"

		level = h.colourizeLevel(level, &r)
	}

	var timestamp string
	timeAttr := slog.Attr{
		Key:   slog.TimeKey,
		Value: slog.StringValue(r.Time.Format(timeFormat)),
	}
	if h.r != nil {
		timeAttr = h.r([]string{}, timeAttr)
	}
	if !timeAttr.Equal(slog.Attr{}) {
		timestamp = colorize(lightGray, timeAttr.Value.String())
	}

	var msg string
	msgAttr := slog.Attr{
		Key:   slog.MessageKey,
		Value: slog.StringValue(r.Message),
	}
	if h.r != nil {
		msgAttr = h.r([]string{}, msgAttr)
	}
	if !msgAttr.Equal(slog.Attr{}) {
		msg = colorize(white, msgAttr.Value.String())
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(attrs)
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	// bytes, err := json.MarshalIndent(attrs, "", "  ")
	// if err != nil {
	// 	return fmt.Errorf("error when marshaling attrs: %w", err)
	// }

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}
	if len(level) > 0 {
		out.WriteString(level)
		out.WriteString(" ")
	}
	if len(msg) > 0 {
		out.WriteString(msg)
		out.WriteString(" ")
	}
	if len(bytes) > 0 {
		out.WriteString(colorize(darkGray, string(bytes)))
	}
	fmt.Fprintln(h.w, out.String())

	return nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}
