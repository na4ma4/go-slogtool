package slogtool

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

// ErrUnimplemented is returned when a method is unimplemented.
var ErrUnimplemented = errors.New("unimplemented method")

// loggingHandler is the [http.Handler] implementation for LoggingHandlerTo and its
// friends.
type loggingHandler struct {
	logger  *slog.Logger
	handler http.Handler
	opts    *loggingOptions
}

// ServeHTTP wraps the next handler ServeHTTP.
func (h loggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	logger := makeLogger(w)
	url := *req.URL
	h.handler.ServeHTTP(logger, req)
	writeLog(context.Background(), &h, req, url, t, logger.Status(), logger.Size())
}

type loggingResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Pusher
	Status() int
	Size() int
}

func makeLogger(w http.ResponseWriter) loggingResponseWriter {
	return &responseLogger{w: w, status: http.StatusOK, size: 0}
}

// responseLogger is wrapper of [http.ResponseWriter] that keeps track of its HTTP
// status code and body size.
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

// Push implements the [http.Pusher] interface, if the underlying ResponseWriter does not implement [http.Pusher],
// it returns ErrUnimplemented.
//
//nolint:wrapcheck // wrapping adds nothing.
func (l *responseLogger) Push(target string, opts *http.PushOptions) error {
	p, ok := l.w.(http.Pusher)
	if !ok {
		return ErrUnimplemented
	}

	return p.Push(target, opts)
}

// Header implements the [http.ResponseWriter] interface.
func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

// Write implements the [http.ResponseWriter] interface, it writes to the underlying ResponseWriter and keeps track of the size of the response body.
func (l *responseLogger) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size

	if err != nil {
		return size, fmt.Errorf("unable to write: %w", err)
	}

	return size, nil
}

// WriteHeader implements the [http.ResponseWriter] interface, it writes the header to the underlying ResponseWriter and keeps track of the status code.
func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

// Status returns the HTTP status code of the response.
func (l *responseLogger) Status() int {
	return l.status
}

// Size returns the size of the response body.
func (l *responseLogger) Size() int {
	return l.size
}

// Flush implements the [http.Flusher] interface, it flushes the underlying ResponseWriter if it implements [http.Flusher].
func (l *responseLogger) Flush() {
	f, ok := l.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

// slogFieldOrSkip is a helper function that returns the provided [slog.Attr] if returnField is true, otherwise it returns an empty [slog.Attr].
func slogFieldOrSkip(returnField bool, field slog.Attr) slog.Attr {
	if returnField {
		return field
	}

	return slog.Attr{}
}

// writeLog writes a log entry for req to w in Apache Combined Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func writeLog(ctx context.Context, lh *loggingHandler, req *http.Request, url url.URL, ts time.Time, status, size int) {
	if lh.opts.ignoreRequestCallback != nil && lh.opts.ignoreRequestCallback(req) {
		return
	}

	// Extract `X-Logging-Username` from request, added by authentication function earlier in process.
	username := "-"
	if lh.opts.extractUsernameCallback != nil {
		if u, ok := lh.opts.extractUsernameCallback(req); ok {
			username = sanitizeUsername(u)
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.RequestURI
	if req.ProtoMajor == 2 && req.Method == http.MethodConnect {
		uri = req.Host
	}

	if uri == "" {
		uri = url.RequestURI()
	}

	fields := []slog.Attr{
		slog.Group("http", // 0
			slog.String("host", host),         // 1
			slog.String("username", username), // 2
			slogFieldOrSkip(lh.opts.includeTimestamp,
				slog.String("timestamp", ts.Format(time.RFC3339Nano)),
			), // 3
			slog.String("method", req.Method),                             // 4
			slog.String("uri", sanitizeURI(uri)),                          // 5
			slog.String("proto", req.Proto),                               // 6
			slog.Int("status", status),                                    // 7
			slog.Int("size", size),                                        // 8
			slog.String("referer", sanitizeURI(req.Referer())),            // 9
			slog.String("user-agent", sanitizeUserAgent(req.UserAgent())), // 10
			slogFieldOrSkip(lh.opts.includeTiming,
				slog.Duration("request-time", time.Since(ts)),
			), // 11
			slogFieldOrSkip(lh.opts.includeXForwardedFor,
				slog.String("forwarded_for", req.Header.Get("X-Forwarded-For")),
			), // 12
		),
	}

	lh.logger.LogAttrs(
		ctx,
		lh.opts.logLevel.Level(),
		"Request",
		fields...,
	)
}

// LoggingHTTPHandler return a [http.Handler] that wraps h and logs requests to out using
// a [slog.Logger].
func LoggingHTTPHandler(logger *slog.Logger, httpHandler http.Handler, opts ...loggingOptionsFunc) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	opt := &loggingOptions{
		includeTiming:        true,
		includeTimestamp:     true,
		includeXForwardedFor: false,
		logLevel:             slog.LevelInfo,
	}

	for _, f := range opts {
		f(opt)
	}

	return loggingHandler{
		logger,
		httpHandler,
		opt,
	}
}

// LoggingHTTPHandlerWrapper is a wrapper for LoggingHTTPHandler that returns a function that can be used in middleware chains.
func LoggingHTTPHandlerWrapper(logger *slog.Logger, opts ...loggingOptionsFunc) func(next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	opt := &loggingOptions{
		includeTiming:        true,
		includeTimestamp:     true,
		includeXForwardedFor: false,
		logLevel:             slog.LevelInfo,
	}

	for _, f := range opts {
		f(opt)
	}

	return func(next http.Handler) http.Handler {
		return loggingHandler{
			logger,
			next,
			opt,
		}
	}
}
