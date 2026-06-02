package slogtool_test

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/na4ma4/go-slogtool"
)

func TestLevelWrapperSetLevel(t *testing.T) {
	t.Parallel()

	// This test just verifies that the LevelWrapper's SetLevel method can be called without error.
	// The actual level filtering behavior is tested in the handler tests.
	wrapper := slogtool.NewSlogHandlerWrapper(slog.DiscardHandler, slog.LevelInfo)
	if v := wrapper.Enabled(t.Context(), slog.LevelInfo); !v {
		t.Fatalf("a: expected LevelInfo to be true, got=%t", v)
	}
	if v := wrapper.Enabled(t.Context(), slog.LevelDebug); v {
		t.Fatalf("a: expected LevelDebug to be false, got=%t", v)
	}

	wrapper.SetLevel(slog.LevelDebug)
	if v := wrapper.Enabled(t.Context(), slog.LevelInfo); !v {
		t.Fatalf("b: expected LevelInfo to be true, got=%t", v)
	}
	if v := wrapper.Enabled(t.Context(), slog.LevelDebug); !v {
		t.Fatalf("b: expected LevelDebug to be true, got=%t", v)
	}
}

func TestLevelWrapperHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	wrapper := slogtool.NewSlogHandlerWrapper(inner, slog.LevelInfo)

	rec := slog.NewRecord(time.Time{}, slog.LevelInfo, "test message", 0)
	if err := wrapper.Handle(t.Context(), rec); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected Handle to delegate to inner handler")
	}
}

func TestLevelWrapperWithAttrs(t *testing.T) {
	t.Parallel()

	inner := slog.DiscardHandler
	wrapper := slogtool.NewSlogHandlerWrapper(inner, slog.LevelInfo)

	withAttrs := wrapper.WithAttrs([]slog.Attr{slog.String("key", "val")})
	if withAttrs == wrapper {
		t.Fatal("WithAttrs should return a new wrapper")
	}

	if !withAttrs.Enabled(t.Context(), slog.LevelInfo) {
		t.Fatal("WithAttrs wrapper should preserve level")
	}
}

func TestLevelWrapperWithGroup(t *testing.T) {
	t.Parallel()

	inner := slog.DiscardHandler
	wrapper := slogtool.NewSlogHandlerWrapper(inner, slog.LevelInfo)

	withGroup := wrapper.WithGroup("testgroup")
	if withGroup == wrapper {
		t.Fatal("WithGroup should return a new wrapper")
	}

	if !withGroup.Enabled(t.Context(), slog.LevelInfo) {
		t.Fatal("WithGroup wrapper should preserve level")
	}
}
