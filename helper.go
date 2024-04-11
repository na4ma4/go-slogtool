package slogtool

import "log/slog"

func ErrorLevel(err error) slog.Level {
	if err != nil {
		return slog.LevelError
	}

	return slog.LevelInfo
}
