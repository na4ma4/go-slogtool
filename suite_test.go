package slogtool_test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	timeTestString = `2006-01-02T15:04:05.999+07:00`
)

func replaceSourcePath(t *testing.T, in string) string {
	t.Helper()

	idx := strings.Index(in, "source=")
	if idx > 0 {
		if i, j := strings.Index(in[idx:], ":")+idx, strings.Index(in[idx:], " ")+idx; i > 0 && j > 0 {
			in = in[:idx] + "source=WORKFILE" + in[j:]
		}
	}

	return in
}

func TestReplaceSourcePath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("ReplaceSourcePath: error : got '%s' want 'nil'", err)
	}

	line := "time=" + timeTestString + " " +
		"level=DEBUG " +
		"source=" + wd + ":999 " +
		"msg=debug foo=bar"

	expect := "time=" + timeTestString + " " +
		"level=DEBUG " +
		"source=WORKFILE " +
		"msg=debug foo=bar"

	if diff := cmp.Diff(replaceSourcePath(t, line), expect); diff != "" {
		t.Errorf("ReplaceSourcePath: output lines : -got +want:\n%s", diff)
	}
}

func replaceTimefunc(t *testing.T, in string) string {
	if idx := strings.Index(in, " "); strings.HasPrefix(in, "time=") && idx > 0 {
		in = replaceSourcePath(t, in[idx:])
	}
	if idx := strings.Index(in, `"time":"`); idx > 0 {
		idx += 8
		iidx := strings.Index(in[idx:], `"`) + idx
		in = in[:idx] + timeTestString + in[iidx:]
	}
	return in
}

func TestReplaceTime(t *testing.T) {
	{ // Text
		input := `time=2024-04-11T14:40:27.064+10:00 level=DEBUG msg=debug2 foo2=bar2`
		expect := ` level=DEBUG msg=debug2 foo2=bar2`

		time := replaceTimefunc(t, input)
		if diff := cmp.Diff(time, expect); diff != "" {
			t.Errorf("ReplaceTime[JSON]: output lines : -got +want:\n%s", diff)
		}
	}

	{ // JSON
		input := `{"time":"2024-04-11T14:40:27.064+10:00","level":"DEBUG","msg":"sublog:debug1"}`
		expect := `{"time":"` + timeTestString + `","level":"DEBUG","msg":"sublog:debug1"}`

		time := replaceTimefunc(t, input)
		if diff := cmp.Diff(time, expect); diff != "" {
			t.Errorf("ReplaceTime[JSON]: output lines : -got +want:\n%s", diff)
		}
	}
}
