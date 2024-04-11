package slogtool

import "log/slog"

type loggingOptions struct {
	includeTiming        bool
	includeTimestamp     bool
	includeXForwardedFor bool
	logLevel             slog.Leveler
}

type loggingOptionsFunc func(o *loggingOptions)

// LoggingOptionTiming defines if the logging should contain a `http.request_time` field.
//
//nolint:revive // deliberately not-exported function type.
func LoggingOptionTiming(state bool) loggingOptionsFunc {
	return func(o *loggingOptions) {
		o.includeTiming = state
	}
}

// LoggingOptionTimestamp defines if the logging should contain a `http.timestamp` field.
//
//nolint:revive // deliberately not-exported function type.
func LoggingOptionTimestamp(state bool) loggingOptionsFunc {
	return func(o *loggingOptions) {
		o.includeTimestamp = state
	}
}

// LoggingOptionForwardedFor defines if the logging should contain a `http.forwarded_for` field.
//
//nolint:revive // deliberately not-exported function type.
func LoggingOptionForwardedFor(state bool) loggingOptionsFunc {
	return func(o *loggingOptions) {
		o.includeXForwardedFor = state
	}
}

// LoggingOptionLogLevel defines the log level that http messages should output to,
// defaults to Info.
//
//nolint:revive // deliberately not-exported function type.
func LoggingOptionLogLevel(level slog.Leveler) loggingOptionsFunc {
	return func(o *loggingOptions) {
		o.logLevel = level
	}
}
