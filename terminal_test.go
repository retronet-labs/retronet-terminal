package terminal

import (
	"io"
	"strings"
	"sync"
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

func TestSnapshotDrainAndResize(t *testing.T) {
	term := New(Config{Width: 6, Height: 2, ANSI: true})
	term.QueueInputString("AB")
	_, _ = term.Write([]byte("HI"))
	snap := term.Snapshot()
	if snap.Width != 6 || snap.Height != 2 || snap.CursorRow != 0 || snap.CursorCol != 2 {
		t.Fatalf("snapshot=%+v", snap)
	}
	if snap.PendingInput != 2 || snap.OutputBytes != 2 {
		t.Fatalf("snapshot pending/output=%d/%d", snap.PendingInput, snap.OutputBytes)
	}
	if snap.Rows[0] != "HI    " {
		t.Fatalf("row=%q", snap.Rows[0])
	}
	if got := string(term.DrainOutput()); got != "HI" {
		t.Fatalf("drain=%q", got)
	}
	if term.Snapshot().OutputBytes != 0 {
		t.Fatalf("output not drained")
	}
	term.Resize(4, 3)
	resized := term.Snapshot()
	if resized.Width != 4 || resized.Height != 3 || resized.Rows[0] != "HI  " {
		t.Fatalf("resized=%+v", resized)
	}
}

func TestANSICursorMovementAndEraseModes(t *testing.T) {
	term := New(Config{Width: 8, Height: 3, ANSI: true})
	_, _ = term.Write([]byte("ABCDEFG\r\n1234567\x1b[1A\x1b[3D!"))
	if got := term.ScreenString(); got != "ABCD!FG\n1234567\n" {
		t.Fatalf("cursor screen=%q", got)
	}
	_, _ = term.Write([]byte("\x1b[2;4H\x1b[1K"))
	if got := term.Snapshot().Rows[1]; got != "    567 " {
		t.Fatalf("erase line left=%q", got)
	}
	_, _ = term.Write([]byte("\x1b[1J"))
	if got := term.Snapshot().Rows[0]; got != "        " {
		t.Fatalf("erase display start=%q", got)
	}
	_, _ = term.Write([]byte("\x1b[2JX"))
	if got := term.ScreenString(); got != "X\n\n" {
		t.Fatalf("erase all=%q", got)
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

func TestIncompleteAndUnknownANSISequencesDoNotPanic(t *testing.T) {
	term := New(Config{Width: 8, Height: 2, ANSI: true})
	_, _ = term.Write([]byte("A\x1b[12"))
	if got := term.ScreenString(); got != "A\n" {
		t.Fatalf("incomplete screen=%q", got)
	}
	_, _ = term.Write([]byte("zB\x1b[?25lC"))
	if got := term.OutputString(); !strings.Contains(got, "\x1b[12zB\x1b[?25lC") {
		t.Fatalf("raw output lost escapes: %q", got)
	}
	if got := term.ScreenString(); got != "ABC\n" {
		t.Fatalf("unknown screen=%q", got)
	}
}

func TestConcurrentAccess(t *testing.T) {
	term := New(Config{Width: 40, Height: 5, ANSI: true})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				term.QueueInputString("x")
				_, _ = term.ReadByte()
				_ = term.WriteByte('A')
				_ = term.Snapshot()
			}
		}()
	}
	wg.Wait()
	if term.Snapshot().Width != 40 {
		t.Fatalf("snapshot corrupted")
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
