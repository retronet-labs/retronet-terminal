package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/retronet-labs/retronet-terminal"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("retronet-terminal", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("input", "", "testo da accodare nella coda input")
	write := fs.String("write", "", "testo da scrivere sul terminale")
	screen := fs.Bool("screen", false, "stampa lo schermo testuale invece dell'output raw")
	demo := fs.Bool("demo", false, "mostra una demo ANSI locale")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	term := terminal.New(terminal.Config{ANSI: true})
	if *input != "" {
		term.QueueInputString(*input)
	}
	if *demo {
		_, _ = term.Write([]byte("RetroNet Terminal\r\n\x1b[3;1HREADY>\x1b[4;1HANSI OK"))
	}
	if *write != "" {
		_, _ = term.Write([]byte(*write))
	}
	if *screen {
		fmt.Fprintln(stdout, term.ScreenString())
		return 0
	}
	fmt.Fprint(stdout, term.OutputString())
	return 0
}
