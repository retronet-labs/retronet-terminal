// Package live contiene il runner interattivo riusabile per terminali RetroNet.
package live

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	terminal "github.com/retronet-labs/retronet-terminal"
)

const DefaultFooter = "Ctrl+Q/Ctrl+C esce | Ctrl+L pulisce"

var (
	ErrNilHandler         = errors.New("handler live non inizializzato")
	ErrRawModeUnavailable = errors.New("raw mode non disponibile")
)

type Handler interface {
	Start(term *terminal.Terminal) error
	HandleByte(term *terminal.Terminal, value byte) (bool, error)
}

type Config struct {
	Terminal   *terminal.Terminal
	Width      int
	Height     int
	Input      io.Reader
	Output     io.Writer
	Raw        bool
	LineMode   bool
	ScriptMode bool
	Script     []byte
	Footer     string
	Handler    Handler
}

func Run(config Config) error {
	if config.Handler == nil {
		return ErrNilHandler
	}
	input := config.Input
	if input == nil {
		input = strings.NewReader("")
	}
	output := config.Output
	if output == nil {
		output = io.Discard
	}
	term := config.Terminal
	if term == nil {
		term = terminal.New(terminal.Config{
			Width:  config.Width,
			Height: config.Height,
			ANSI:   true,
		})
	}
	if err := config.Handler.Start(term); err != nil {
		return err
	}

	fmt.Fprint(output, "\x1b[2J")
	if config.ScriptMode {
		if err := FeedBytes(config.Handler, term, config.Script); err != nil {
			return err
		}
		RenderSnapshot(output, term, config.Footer)
		fmt.Fprint(output, "\x1b[?25h\r\n")
		return nil
	}

	restore := func() error { return nil }
	if config.Raw && !config.LineMode {
		inFile, inOK := input.(*os.File)
		outFile, outOK := output.(*os.File)
		if !inOK || !outOK {
			return fmt.Errorf("%w: input/output non sono file console", ErrRawModeUnavailable)
		}
		nextRestore, err := EnterConsoleRaw(inFile, outFile)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrRawModeUnavailable, err)
		}
		restore = nextRestore
	}
	defer func() {
		_ = restore()
		fmt.Fprint(output, "\x1b[?25h\r\n")
	}()

	RenderSnapshot(output, term, config.Footer)
	_ = term.DrainOutput()
	buf := make([]byte, 1)
	for {
		n, err := input.Read(buf)
		if n > 0 {
			keepRunning, handleErr := config.Handler.HandleByte(term, buf[0])
			if handleErr != nil {
				return handleErr
			}
			if !keepRunning {
				return nil
			}
			if err := WriteDelta(output, term); err != nil {
				return err
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func FeedBytes(handler Handler, term *terminal.Terminal, data []byte) error {
	for _, value := range data {
		keepRunning, err := handler.HandleByte(term, value)
		if err != nil {
			return err
		}
		if !keepRunning {
			return nil
		}
	}
	return nil
}

func WriteDelta(output io.Writer, term *terminal.Terminal) error {
	data := term.DrainOutput()
	if len(data) == 0 {
		return nil
	}
	_, err := output.Write(data)
	return err
}

func RenderSnapshot(output io.Writer, term *terminal.Terminal, footer string) {
	if footer == "" {
		footer = DefaultFooter
	}
	snapshot := term.Snapshot()
	fmt.Fprint(output, "\x1b[?25l\x1b[H")
	for row, text := range snapshot.Rows {
		if row > 0 {
			fmt.Fprint(output, "\r\n")
		}
		fmt.Fprint(output, text)
	}
	fmt.Fprintf(output, "\r\n\x1b[2K%s", footer)
	fmt.Fprintf(output, "\x1b[%d;%dH\x1b[?25h", snapshot.CursorRow+1, snapshot.CursorCol+1)
}
