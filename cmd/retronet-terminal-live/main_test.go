package main

import (
	"bytes"
	"strings"
	"testing"

	terminal "github.com/retronet-labs/retronet-terminal"
	"github.com/retronet-labs/retronet-terminal/live"
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
	handler := &localHandler{}
	for _, value := range []byte{'A', 'B', '\b', 'C'} {
		keepRunning, err := handler.HandleByte(term, value)
		if err != nil {
			t.Fatal(err)
		}
		if !keepRunning {
			t.Fatalf("unexpected stop")
		}
	}
	if got := term.Snapshot().Rows[0]; got != "AC          " {
		t.Fatalf("row=%q", got)
	}
	keepRunning, err := handler.HandleByte(term, 0x11)
	if err != nil {
		t.Fatal(err)
	}
	if keepRunning {
		t.Fatalf("Ctrl+Q should stop")
	}
}

func TestWriteDeltaDrainsRawOutput(t *testing.T) {
	term := terminal.New(terminal.Config{Width: 12, Height: 3, ANSI: true})
	handler := &localHandler{}
	_, _ = handler.HandleByte(term, 'A')
	_, _ = handler.HandleByte(term, 'B')
	var stdout bytes.Buffer
	if err := live.WriteDelta(&stdout, term); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "AB" {
		t.Fatalf("delta=%q", got)
	}
	if term.Snapshot().OutputBytes != 0 {
		t.Fatalf("output not drained")
	}
}
