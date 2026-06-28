package live

import (
	"bytes"
	"strings"
	"testing"

	terminal "github.com/retronet-labs/retronet-terminal"
)

func TestRunScriptRendersSnapshot(t *testing.T) {
	var stdout bytes.Buffer
	handler := &echoHandler{}
	err := Run(Config{
		Width:      20,
		Height:     4,
		Output:     &stdout,
		ScriptMode: true,
		Script:     []byte("HI\r\n"),
		Footer:     "test footer",
		Handler:    handler,
	})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, want := range []string{"READY> HI", "test footer"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in %q", want, output)
		}
	}
}

func TestWriteDeltaDrainsRawOutput(t *testing.T) {
	term := terminal.New(terminal.Config{Width: 20, Height: 4, ANSI: true})
	_, _ = term.Write([]byte("ABC"))
	var stdout bytes.Buffer
	if err := WriteDelta(&stdout, term); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "ABC" {
		t.Fatalf("delta=%q", got)
	}
	if term.Snapshot().OutputBytes != 0 {
		t.Fatalf("output not drained")
	}
}

type echoHandler struct{}

func (h *echoHandler) Start(term *terminal.Terminal) error {
	_, err := term.Write([]byte("READY> "))
	return err
}

func (h *echoHandler) HandleByte(term *terminal.Terminal, value byte) (bool, error) {
	switch value {
	case '\r', '\n':
		_, err := term.Write([]byte("\r\nREADY> "))
		return true, err
	case 0x11:
		return false, nil
	default:
		return true, term.WriteByte(value)
	}
}
