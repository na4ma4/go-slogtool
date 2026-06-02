package slogtool_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/na4ma4/go-slogtool"
)

func expectLogLines(t *testing.T, rd io.Reader, expect []string) {
	t.Helper()

	body, err := io.ReadAll(rd)
	if err != nil {
		t.Errorf("SlogManager: error reading logs: %s", err)
	}

	lines := []string{}
	for line := range strings.SplitSeq(string(body), "\n") {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	if i, j := len(lines), len(expect); i != j {
		t.Errorf("SlogManager: line count : got '%d' want '%d'", i, j)
	}

	cmpfn := cmp.Options{
		cmp.Comparer(func(i string, j string) bool {
			i = replaceTimefunc(t, i)
			j = replaceTimefunc(t, j)
			return (i == j)
		}),
	}

	if diff := cmp.Diff(lines, expect, cmpfn...); diff != "" {
		t.Errorf("SlogManager: output lines : -got +want:\n%s", diff)
	}
}

func TestSlogManagerDefaultLevel(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
	)

	sublog := testLog.Named("sublog")
	sublog.Debug("debug", slog.String("foo", "bar"))
	testLog.SetLevel("sublog", slog.LevelDebug)
	sublog.Debug("debug2", slog.String("foo2", "bar2"))

	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=debug2 foo2=bar2",
	})
}

func TestSlogManagerChangeDefaultLevelNotInternal(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	sublog := testLog.Named("sublog")
	sublog.Debug("debug", slog.String("foo", "bar"))
	testLog.SetLevel("sublog", slog.LevelInfo)
	sublog.Debug("debug2", slog.String("foo2", "bar2"))

	expectLogLines(t, buf, []string{
		`time=` + timeTestString + ` level=DEBUG msg=debug foo=bar`,
	})
}

func TestSlogManagerChangeDefaultLevelInternal(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelInfo),
		slogtool.WithInternalLevel(slog.LevelDebug),
	)

	sublog := testLog.Named("sublog")
	sublog.Debug("debug", slog.String("round", "1"))
	sublog.Info("info", slog.String("round", "1"))
	testLog.SetLevel("sublog", slog.LevelDebug)
	sublog.Debug("debug", slog.String("round", "2"))
	sublog.Info("info", slog.String("round", "2"))

	expectLogLines(t, buf, []string{
		`time=` + timeTestString + ` level=INFO msg=info round=1`,
		`time=` + timeTestString + ` level=DEBUG msg=SetLevel name=sublog`,
		`time=` + timeTestString + ` level=DEBUG msg="setting level for name" name=sublog match=sublog level=DEBUG`,
		`time=` + timeTestString + ` level=DEBUG msg=debug round=2`,
		`time=` + timeTestString + ` level=INFO msg=info round=2`,
	})
}

func TestSlogManagerWithSource(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
		slogtool.WithSource(true),
	)

	testLog.SetLevel("Internal.SlogManager", slog.LevelDebug)

	sublog := testLog.Named("sublog")
	sublog.Debug("debug", slog.String("foo", "bar"))

	expectLogLines(t, buf, []string{
		`time=` + timeTestString + ` level=DEBUG source=WORKDIR msg=debug foo=bar`,
	})
}

func TestSlogManagerWithTextLevels(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel("debug"),
		slogtool.WithSource(true),
	)

	sublog := testLog.Named("sublog")

	sublog.Debug("debug", slog.String("round", "1"))
	sublog.Info("info", slog.String("round", "1"))
	sublog.Warn("warn", slog.String("round", "1"))
	sublog.Error("error", slog.String("round", "1"))

	testLog.SetLevel("sublog", slog.LevelInfo)

	sublog.Debug("debug", slog.String("round", "2"))
	sublog.Info("info", slog.String("round", "2"))
	sublog.Warn("warn", slog.String("round", "2"))
	sublog.Error("error", slog.String("round", "2"))

	testLog.SetLevel("sublog", slog.LevelWarn)

	sublog.Debug("debug", slog.String("round", "3"))
	sublog.Info("info", slog.String("round", "3"))
	sublog.Warn("warn", slog.String("round", "3"))
	sublog.Error("error", slog.String("round", "3"))

	testLog.SetLevel("sublog", slog.LevelError)

	sublog.Debug("debug", slog.String("round", "4"))
	sublog.Info("info", slog.String("round", "4"))
	sublog.Warn("warn", slog.String("round", "4"))
	sublog.Error("error", slog.String("round", "4"))

	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG source=WORKFILE msg=debug round=1",
		"time=" + timeTestString + " level=INFO source=WORKFILE msg=info round=1",
		"time=" + timeTestString + " level=WARN source=WORKFILE msg=warn round=1",
		"time=" + timeTestString + " level=ERROR source=WORKFILE msg=error round=1",
		"time=" + timeTestString + " level=INFO source=WORKFILE msg=info round=2",
		"time=" + timeTestString + " level=WARN source=WORKFILE msg=warn round=2",
		"time=" + timeTestString + " level=ERROR source=WORKFILE msg=error round=2",
		"time=" + timeTestString + " level=WARN source=WORKFILE msg=warn round=3",
		"time=" + timeTestString + " level=ERROR source=WORKFILE msg=error round=3",
		"time=" + timeTestString + " level=ERROR source=WORKFILE msg=error round=4",
	})
}

