package slogtool

import (
	"log/slog"
	"runtime/debug"
)

func ErrorLevel(err error) slog.Level {
	if err != nil {
		return slog.LevelError
	}

	return slog.LevelInfo
}

func ErrorAttr(err error) slog.Attr {
	if err != nil {
		return slog.Attr{
			Key:   "error",
			Value: slog.AnyValue(err),
		}
	}

	return slog.Attr{}
}

func Stack(name string) slog.Attr {
	return slog.Any(name, debug.Stack())
}
