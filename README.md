# go-slogtool

[![CI](https://github.com/na4ma4/go-slogtool/workflows/CI/badge.svg)](https://github.com/na4ma4/go-slogtool/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/na4ma4/go-slogtool/?status.svg)](https://godoc.org/github.com/na4ma4/go-slogtool)
[![GitHub issues](https://img.shields.io/github/issues/na4ma4/go-slogtool)](https://github.com/na4ma4/go-slogtool/issues)
[![GitHub forks](https://img.shields.io/github/forks/na4ma4/go-slogtool)](https://github.com/na4ma4/go-slogtool/network)
[![GitHub stars](https://img.shields.io/github/stars/na4ma4/go-slogtool)](https://github.com/na4ma4/go-slogtool/stargazers)
[![GitHub license](https://img.shields.io/github/license/na4ma4/go-slogtool)](https://github.com/na4ma4/go-slogtool/blob/main/LICENSE)

[log/slog](https://pkg.go.dev/log/slog) wrappers and tools.

## Install

```shell
go get -u github.com/na4ma4/go-slogtool
```

## Tools

### LogLevels

```golang
ctx := context.Background()
logmgr := slogtool.NewSlogManager(
    ctx,
    slogtool.WithDefaultLevel(slog.LevelDebug),
)

processOne := server.NewProcess(logmgr.Named("Server.Process"))

// somewhere else.

logmgr.SetLevel("Server.Process", "debug")

// and triggered somewhere else again.

logmgr.SetLevel("Server.Process", "info")
```

### HTTP Logging Handler

```golang
r := mux.NewRouter()
r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is a catch-all route"))
})

loggedRouter := slogtool.LoggingHTTPHandler(logmgr.Named("WebServer"), r)
http.ListenAndServe(":1123", loggedRouter)
```
