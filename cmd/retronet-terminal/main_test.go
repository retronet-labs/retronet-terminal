package main

import (
	"bytes"
	"testing"
)

func TestRunDemoScreen(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-demo", "-screen"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("RetroNet Terminal")) || !bytes.Contains(stdout.Bytes(), []byte("ANSI OK")) {
		t.Fatalf("stdout=%q", stdout.String())
	}
}
