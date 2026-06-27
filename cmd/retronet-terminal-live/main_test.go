package main

import (
	"bytes"
	"strings"
	"testing"

	terminal "github.com/retronet-labs/retronet-terminal"
)

func TestRunScriptRendersLiveScreen(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-width", "40", "-height", "6", "-script", "HELLO\r\nBYE"}, strings.NewReader(""), &stdout, &stderr, false)
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "RetroNet Terminal Live") || !strings.Contains(output, "HELLO") || !strings.Contains(output, "BYE") {
		t.Fatalf("stdout=%q", output)
	}
	if !strings.Contains(output, footer) {
		t.Fatalf("footer missing in stdout=%q", output)
	}
}

func TestHandleByteControls(t *testing.T) {
	term := terminal.New(terminal.Config{Width: 12, Height: 3, ANSI: true})
	for _, value := range []byte{'A', 'B', '\b', 'C'} {
		if !handleByte(term, value) {
			t.Fatalf("unexpected stop")
		}
	}
	if got := term.Snapshot().Rows[0]; got != "AC          " {
		t.Fatalf("row=%q", got)
	}
	if handleByte(term, 0x11) {
		t.Fatalf("Ctrl+Q should stop")
	}
}
