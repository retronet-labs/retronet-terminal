package terminal

import (
	"io"
	"testing"
)

func TestInputQueue(t *testing.T) {
	term := NewMemory([]byte("AB"))
	if !term.Status() || term.PendingInput() != 2 {
		t.Fatalf("input status=%v pending=%d", term.Status(), term.PendingInput())
	}
	a, err := term.ReadByte()
	if err != nil || a != 'A' {
		t.Fatalf("ReadByte A = 0x%02X err=%v", a, err)
	}
	buf := make([]byte, 2)
	n, err := term.Read(buf)
	if n != 1 || err != nil || string(buf[:n]) != "B" {
		t.Fatalf("Read = n=%d data=%q err=%v", n, string(buf[:n]), err)
	}
	if _, err := term.ReadByte(); err != io.EOF {
		t.Fatalf("empty ReadByte err=%v", err)
	}
}

func TestOutputBufferAndScreen(t *testing.T) {
	term := New(Config{Width: 10, Height: 3})
	if _, err := term.Write([]byte("HI\r\nOK")); err != nil {
		t.Fatal(err)
	}
	if got := term.OutputString(); got != "HI\r\nOK" {
		t.Fatalf("output=%q", got)
	}
	if got := term.ScreenString(); got != "HI\nOK\n" {
		t.Fatalf("screen=%q", got)
	}
}

func TestANSIBase(t *testing.T) {
	term := New(Config{Width: 8, Height: 3, ANSI: true})
	_, _ = term.Write([]byte("ABCDEFGH\r\n1234\x1b[2J\x1b[2;3HOK\x1b[H>"))
	want := ">\n  OK\n"
	if got := term.ScreenString(); got != want {
		t.Fatalf("screen=%q want=%q", got, want)
	}
}

func TestANSIStylesAreIgnoredButBuffered(t *testing.T) {
	term := New(Config{Width: 8, Height: 2, ANSI: true})
	_, _ = term.Write([]byte("\x1b[31mRED\x1b[0m"))
	if got := term.OutputString(); got != "\x1b[31mRED\x1b[0m" {
		t.Fatalf("output=%q", got)
	}
	if got := term.ScreenString(); got != "RED\n" {
		t.Fatalf("screen=%q", got)
	}
}

func TestReset(t *testing.T) {
	term := NewMemory([]byte("A"))
	_, _ = term.Write([]byte("X"))
	term.Reset()
	if term.PendingInput() != 0 || term.OutputString() != "" {
		t.Fatalf("reset input=%d output=%q", term.PendingInput(), term.OutputString())
	}
	if got := term.ScreenString(); got != "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n" {
		t.Fatalf("screen after reset=%q", got)
	}
}
