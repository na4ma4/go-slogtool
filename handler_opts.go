package slogtool

import (
	"log/slog"
	"net/http"
)

type (
	LoggingIgnoreRequestCallback   func(req *http.Request) bool
	LoggingExtractUsernameCallback func(req *http.Request) (string, bool)
)

type loggingOptions struct {
	includeTiming           bool
	includeTimestamp        bool
	includeXForwardedFor    bool
	ignoreRequestCallback   LoggingIgnoreRequestCallback
	extractUsernameCallback LoggingExtractUsernameCallback
	logLevel                slog.Leveler
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

// LoggingOptionIgnoreRequest defines a callback that determines if a request should be ignored in logging.
//
//nolint:revive // deliberately not-exported function type.
func LoggingOptionIgnoreRequest(callback LoggingIgnoreRequestCallback) loggingOptionsFunc {
	return func(o *loggingOptions) {
		o.ignoreRequestCallback = callback
	}
}

// LoggingOptionExtractUsername defines a callback that extracts the username from a request.
//
//nolint:revive // deliberately not-exported function type.
func LoggingOptionExtractUsername(callback LoggingExtractUsernameCallback) loggingOptionsFunc {
	return func(o *loggingOptions) {
		o.extractUsernameCallback = callback
	}
}
