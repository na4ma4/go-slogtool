package slogtool_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/na4ma4/go-slogtool"
)

func readSingleLogObject(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected one log line, got empty output")
	}

	var item map[string]any
	if err := json.Unmarshal([]byte(line), &item); err != nil {
		t.Fatalf("expected JSON log line, got error: %v", err)
	}

	return item
}

func TestLoggingHTTPHandlerOptionsApplied(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	base := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	h := slogtool.LoggingHTTPHandler(
		base,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		slogtool.LoggingOptionTiming(false),
		slogtool.LoggingOptionTimestamp(false),
		slogtool.LoggingOptionForwardedFor(true),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test?q=1", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.10")
	req.Header.Set("Referer", "http://example.com/ref")
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:2222"

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("status code mismatch: got=%d want=%d", rw.Code, http.StatusOK)
	}

	obj := readSingleLogObject(t, buf)
	if lvl, ok := obj["level"].(string); !ok || lvl != "INFO" {
		t.Fatalf("level mismatch: got=%v want=INFO", obj["level"])
	}

	var httpItem map[string]any
	{
		var ok bool
		httpItem, ok = obj["http"].(map[string]any)
		if !ok {
			t.Fatalf("expected http group, got=%T", obj["http"])
		}
	}

	if _, ok := httpItem["request-time"]; ok {
		t.Fatalf("request-time should be omitted when timing is disabled: %v", httpItem["request-time"])
	}
	if _, ok := httpItem["timestamp"]; ok {
		t.Fatalf("timestamp should be omitted when timestamp is disabled: %v", httpItem["timestamp"])
	}
	if fwd, ok := httpItem["forwarded_for"].(string); !ok || fwd != "198.51.100.10" {
		t.Fatalf("forwarded_for mismatch: got=%v", httpItem["forwarded_for"])
	}
}

func TestLoggingHTTPHandlerWrapperRespectsLogLevelOption(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	base := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mw := slogtool.LoggingHTTPHandlerWrapper(
		base,
		slogtool.LoggingOptionLogLevel(slog.LevelWarn),
	)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("done"))
	}))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/create", nil)
	req.RemoteAddr = "127.0.0.1:3333"

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatalf("status code mismatch: got=%d want=%d", rw.Code, http.StatusCreated)
	}

	obj := readSingleLogObject(t, buf)
	if lvl, ok := obj["level"].(string); !ok || lvl != "WARN" {
		t.Fatalf("level mismatch: got=%v want=WARN", obj["level"])
	}

	var httpItem map[string]any
	{
		var ok bool
		httpItem, ok = obj["http"].(map[string]any)
		if !ok {
			t.Fatalf("expected http group, got=%T", obj["http"])
		}
	}

	if status, ok := httpItem["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Fatalf("status mismatch: got=%v want=%d", httpItem["status"], http.StatusCreated)
	}

	if size, ok := httpItem["size"].(float64); !ok || int(size) != len("done") {
		t.Fatalf("size mismatch: got=%v want=%d", httpItem["size"], len("done"))
	}
}

func TestLoggingHTTPHandlerIgnoreRequestCallback(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	base := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	h := slogtool.LoggingHTTPHandler(
		base,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		slogtool.LoggingOptionIgnoreRequest(func(req *http.Request) bool {
			return strings.EqualFold(req.Header.Get("X-Logging-Noop"), "true")
		}),
	)

	skipReq := httptest.NewRequest(http.MethodGet, "http://example.com/skip", nil)
	skipReq.RemoteAddr = "127.0.0.1:4444"
	skipReq.Header.Set("X-Logging-Noop", "true")

	skipRW := httptest.NewRecorder()
	h.ServeHTTP(skipRW, skipReq)

	if skipRW.Code != http.StatusOK {
		t.Fatalf("skip request status mismatch: got=%d want=%d", skipRW.Code, http.StatusOK)
	}

	logReq := httptest.NewRequest(http.MethodGet, "http://example.com/log", nil)
	logReq.RemoteAddr = "127.0.0.1:5555"

	logRW := httptest.NewRecorder()
	h.ServeHTTP(logRW, logReq)

	if logRW.Code != http.StatusOK {
		t.Fatalf("log request status mismatch: got=%d want=%d", logRW.Code, http.StatusOK)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly one log line after skip+log requests, got=%d", len(lines))
	}

	obj := readSingleLogObject(t, buf)
	httpItem, ok := obj["http"].(map[string]any)
	if !ok {
		t.Fatalf("expected http group, got=%T", obj["http"])
	}
	uri, uriOK := httpItem["uri"].(string)
	if !uriOK || !strings.Contains(uri, "/log") {
		t.Fatalf("uri mismatch: got=%v want contains=%s", httpItem["uri"], "/log")
	}
}

