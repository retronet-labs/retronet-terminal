package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	terminal "github.com/retronet-labs/retronet-terminal"
	"github.com/retronet-labs/retronet-terminal/live"
)

const footer = "Ctrl+Q/Ctrl+C esce | Ctrl+L pulisce | Backspace cancella"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, true))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, raw bool) int {
	fs := flag.NewFlagSet("retronet-terminal-live", flag.ContinueOnError)
	fs.SetOutput(stderr)
	width := fs.Int("width", terminal.DefaultWidth, "larghezza dello schermo")
	height := fs.Int("height", terminal.DefaultHeight, "altezza dello schermo")
	demo := fs.Bool("demo", false, "mostra una demo iniziale")
	script := fs.String("script", "", "testo da inviare al terminale e poi uscire")
	lineMode := fs.Bool("line", false, "usa input a righe invece del raw mode")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	err := live.Run(live.Config{
		Width:      *width,
		Height:     *height,
		Input:      stdin,
		Output:     stdout,
		Raw:        raw,
		LineMode:   *lineMode,
		ScriptMode: *script != "",
		Script:     []byte(*script),
		Footer:     footer,
		Handler:    &localHandler{demo: *demo},
	})
	if err != nil {
		fmt.Fprintf(stderr, "errore terminale live: %v\n", err)
		if !*lineMode {
			fmt.Fprintln(stderr, "Riprova da una console interattiva, oppure usa -line per la modalita' a righe.")
		}
		return 1
	}
	return 0
}

type localHandler struct {
	demo    bool
	lineLen int
}

func (h *localHandler) Start(term *terminal.Terminal) error {
	if h.demo {
		_, err := term.Write([]byte("RetroNet Terminal Live\r\n\x1b[3;1HANSI OK\r\nREADY> "))
		return err
	}
	_, err := term.Write([]byte("RetroNet Terminal Live\r\nREADY> "))
	return err
}

func (h *localHandler) HandleByte(term *terminal.Terminal, value byte) (bool, error) {
	switch value {
	case 0x03, 0x04, 0x11: // Ctrl+C, Ctrl+D, Ctrl+Q.
		return false, nil
	case 0x0C: // Ctrl+L.
		_, err := term.Write([]byte("\x1b[2J\x1b[HREADY> "))
		h.lineLen = 0
		return true, err
	case '\r', '\n':
		_, err := term.Write([]byte("\r\nREADY> "))
		h.lineLen = 0
		return true, err
	case '\b', 0x7F:
		if h.lineLen > 0 {
			_, err := term.Write([]byte{'\b', ' ', '\b'})
			h.lineLen--
			return true, err
		}
	default:
		if value == '\t' || value == 0x1B || (value >= 0x20 && value <= 0x7E) {
			err := term.WriteByte(value)
			if value == '\t' || (value >= 0x20 && value <= 0x7E) {
				h.lineLen++
			}
			return true, err
		}
	}
	return true, nil
}
