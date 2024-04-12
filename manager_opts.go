package slogtool

import (
	"io"
	"log/slog"
)

func WithWriter(out io.Writer) SlogManagerOpts {
	return func(sm *SlogManager) {
		sm.defaultWriter = out
	}
}

func WithDefaultLevel(lvl interface{}) SlogManagerOpts {
	return func(sm *SlogManager) {
		v, _ := slogParseLevel(lvl)
		sm.defaultHandlerOpts.Level = v
	}
}

func WithInternalLevel(lvl interface{}) SlogManagerOpts {
	return func(sm *SlogManager) {
		l := sm.NewLevel(slogManagerInternalName)
		if v, ok := l.(*slog.LevelVar); ok {
			if lv, lok := slogParseLevel(lvl); lok {
				v.Set(lv)
			}
		}
	}
}

func WithSource(withSource bool) SlogManagerHandlerOpts {
	return func(ho *slog.HandlerOptions) {
		ho.AddSource = withSource
	}
}

func WithLevel(lvl interface{}) SlogManagerHandlerOpts {
	return func(ho *slog.HandlerOptions) {
		v, _ := slogParseLevel(lvl)
		ho.Level = v
	}
}

type CustomNewHandler func(name string, w io.Writer, opts *slog.HandlerOptions) slog.Handler

func WithCustomHandler(custom CustomNewHandler) SlogManagerOpts {
	return func(sm *SlogManager) {
		sm.coreNewHandler = custom
	}
}

func WithTextHandler() SlogManagerOpts {
	return func(sm *SlogManager) {
		sm.coreNewHandler = func(_ string, w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewTextHandler(w, opts)
		}
	}
}

func WithJSONHandler() SlogManagerOpts {
	return func(sm *SlogManager) {
		sm.coreNewHandler = func(_ string, w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		}
	}
}
