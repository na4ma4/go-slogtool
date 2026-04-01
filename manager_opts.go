package slogtool

import (
	"fmt"
	"log/slog"
)

func WithSource(withSource bool) SlogManagerHandlerOpts {
	return func(ho *slog.HandlerOptions) error {
		ho.AddSource = withSource
		return nil
	}
}

func WithLevel(lvl any) SlogManagerHandlerOpts {
	return func(ho *slog.HandlerOptions) error {
		v, ok := slogParseLevel(lvl)
		if !ok {
			return fmt.Errorf("invalid log level: %v", lvl)
		}
		ho.Level = v
		return nil
	}
}
