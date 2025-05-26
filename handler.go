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

const (
	HeaderUsername = "X-Logging-Username"
	HeaderNoop     = "X-Logging-Noop"
)

// ErrUnimplemented is returned when a method is unimplemented.
var ErrUnimplemented = errors.New("unimplemented method")

// loggingHandler is the http.Handler implementation for LoggingHandlerTo and its
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
	req.Header.Del(HeaderNoop)
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

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP
// status code and body size.
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

//nolint:wrapcheck // wrapping adds nothing.
func (l *responseLogger) Push(target string, opts *http.PushOptions) error {
	p, ok := l.w.(http.Pusher)
	if !ok {
		return ErrUnimplemented
	}

	return p.Push(target, opts)
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size

	if err != nil {
		return size, fmt.Errorf("unable to write: %w", err)
	}

	return size, nil
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) Flush() {
	f, ok := l.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

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
	if req.Header.Get(HeaderNoop) != "" {
		return
	}

	// Extract `X-Logging-Username` from request, added by authentication function earlier in process.
	username := "-"
	if req.Header.Get(HeaderUsername) != "" {
		username = sanitizeUsername(req.Header.Get(HeaderUsername))
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

// LoggingHTTPHandler return a http.Handler that wraps h and logs requests to out using
// a *slog.Logger.
func LoggingHTTPHandler(logger *slog.Logger, httpHandler http.Handler, opts ...loggingOptionsFunc) http.Handler {
	opt := &loggingOptions{
		includeTiming:        true,
		includeTimestamp:     true,
		includeXForwardedFor: false,
		logLevel:             slog.LevelInfo,
	}

	for _, f := range opts {
		f(opt)
	}

	logger = slog.New(
		NewSlogHandlerWrapper(
			logger.Handler(),
			opt.logLevel,
		),
	)

	// logger = slog.New(
	// 	logger.Handler().
	// )

	return loggingHandler{
		logger,
		httpHandler,
		opt,
	}
}

func LoggingHTTPHandlerWrapper(logger *slog.Logger, opts ...loggingOptionsFunc) func(next http.Handler) http.Handler {
	opt := &loggingOptions{
		includeTiming:        true,
		includeTimestamp:     true,
		includeXForwardedFor: false,
		logLevel:             slog.LevelInfo,
	}

	for _, f := range opts {
		f(opt)
	}

	logger = slog.New(
		NewSlogHandlerWrapper(
			logger.Handler(),
			opt.logLevel,
		),
	)

	return func(next http.Handler) http.Handler {
		return loggingHandler{
			logger,
			next,
			opt,
		}
	}
}
