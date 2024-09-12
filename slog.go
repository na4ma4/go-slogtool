package slogtool

import (
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
)

const (
	slogManagerInternalName         = "Internal.SlogManager"
	slogManagerInternalDefaultLevel = 255
)

// type slogHandlerNewFunc func(io.Writer, *slog.HandlerOptions) slog.Handler

// SlogManager provides a wrapper for multiple *slog.Logger levels,
// the individual loggers are not kept, but levels are kept
// indexed by name.
type SlogManager struct {
	coreNewHandler     CustomNewHandler
	defaultHandlerOpts *slog.HandlerOptions
	defaultWriter      io.Writer
	iLogger            *slog.Logger
	levels             map[string]*slog.LevelVar
	lock               sync.RWMutex
}

// NewSlogManager returns a new SlogManager ready for use.
func NewSlogManager(opts ...interface{}) *SlogManager {
	defaultLevel := new(slog.LevelVar)
	defaultLevel.Set(slog.LevelInfo)
	var defaultWriter io.Writer = os.Stdout

	defaultHandlerOpts := &slog.HandlerOptions{
		Level: defaultLevel,
	}

	for _, opti := range opts {
		if opt, ok := opti.(SlogManagerHandlerOpts); ok {
			opt(defaultHandlerOpts)
		}
	}

	out := &SlogManager{
		coreNewHandler: func(_ string, w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewTextHandler(w, opts)
		},
		defaultHandlerOpts: defaultHandlerOpts,
		defaultWriter:      defaultWriter,
		levels:             map[string]*slog.LevelVar{},
		lock:               sync.RWMutex{},
	}

	out.levels[slogManagerInternalName] = new(slog.LevelVar)
	out.levels[slogManagerInternalName].Set(slogManagerInternalDefaultLevel)

	for _, opti := range opts {
		if opt, ok := opti.(SlogManagerOpts); ok {
			opt(out)
		}
	}

	out.iLogger = out.Named(slogManagerInternalName)

	return out
}

// NewLevel returns as a log.Leveler reference to the stored named level.
func (a *SlogManager) NewLevel(name string) slog.Leveler {
	a.lock.Lock()
	defer a.lock.Unlock()

	if _, ok := a.levels[name]; !ok {
		a.levels[name] = new(slog.LevelVar)
		a.levels[name].Set(a.defaultHandlerOpts.Level.Level())
	}

	return a.levels[name]
}

func slogParseLevel(v interface{}) (slog.Level, bool) {
	switch lvl := v.(type) {
	case slog.Leveler:
		return lvl.Level(), true
	case int:
		return slog.Level(lvl), true
	case string:
		switch v := strings.ToLower(lvl); v {
		case "debug", "debg", "d":
			return slog.LevelDebug, true
		case "info", "inf", "i":
			return slog.LevelInfo, true
		case "warn", "wrn", "w":
			return slog.LevelWarn, true
		case "error", "erro", "err", "e":
			return slog.LevelError, true
		}
	}

	return slog.LevelInfo, false
}

// doesKeyMatch tests if a key matches.
func (a *SlogManager) doesKeyMatch(key, check string) bool {
	if strings.EqualFold(key, check) {
		return true
	}

	// if it's the internal logger, don't match wildcards.
	if strings.EqualFold(key, slogManagerInternalName) {
		return false
	}

	// if is is a single '*' then it's a wildcard and true.
	if len(check) == 1 && check == "*" {
		return true
	}

	switch {
	case strings.HasPrefix(check, "*") && strings.HasSuffix(check, "*"):
		// wildcard both ends (can't be a single * otherwise initial check would fail)
		// log.Printf("strings.Contains(%s, %s): %t",
		// 	key, check[1:len(check)-1], strings.Contains(key, check[1:len(check)-1]))
		return strings.Contains(key, check[1:len(check)-1])
	case strings.HasPrefix(check, "*"):
		// wildcard at start
		return strings.HasSuffix(key, check[1:])
	case strings.HasSuffix(check, "*"):
		// wildcard at end
		return strings.HasPrefix(key, check[:len(check)-1])
	}

	return false
}

// SetLevel attempts to set the level supplied, it will attempt to typecast the value
// against string, slog.Level and slog.Leveler.
func (a *SlogManager) SetLevel(name string, lvl interface{}) bool {
	a.iLogger.Debug("SetLevel", slog.String("name", name))

	found := false

	a.lock.Lock()
	defer a.lock.Unlock()

	for itemKey, val := range a.levels {
		if a.doesKeyMatch(itemKey, name) {
			if level, ok := slogParseLevel(lvl); ok {
				a.iLogger.Debug(
					"setting level for name",
					slog.String("name", name),
					slog.String("match", itemKey),
					slog.String("level", level.String()),
				)
				val.Set(level)

				found = true
			}
		}
	}

	return found
}

// String returns a string representation of the currently stored loggers and their levels.
func (a *SlogManager) String() string {
	a.lock.RLock()
	defer a.lock.RUnlock()

	out := []string{}

	for k, v := range a.levels {
		out = append(out, k+":"+v.Level().String())
	}

	sort.Strings(out)

	return strings.Join(out, ",")
}

// Iterator runs a callback function over the levels map item by item.
func (a *SlogManager) Iterator(f func(string, slog.Leveler) error) error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	for k, v := range a.levels {
		if err := f(k, v); err != nil {
			return err
		}
	}

	return nil
}

// IsLogger returns true if there is a logger that matches.
func (a *SlogManager) IsLogger(name string) bool {
	a.lock.RLock()
	defer a.lock.RUnlock()

	_, ok := a.levels[name]
	return ok
}

// Delete removes the named logger entry from the list.
func (a *SlogManager) Delete(name string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.levels, name)
}

// Named returns a named *slog.Logger if any additional parameters are specified it will
// try to determine if they represent a log level (by string, zapcore.Level or slog.Leveler).
func (a *SlogManager) Named(name string, opts ...interface{}) *slog.Logger {
	lvl := a.NewLevel(name)

	handlerOpts := &slog.HandlerOptions{
		AddSource:   a.defaultHandlerOpts.AddSource,
		Level:       lvl,
		ReplaceAttr: a.defaultHandlerOpts.ReplaceAttr,
	}

	for _, opt := range opts {
		switch v := opt.(type) {
		case int, slog.Level, slog.Leveler:
			a.SetLevel(name, v)
		case SlogManagerHandlerOpts:
			v(handlerOpts)
		}
	}

	namedLogger := a.coreNewHandler(name, a.defaultWriter, handlerOpts)

	return slog.New(namedLogger)
}
