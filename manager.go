package slogtool

import (
	"log/slog"
)

type LogManager interface {
	NewLevel(name string) slog.Leveler
	Named(name string, opts ...any) *slog.Logger
	Iterator(f func(name string, level slog.Leveler) error) error
	IsLogger(name string) bool
	SetLevel(name string, lvl any) bool
	Delete(name string)
	String() string
}

type SlogManagerHandlerOpts func(*slog.HandlerOptions)

type SlogManagerOpts func(*SlogManager)
