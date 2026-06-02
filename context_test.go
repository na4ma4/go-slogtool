package slogtool_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/na4ma4/go-slogtool"
)

func TestNewSlogManagerInContext(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	ctx, logmgr, coreLogger, err := slogtool.NewSlogManagerInContext(
		ctx, false,
		slogtool.WithWriter(buf),
	)
	if err != nil {
		t.Fatalf("NewSlogManagerInContext returned error: %v", err)
	}
	if logmgr == nil {
		t.Fatal("NewSlogManagerInContext returned nil LogManager")
	}
	if coreLogger == nil {
		t.Fatal("NewSlogManagerInContext returned nil core logger")
	}

	coreLogger.InfoContext(ctx, "test message")
	if buf.Len() == 0 {
		t.Fatal("expected log output after writing a message")
	}

	retrieved := slogtool.LogManagerFromContext(ctx)
	if retrieved == nil {
		t.Fatal("LogManagerFromContext returned nil")
	}
	if retrieved != logmgr {
		t.Fatal("LogManagerFromContext returned a different LogManager")
	}
}

func TestNewSlogManagerInContextDebug(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, _, _, err := slogtool.NewSlogManagerInContext(ctx, true)
	if err != nil {
		t.Fatalf("NewSlogManagerInContext with debug=true returned error: %v", err)
	}
}

//nolint:paralleltest // these tests modify global slog state, so they cannot be run in parallel
func TestLogManagerFromContextFallback(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	saved := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(saved)

	ctx := context.Background()
	logmgr := slogtool.LogManagerFromContext(ctx)
	if logmgr == nil {
		t.Fatal("LogManagerFromContext returned nil")
	}

	core := logmgr.Named("Core")
	if core == nil {
		t.Fatal("Named returned nil")
	}
}

func TestLogManagerFromContextWithExisting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx, logmgr, _, err := slogtool.NewSlogManagerInContext(ctx, false)
	if err != nil {
		t.Fatalf("NewSlogManagerInContext returned error: %v", err)
	}

	retrieved := slogtool.LogManagerFromContext(ctx)
	if retrieved != logmgr {
		t.Fatal("LogManagerFromContext returned a different LogManager")
	}
}
