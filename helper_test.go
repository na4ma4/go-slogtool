package slogtool_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/na4ma4/go-slogtool"
)

func TestErrorLevel(t *testing.T) {
	if lvl := slogtool.ErrorLevel(nil); lvl != slog.LevelInfo {
		t.Fatalf("expected info level for nil error, got=%v", lvl)
	}
	err := errors.New("test error")
	if lvl := slogtool.ErrorLevel(err); lvl != slog.LevelError {
		t.Fatalf("expected error level for non-nil error, got=%v", lvl)
	}
}

func TestErrorAttr(t *testing.T) {
	if attr := slogtool.ErrorAttr(nil); !attr.Equal(slog.Attr{}) {
		t.Fatalf("expected empty attr for nil error, got=%v", attr)
	}
	err := errors.New("test error")
	attr := slogtool.ErrorAttr(err)
	if attr.Key != "error" {
		t.Fatalf("expected key 'error', got=%q", attr.Key)
	}
	if val, ok := attr.Value.Any().(error); !ok || val.Error() != "test error" {
		t.Fatalf("expected error value 'test error', got=%v", attr.Value.Any())
	}
}

func TestStack(t *testing.T) {
	attr := slogtool.Stack("stacktrace")
	if attr.Key != "stacktrace" {
		t.Fatalf("expected key 'stacktrace', got=%q", attr.Key)
	}
	if val, ok := attr.Value.Any().(string); !ok || val == "" {
		t.Fatalf("expected non-empty string value for stack trace, got=%v", attr.Value.Any())
	}
}
