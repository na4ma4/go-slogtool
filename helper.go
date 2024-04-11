package slogtool

import "log/slog"

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