func TestSlogManagerString(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel("debug"),
		slogtool.WithSource(true),
	)

	_ = testLog.Named("sublog")
	_ = testLog.Named("sublog2")

	expect := "Internal.SlogManager:ERROR+247,sublog2:DEBUG,sublog:DEBUG"

	if diff := cmp.Diff(testLog.String(), expect); diff != "" {
		t.Errorf("SlogManager: string : -got +want:\n%s", diff)
	}
}

func TestSlogManagerIsLogger(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel("debug"),
		slogtool.WithSource(true),
	)

	_ = testLog.Named("sublog")
	_ = testLog.Named("sublog2")

	if v := testLog.IsLogger("sublog"); !v {
		t.Errorf("SlogManager: IsLogger(sublog) : got '%t', want '%t'", v, true)
	}

	if v := testLog.IsLogger("foolog"); v {
		t.Errorf("SlogManager: IsLogger(foolog) : got '%t', want '%t'", v, false)
	}
}

func TestSlogManagerIterator(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel("debug"),
		slogtool.WithSource(true),
	)

	_ = testLog.Named("sublog")
	_ = testLog.Named("sublog2")

	{
		loggers := []string{}
		err := testLog.Iterator(func(s string, l slog.Leveler) error {
			loggers = append(loggers, s+":"+l.Level().String())
			return nil
		})
		if err != nil {
			t.Errorf("SlogManager: Iterator return : got '%s', want 'nil'", err)
		}

		expect := []string{
			"Internal.SlogManager:ERROR+247",
			"sublog2:DEBUG",
			"sublog:DEBUG",
		}

		sort.Strings(loggers)
		sort.Strings(expect)

		if diff := cmp.Diff(loggers, expect); diff != "" {
			t.Errorf("SlogManager: Iterator loggers : -got +want:\n%s", diff)
		}
	}

	{
		targetErr := errors.New("target error")
		err := testLog.Iterator(func(_ string, _ slog.Leveler) error {
			return targetErr
		})

		if !errors.Is(err, targetErr) {
			t.Errorf("SlogManager: Iterator return : got '%s', want 'target error'", err)
		}
	}
}

func TestSlogManagerNamedOpts(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	sublog0 := testLog.Named("sublog0")
	sublog1 := testLog.Named("sublog1",
		slogtool.WithLevel(slog.LevelInfo),
	)
	sublog2 := testLog.Named("sublog2",
		slog.LevelInfo,
	)
	sublog3 := testLog.Named("sublog3")

	sublog0.DebugContext(ctx, "sublog0:info")
	sublog1.DebugContext(ctx, "sublog1:info")
	sublog2.DebugContext(ctx, "sublog2:info")
	sublog3.DebugContext(ctx, "sublog3:info")

	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=sublog0:info",
		"time=" + timeTestString + " level=DEBUG msg=sublog3:info",
	})
}

