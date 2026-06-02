package prettylog_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/na4ma4/go-slogtool/prettylog"
)

func TestColorizeWrapsWithEscapeCodes(t *testing.T) {
	t.Parallel()

	got := prettylog.Colorize(prettylog.Red, "hello")
	want := "\033[31mhello\033[0m"
	if got != want {
		t.Fatalf("colorize mismatch: got=%q want=%q", got, want)
	}
}

func TestSuppressDefaultsRemovesStandardAttrs(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func TestHandlerWithAttrs(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	withAttrs := h.WithAttrs([]slog.Attr{slog.String("key", "val")})
	if withAttrs == h {
		t.Fatal("WithAttrs should return a new Handler")
	}

	if !withAttrs.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("WithAttrs should preserve Enabled")
	}
}

func TestHandlerWithGroup(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	withGroup := h.WithGroup("testgroup")
	if withGroup == h {
		t.Fatal("WithGroup should return a new Handler")
	}

	if !withGroup.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("WithGroup should preserve Enabled")
	}
}

func TestHandlerHandleWithReplaceAttr(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "foo" {
				return slog.String("foo", "replaced")
			}
			return a
		},
	})

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
	if !strings.Contains(line, `"foo":"replaced"`) {
		t.Fatalf("expected output to contain replaced attr, got=%q", line)
	}
}

func TestHandlerHandleWithSuppressedDefaults(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.MessageKey {
				return slog.Attr{}
			}
			return a
		},
	})

	rec := slog.NewRecord(
		time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		slog.LevelInfo,
		"hello",
		0,
	)

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if strings.Contains(line, "INFO:") || strings.Contains(line, "hello") {
		t.Fatalf("expected suppressed output without level or message, got=%q", line)
	}
}

func TestHandlerHandleWithBetweenLevel(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	rec := slog.NewRecord(
		time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		slog.Level(2),
		"between msg",
		0,
	)

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, "between msg") {
		t.Fatalf("expected message in output, got=%q", line)
	}
}

func TestHandlerHandleWithVeryHighLevel(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	rec := slog.NewRecord(
		time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		slog.Level(10),
		"critical msg",
		0,
	)

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, "critical msg") {
		t.Fatalf("expected message in output, got=%q", line)
	}
}

func TestHandlerHandleWithDebugLevel(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	rec := slog.NewRecord(
		time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		slog.LevelDebug,
		"debug msg",
		0,
	)

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, "DEBUG:") {
		t.Fatalf("expected DEBUG level in output, got=%q", line)
	}
}

func TestHandlerHandleWithWarnLevel(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	rec := slog.NewRecord(
		time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		slog.LevelWarn,
		"warn msg",
		0,
	)

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, "WARN:") {
		t.Fatalf("expected WARN level in output, got=%q", line)
	}
}

func TestHandlerHandleWithErrorLevel(t *testing.T) {
	t.Parallel()

	out := bytes.NewBuffer(nil)
	h := prettylog.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})

	rec := slog.NewRecord(
		time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		slog.LevelError,
		"error msg",
		0,
	)

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, "ERROR:") {
		t.Fatalf("expected ERROR level in output, got=%q", line)
	}
}

func TestHandlerEnabled(t *testing.T) {
	t.Parallel()

	h := prettylog.NewHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})

	if h.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("expected DEBUG to be disabled at WARN level")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Fatal("expected WARN to be enabled at WARN level")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("expected ERROR to be enabled at WARN level")
	}
}
