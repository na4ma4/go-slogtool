package prettylog_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/na4ma4/go-slogtool/prettylog"
)

func TestColorizeWrapsWithEscapeCodes(t *testing.T) {
	got := prettylog.Colorize(prettylog.Red, "hello")
	want := "\033[31mhello\033[0m"
	if got != want {
		t.Fatalf("colorize mismatch: got=%q want=%q", got, want)
	}
}

func TestSuppressDefaultsRemovesStandardAttrs(t *testing.T) {
	replace := prettylog.SuppressDefaults(nil)

	if got := replace(nil, slog.Attr{Key: slog.TimeKey, Value: slog.StringValue("x")}); !got.Equal(slog.Attr{}) {
		t.Fatalf("expected time attr to be suppressed, got=%v", got)
	}
	if got := replace(nil, slog.Attr{Key: slog.LevelKey, Value: slog.StringValue("x")}); !got.Equal(slog.Attr{}) {
		t.Fatalf("expected level attr to be suppressed, got=%v", got)
	}
	if got := replace(nil, slog.Attr{Key: slog.MessageKey, Value: slog.StringValue("x")}); !got.Equal(slog.Attr{}) {
		t.Fatalf("expected message attr to be suppressed, got=%v", got)
	}

	plain := slog.String("foo", "bar")
	if got := replace(nil, plain); !got.Equal(plain) {
		t.Fatalf("expected non-default attr to pass through, got=%v want=%v", got, plain)
	}
}

func TestHandlerHandleWritesExpectedFields(t *testing.T) {
	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	rec := slog.NewRecord(
		time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		slog.LevelInfo,
		"hello",
		0,
	)
	rec.AddAttrs(slog.String("foo", "bar"))

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	checks := []string{
		"[03:04:05.000]",
		"INFO:",
		"hello",
		`"foo":"bar"`,
		"\033[",
	}

	for _, needle := range checks {
		if !strings.Contains(line, needle) {
			t.Fatalf("expected output to contain %q, got=%q", needle, line)
		}
	}
}
