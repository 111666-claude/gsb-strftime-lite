package main

import (
	"bytes"
	"testing"
)

func run(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func TestDefaultOutput(t *testing.T) {
	stdout, _, code := run(t)
	if code != 0 || stdout != "2026-02-02 02:40:00\n" {
		t.Fatalf("默认输出不对：code=%d out=%q", code, stdout)
	}
}

func TestPercentLayout(t *testing.T) {
	stdout, _, code := run(t, "--layout", "100%% done")
	if code != 0 || stdout != "100% done\n" {
		t.Fatalf("%%%% 应还原：code=%d out=%q", code, stdout)
	}
}

func TestParseFlag(t *testing.T) {
	stdout, _, code := run(t, "--parse", "2026-02-02 02:40:00", "--layout", "%Y-%m-%d %H:%M:%S")
	if code != 0 || stdout != "1770000000\n" {
		t.Fatalf("解析输出不对：code=%d out=%q", code, stdout)
	}
}

func TestParseFailure(t *testing.T) {
	_, stderr, code := run(t, "--parse", "2026-02-30 02:40:00")
	if code != 1 || stderr == "" {
		t.Fatalf("非法日期应报错：code=%d err=%q", code, stderr)
	}
}

func TestUnknownDirective(t *testing.T) {
	_, stderr, code := run(t, "--layout", "%q")
	if code != 1 || stderr == "" {
		t.Fatalf("未知指令应报错：code=%d err=%q", code, stderr)
	}
}
