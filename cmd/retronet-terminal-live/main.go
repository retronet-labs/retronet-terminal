package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	terminal "github.com/retronet-labs/retronet-terminal"
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
	if err := fs.Parse(args); err != nil {
		return 2
	}

	term := terminal.New(terminal.Config{Width: *width, Height: *height, ANSI: true})
	if *demo {
		_, _ = term.Write([]byte("RetroNet Terminal Live\r\n\x1b[3;1HANSI OK\r\nREADY> "))
	} else {
		_, _ = term.Write([]byte("RetroNet Terminal Live\r\nREADY> "))
	}

	fmt.Fprint(stdout, "\x1b[2J")
	if *script != "" {
		feedBytes(term, []byte(*script))
		render(stdout, term)
		fmt.Fprint(stdout, "\x1b[?25h\r\n")
		return 0
	}

	restore := func() error { return nil }
	if raw {
		if inFile, ok := stdin.(*os.File); ok {
			if outFile, ok := stdout.(*os.File); ok {
				nextRestore, err := enterConsoleRaw(inFile, outFile)
				if err != nil {
					fmt.Fprintf(stderr, "raw mode non disponibile, continuo in modalita' linea: %v\n", err)
				} else {
					restore = nextRestore
				}
			}
		}
	}
	defer func() {
		_ = restore()
		fmt.Fprint(stdout, "\x1b[?25h\r\n")
	}()

	render(stdout, term)
	buf := make([]byte, 1)
	for {
		n, err := stdin.Read(buf)
		if n > 0 {
			if !handleByte(term, buf[0]) {
				return 0
			}
			render(stdout, term)
		}
		if err == io.EOF {
			return 0
		}
		if err != nil {
			fmt.Fprintf(stderr, "lettura input fallita: %v\n", err)
			return 1
		}
	}
}

func feedBytes(term *terminal.Terminal, data []byte) {
	for _, value := range data {
		if !handleByte(term, value) {
			return
		}
	}
}

func handleByte(term *terminal.Terminal, value byte) bool {
	switch value {
	case 0x03, 0x04, 0x11: // Ctrl+C, Ctrl+D, Ctrl+Q.
		return false
	case 0x0C: // Ctrl+L.
		_, _ = term.Write([]byte("\x1b[2J\x1b[HREADY> "))
	case '\r', '\n':
		_, _ = term.Write([]byte("\r\nREADY> "))
	case '\b', 0x7F:
		_, _ = term.Write([]byte{'\b', ' ', '\b'})
	default:
		if value == '\t' || value == 0x1B || (value >= 0x20 && value <= 0x7E) {
			_ = term.WriteByte(value)
		}
	}
	return true
}

func render(stdout io.Writer, term *terminal.Terminal) {
	snapshot := term.Snapshot()
	fmt.Fprint(stdout, "\x1b[?25l\x1b[H")
	for row, text := range snapshot.Rows {
		if row > 0 {
			fmt.Fprint(stdout, "\r\n")
		}
		fmt.Fprint(stdout, text)
	}
	fmt.Fprintf(stdout, "\r\n\x1b[2K%s", footer)
	fmt.Fprintf(stdout, "\x1b[%d;%dH\x1b[?25h", snapshot.CursorRow+1, snapshot.CursorCol+1)
}