func TestSlogManagerSetLevelMatchers(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	sublogaaa := testLog.Named("sublogaaa")
	sublogaca := testLog.Named("sublogaca")
	sublogbaa := testLog.Named("sublogbaa")
	sublogbca := testLog.Named("sublogbca")
	sublogbat := testLog.Named("sublogbat")

	testLog.SetLevel("subloga*", slog.LevelInfo)

	sublogaaa.DebugContext(ctx, "sublogaaa:debug1")
	sublogaca.DebugContext(ctx, "sublogaca:debug1")
	sublogbaa.DebugContext(ctx, "sublogbaa:debug1")
	sublogbca.DebugContext(ctx, "sublogbca:debug1")

	expect := []string{
		"time=" + timeTestString + " level=DEBUG msg=sublogbaa:debug1",
		"time=" + timeTestString + " level=DEBUG msg=sublogbca:debug1",
	}
	expectLogLines(t, buf, expect)

	testLog.SetLevel("*a", slog.LevelInfo)

	sublogaaa.DebugContext(ctx, "sublogaaa:debug2")
	sublogaca.DebugContext(ctx, "sublogaca:debug2")
	sublogbaa.DebugContext(ctx, "sublogbaa:debug2")
	sublogbca.DebugContext(ctx, "sublogbca:debug2")
	sublogbat.DebugContext(ctx, "sublogbat:debug2")

	expect = []string{
		"time=" + timeTestString + " level=DEBUG msg=sublogbat:debug2",
	}
	expectLogLines(t, buf, expect)

	testLog.SetLevel("*", slog.LevelDebug)

	sublogaaa.DebugContext(ctx, "sublogaaa:debug3")
	sublogaca.DebugContext(ctx, "sublogaca:debug3")
	sublogbaa.DebugContext(ctx, "sublogbaa:debug3")
	sublogbca.DebugContext(ctx, "sublogbca:debug3")

	expect = []string{
		"time=" + timeTestString + " level=DEBUG msg=sublogaaa:debug3",
		"time=" + timeTestString + " level=DEBUG msg=sublogaca:debug3",
		"time=" + timeTestString + " level=DEBUG msg=sublogbaa:debug3",
		"time=" + timeTestString + " level=DEBUG msg=sublogbca:debug3",
	}
	expectLogLines(t, buf, expect)

	testLog.SetLevel("*c*", slog.LevelInfo)

	sublogaaa.DebugContext(ctx, "sublogaaa:debug4")
	sublogaca.DebugContext(ctx, "sublogaca:debug4")
	sublogbaa.DebugContext(ctx, "sublogbaa:debug4")
	sublogbca.DebugContext(ctx, "sublogbca:debug4")
	sublogbat.DebugContext(ctx, "sublogbat:debug4")

	expect = []string{
		"time=" + timeTestString + " level=DEBUG msg=sublogaaa:debug4",
		"time=" + timeTestString + " level=DEBUG msg=sublogbaa:debug4",
		"time=" + timeTestString + " level=DEBUG msg=sublogbat:debug4",
	}
	expectLogLines(t, buf, expect)
}

func TestSlogManagerSetLevelValues(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	sublog := testLog.Named("sublog")
	canary := testLog.Named("canary")

	sublog.DebugContext(ctx, "sublog:debug1")
	canary.DebugContext(ctx, "canary:debug1")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=sublog:debug1",
		"time=" + timeTestString + " level=DEBUG msg=canary:debug1",
	})

	// SetLevel error as slog.Level

	testLog.SetLevel("sublog", slog.LevelError)
	sublog.DebugContext(ctx, "sublog:debug2")
	canary.DebugContext(ctx, "canary:debug2")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=canary:debug2",
	})

	// reset to all as int

	testLog.SetLevel("sublog", -255)
	sublog.DebugContext(ctx, "sublog:debug3")
	canary.DebugContext(ctx, "canary:debug3")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=sublog:debug3",
		"time=" + timeTestString + " level=DEBUG msg=canary:debug3",
	})

	// SetLevel info as text

	testLog.SetLevel("sublog", "info")
	sublog.DebugContext(ctx, "sublog:debug4")
	sublog.InfoContext(ctx, "sublog:info4")
	sublog.WarnContext(ctx, "sublog:warn4")
	sublog.ErrorContext(ctx, "sublog:error4")
	canary.DebugContext(ctx, "canary:debug4")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=INFO msg=sublog:info4",
		"time=" + timeTestString + " level=WARN msg=sublog:warn4",
		"time=" + timeTestString + " level=ERROR msg=sublog:error4",
		"time=" + timeTestString + " level=DEBUG msg=canary:debug4",
	})

	// SetLevel warn as text

	testLog.SetLevel("sublog", "warn")
	sublog.DebugContext(ctx, "sublog:debug5")
	sublog.InfoContext(ctx, "sublog:info5")
	sublog.WarnContext(ctx, "sublog:warn5")
	sublog.ErrorContext(ctx, "sublog:error5")
	canary.DebugContext(ctx, "canary:debug5")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=WARN msg=sublog:warn5",
		"time=" + timeTestString + " level=ERROR msg=sublog:error5",
		"time=" + timeTestString + " level=DEBUG msg=canary:debug5",
	})

	// SetLevel error as text

	testLog.SetLevel("sublog", "error")
	sublog.DebugContext(ctx, "sublog:debug6")
	sublog.InfoContext(ctx, "sublog:info6")
	sublog.WarnContext(ctx, "sublog:warn6")
	sublog.ErrorContext(ctx, "sublog:error6")
	canary.DebugContext(ctx, "canary:debug6")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=ERROR msg=sublog:error6",
		"time=" + timeTestString + " level=DEBUG msg=canary:debug6",
	})

	// SetLevel fail using invalid value

	testLog.SetLevel("sublog", true)
	sublog.DebugContext(ctx, "sublog:debug7")
	sublog.InfoContext(ctx, "sublog:info7")
	sublog.WarnContext(ctx, "sublog:warn7")
	sublog.ErrorContext(ctx, "sublog:error7")
	canary.DebugContext(ctx, "canary:debug7")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=ERROR msg=sublog:error7",
		"time=" + timeTestString + " level=DEBUG msg=canary:debug7",
	})
}

