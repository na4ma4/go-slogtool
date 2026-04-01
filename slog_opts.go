package slogtool

import (
	"fmt"
	"io"
	"log/slog"
)

// WithWriter is a SlogManagerOpts that sets the default writer for all loggers created by the SlogManager.
func WithWriter(out io.Writer) SlogManagerOpts {
	return func(sm *SlogManager) error {
		sm.defaultWriter = out
		return nil
	}
}

// WithDefaultLevel is a SlogManagerOpts that sets the default log level for all loggers created by the SlogManager.
func WithDefaultLevel(lvl any) SlogManagerOpts {
	return func(sm *SlogManager) error {
		v, ok := slogParseLevel(lvl)
		if !ok {
			return fmt.Errorf("invalid log level: %v", lvl)
		}
		sm.defaultHandlerOpts.Level = v
		return nil
	}
}

// WithInternalLevel is a SlogManagerOpts that sets the log level for the internal logger used by the SlogManager.
//
// NOTE: if this is set before WithInternalName, the internal logger will be created with the default name and level.
func WithInternalLevel(lvl any) SlogManagerOpts {
	return func(sm *SlogManager) error {
		l := sm.NewLevel(sm.iLoggerName)
		if v, ok := l.(*slog.LevelVar); ok {
			if lv, lok := slogParseLevel(lvl); lok {
				v.Set(lv)
			}
		}
		return nil
	}
}

// WithInternalName is a SlogManagerOpts that sets the name of the internal logger used by the SlogManager.
//
// NOTE: if this is set after WithInternalLevel, the internal logger will be created with the default name and level.
func WithInternalName(name string) SlogManagerOpts {
	return func(sm *SlogManager) error {
		sm.iLoggerName = name
		return nil
	}
}

// WithCustomHandler is a SlogManagerOpts that sets a custom handler for all loggers created by the SlogManager.
func WithCustomHandler(custom CustomNewHandler) SlogManagerOpts {
	return func(sm *SlogManager) error {
		sm.coreNewHandler = custom
		return nil
	}
}

// WithTextHandler is a SlogManagerOpts that sets the handler for all loggers created by the SlogManager to a TextHandler.
func WithTextHandler() SlogManagerOpts {
	return func(sm *SlogManager) error {
		sm.coreNewHandler = func(_ string, w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewTextHandler(w, opts)
		}
		return nil
	}
}

// WithJSONHandler is a SlogManagerOpts that sets the handler for all loggers created by the SlogManager to a JSONHandler.
func WithJSONHandler() SlogManagerOpts {
	return func(sm *SlogManager) error {
		sm.coreNewHandler = func(_ string, w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		}
		return nil
	}
}