func TestLoggingHTTPHandlerIgnoreRequestCallbackPanic(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	base := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	h := slogtool.LoggingHTTPHandler(
		base,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		slogtool.LoggingOptionIgnoreRequest(func(_ *http.Request) bool {
			panic("ignore callback panic")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/panic", nil)
	req.RemoteAddr = "127.0.0.1:7777"
	rw := httptest.NewRecorder()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from ignore callback, got nil")
		}
		rStr, ok := r.(string)
		if !ok {
			t.Fatalf("unexpected panic type: got=%T", r)
		}
		if !strings.Contains(rStr, "ignore callback panic") {
			t.Fatalf("unexpected panic value: got=%v", rStr)
		}
		if rw.Code != http.StatusOK {
			t.Fatalf("status code mismatch: got=%d want=%d", rw.Code, http.StatusOK)
		}
		if strings.TrimSpace(buf.String()) != "" {
			t.Fatalf("expected no log output when callback panics, got=%q", buf.String())
		}
	}()

	h.ServeHTTP(rw, req)
}

func TestLoggingHTTPHandlerExtractUsernameCallback(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	base := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	h := slogtool.LoggingHTTPHandler(
		base,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		slogtool.LoggingOptionExtractUsername(func(_ *http.Request) (string, bool) {
			return `<admin>`, true
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/user", nil)
	req.RemoteAddr = "127.0.0.1:8888"
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("status code mismatch: got=%d want=%d", rw.Code, http.StatusOK)
	}

	obj := readSingleLogObject(t, buf)
	httpItem, ok := obj["http"].(map[string]any)
	if !ok {
		t.Fatalf("expected http group, got=%T", obj["http"])
	}

	username, ok := httpItem["username"].(string)
	if !ok || username != "&lt;admin&gt;" {
		t.Fatalf("username mismatch: got=%v want=%s", httpItem["username"], "&lt;admin&gt;")
	}
}

func TestLoggingHTTPHandlerExtractUsernameCallbackFallback(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	base := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	h := slogtool.LoggingHTTPHandler(
		base,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		slogtool.LoggingOptionExtractUsername(func(_ *http.Request) (string, bool) {
			return "", false
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/user", nil)
	req.RemoteAddr = "127.0.0.1:8889"
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("status code mismatch: got=%d want=%d", rw.Code, http.StatusOK)
	}

	obj := readSingleLogObject(t, buf)
	httpItem, ok := obj["http"].(map[string]any)
	if !ok {
		t.Fatalf("expected http group, got=%T", obj["http"])
	}

	username, ok := httpItem["username"].(string)
	if !ok || username != "-" {
		t.Fatalf("username mismatch: got=%v want=%s", httpItem["username"], "-")
	}
}

type ignoreRequestCase struct {
	name      string
	ignore    func(req *http.Request) bool
	skipPath  string
	skipQuery string
	skipHdr   string
	logPath   string
	logQuery  string
	logHdr    string
}

func buildRequest(path, query, headerValue, remote string) *http.Request {
	requestURL := "http://example.com" + path
	if query != "" {
		requestURL += "?" + query
	}

	req := httptest.NewRequest(http.MethodGet, requestURL, nil)
	req.RemoteAddr = remote
	if headerValue != "" {
		req.Header.Set("X-Logging-Noop", headerValue)
	}

	return req
}

func runIgnoreRequestCase(t *testing.T, tc ignoreRequestCase) {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	base := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	h := slogtool.LoggingHTTPHandler(
		base,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		slogtool.LoggingOptionIgnoreRequest(tc.ignore),
	)

	skipReq := buildRequest(tc.skipPath, tc.skipQuery, tc.skipHdr, "127.0.0.1:6661")
	skipRW := httptest.NewRecorder()
	h.ServeHTTP(skipRW, skipReq)
	if skipRW.Code != http.StatusOK {
		t.Fatalf("skip request status mismatch: got=%d want=%d", skipRW.Code, http.StatusOK)
	}

	logReq := buildRequest(tc.logPath, tc.logQuery, tc.logHdr, "127.0.0.1:6662")
	logRW := httptest.NewRecorder()
	h.ServeHTTP(logRW, logReq)
	if logRW.Code != http.StatusOK {
		t.Fatalf("log request status mismatch: got=%d want=%d", logRW.Code, http.StatusOK)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly one log line after skip+log requests, got=%d", len(lines))
	}

	obj := readSingleLogObject(t, buf)
	httpItem, ok := obj["http"].(map[string]any)
	if !ok {
		t.Fatalf("expected http group, got=%T", obj["http"])
	}

	uri, ok := httpItem["uri"].(string)
	if !ok || !strings.Contains(uri, tc.logPath) {
		t.Fatalf("uri mismatch: got=%v want contains=%s", httpItem["uri"], tc.logPath)
	}
}

func TestLoggingHTTPHandlerIgnoreRequestCallbackTableDriven(t *testing.T) {
	tests := []ignoreRequestCase{
		{
			name: "header rule",
			ignore: func(req *http.Request) bool {
				return strings.EqualFold(req.Header.Get("X-Logging-Noop"), "yes")
			},
			skipPath: "/skip-h",
			skipHdr:  "yes",
			logPath:  "/log-h",
		},
		{
			name: "query rule",
			ignore: func(req *http.Request) bool {
				return strings.EqualFold(req.URL.Query().Get("nolog"), "1")
			},
			skipPath:  "/skip-q",
			skipQuery: "nolog=1",
			logPath:   "/log-q",
		},
		{
			name: "path rule",
			ignore: func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, "/internal")
			},
			skipPath: "/internal/healthz",
			logPath:  "/public/ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { runIgnoreRequestCase(t, tt) })
	}
}
