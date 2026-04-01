package slogtool_test

import (
	"log/slog"
	"testing"

	"github.com/na4ma4/go-slogtool"
)

func TestLevelWrapperSetLevel(t *testing.T) {
	// This test just verifies that the LevelWrapper's SetLevel method can be called without error.
	// The actual level filtering behavior is tested in the handler tests.
	wrapper := slogtool.NewSlogHandlerWrapper(nil, slog.LevelInfo)
	if v := wrapper.Enabled(t.Context(), slog.LevelInfo); !v {
		t.Fatalf("expected LevelInfo to be true, got=%t", v)
	}
	if v := wrapper.Enabled(t.Context(), slog.LevelDebug); v {
		t.Fatalf("expected LevelDebug to be false, got=%t", v)
	}

	wrapper.SetLevel(slog.LevelDebug)
	if v := wrapper.Enabled(t.Context(), slog.LevelInfo); !v {
		t.Fatalf("expected LevelInfo to be true, got=%t", v)
	}
	if v := wrapper.Enabled(t.Context(), slog.LevelDebug); !v {
		t.Fatalf("expected LevelDebug to be true, got=%t", v)
	}
}