func TestSlogManagerSlogManagerInternalLevel(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
		slogtool.WithInternalLevel(slog.LevelDebug),
	)

	sublog := testLog.Named("sublog")
	canary := testLog.Named("canary")

	sublog.DebugContext(ctx, "sublog:debug1")
	canary.DebugContext(ctx, "canary:debug1")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=sublog:debug1",
		"time=" + timeTestString + " level=DEBUG msg=canary:debug1",
	})
}

func TestSlogManagerTextFormatter(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
		slogtool.WithTextHandler(),
	)

	sublog := testLog.Named("sublog")

	sublog.DebugContext(ctx, "sublog:debug1")
	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=sublog:debug1",
	})
}

func TestSlogManagerJSONFormatter(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
		slogtool.WithJSONHandler(),
	)

	sublog := testLog.Named("sublog")

	sublog.DebugContext(ctx, "sublog:debug1")
	expectLogLines(t, buf, []string{
		`{"time":"` + timeTestString + `","level":"DEBUG","msg":"sublog:debug1"}`,
	})
}

func TestSlogManagerMustNewSlogManager(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	testLog := slogtool.MustNewSlogManager(
		context.Background(),
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)
	if testLog == nil {
		t.Fatal("MustNewSlogManager returned nil")
	}
}

func TestSlogManagerMustNewSlogManagerPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from WithDefaultLevel invalid value")
		}
	}()

	slogtool.MustNewSlogManager(
		context.Background(),
		slogtool.WithDefaultLevel(true),
	)
}

func TestSlogManagerDelete(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	_ = testLog.Named("sublog")
	_ = testLog.Named("sublog2")

	if !testLog.IsLogger("sublog") {
		t.Fatal("expected sublog to exist before Delete")
	}

	testLog.Delete("sublog")

	if testLog.IsLogger("sublog") {
		t.Fatal("expected sublog to be deleted")
	}
}

func TestSlogManagerSetLevelReturnsFalse(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	_ = testLog.Named("sublog")

	if ok := testLog.SetLevel("nonexistent", slog.LevelDebug); ok {
		t.Fatal("expected SetLevel to return false for nonexistent logger")
	}
}

func TestSlogManagerSetLevelInvalidValue(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	_ = testLog.Named("sublog")

	if ok := testLog.SetLevel("sublog", true); ok {
		t.Fatal("expected SetLevel to return false for invalid value")
	}
}

func TestSlogManagerNamedWithSlogManagerHandlerOpts(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
	)

	sublog := testLog.Named("sublog",
		slogtool.WithSource(false),
	)
	sublog.DebugContext(ctx, "sublog:debug1")

	expectLogLines(t, buf, []string{
		"time=" + timeTestString + " level=DEBUG msg=sublog:debug1",
	})
}

func TestSlogManagerWithInternalName(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	testLog, _ := slogtool.NewSlogManager(
		context.Background(),
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
		slogtool.WithInternalName("CustomInternal"),
		slogtool.WithInternalLevel(slog.LevelDebug),
	)

	if !testLog.IsLogger("CustomInternal") {
		t.Fatal("expected CustomInternal logger to exist")
	}
}

func TestSlogManagerWithCustomHandler(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
		slogtool.WithCustomHandler(func(_ string, w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		}),
	)

	sublog := testLog.Named("sublog")
	sublog.DebugContext(ctx, "sublog:debug1")

	expectLogLines(t, buf, []string{
		`{"time":"` + timeTestString + `","level":"DEBUG","msg":"sublog:debug1"}`,
	})
}

func TestSlogManagerSubLoggerWith(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	ctx := context.Background()
	testLog, _ := slogtool.NewSlogManager(
		ctx,
		slogtool.WithWriter(buf),
		slogtool.WithDefaultLevel(slog.LevelDebug),
		slogtool.WithJSONHandler(),
	)

	sublog := testLog.Named("sublog")
	sublog.DebugContext(ctx, "sublog:debug1")

	sslog := sublog.With(slog.String("foo", "bar"))
	sslog.DebugContext(ctx, "sublog:debug2")

	testLog.SetLevel("sublog", slog.LevelInfo)

	sslog.DebugContext(ctx, "sublog:debug3")

	testLog.SetLevel("sublog", slog.LevelDebug)

	sslog.DebugContext(ctx, "sublog:debug4")

	expectLogLines(t, buf, []string{
		`{"time":"` + timeTestString + `","level":"DEBUG","msg":"sublog:debug1"}`,
		`{"time":"` + timeTestString + `","level":"DEBUG","msg":"sublog:debug2","foo":"bar"}`,
		`{"time":"` + timeTestString + `","level":"DEBUG","msg":"sublog:debug4","foo":"bar"}`,
	})
}
