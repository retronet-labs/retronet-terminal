// Package terminal implementa un terminale testuale retro, indipendente dalle
// CPU RetroNet e riusabile da CLI, CP/M-like, BBS e futuri websocket.
package terminal

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"sync"
)

const (
	DefaultWidth  = 80
	DefaultHeight = 24
)

// Config controlla il comportamento del terminale.
type Config struct {
	Width  int
	Height int
	ANSI   bool
}

// Snapshot e' una vista immutabile dello stato utile a UI e websocket.
// Rows contiene righe a larghezza fissa; CursorRow/CursorCol sono zero-based.
type Snapshot struct {
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Rows         []string `json:"rows"`
	CursorRow    int      `json:"cursor_row"`
	CursorCol    int      `json:"cursor_col"`
	PendingInput int      `json:"pending_input"`
	OutputBytes  int      `json:"output_bytes"`
}

// Terminal mantiene una coda input, un buffer output raw e uno schermo testuale.
type Terminal struct {
	mu     sync.Mutex
	config Config
	input  []byte
	output bytes.Buffer
	screen [][]byte
	row    int
	col    int
	esc    []byte
}

// New crea un terminale vuoto. Width/Height pari a zero usano 80x24.
func New(config Config) *Terminal {
	if config.Width <= 0 {
		config.Width = DefaultWidth
	}
	if config.Height <= 0 {
		config.Height = DefaultHeight
	}
	t := &Terminal{config: config}
	t.clearScreen()
	return t
}

// NewMemory crea un terminale con input iniziale accodato.
func NewMemory(input []byte) *Terminal {
	t := New(Config{ANSI: true})
	t.QueueInput(input)
	return t
}

// QueueInput accoda una copia dei byte che verranno consumati da ReadByte.
func (t *Terminal) QueueInput(data []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.input = append(t.input, data...)
}

// QueueInputString accoda una stringa senza conversioni di encoding.
func (t *Terminal) QueueInputString(value string) {
	t.QueueInput([]byte(value))
}

// PendingInput restituisce il numero di byte disponibili.
func (t *Terminal) PendingInput() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.input)
}

// Status indica se almeno un byte di input e' pronto.
func (t *Terminal) Status() bool {
	return t.PendingInput() > 0
}

// ReadByte legge un byte dalla coda input o io.EOF se la coda e' vuota.
func (t *Terminal) ReadByte() (byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.input) == 0 {
		return 0, io.EOF
	}
	value := t.input[0]
	t.input = t.input[1:]
	return value, nil
}

// Read implementa io.Reader sopra la coda input.
func (t *Terminal) Read(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.input) == 0 {
		return 0, io.EOF
	}
	n := copy(p, t.input)
	t.input = t.input[n:]
	return n, nil
}

// WriteByte scrive un byte nel buffer raw e aggiorna lo schermo testuale.
func (t *Terminal) WriteByte(value byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.writeByteLocked(value)
	return nil
}

// Write implementa io.Writer.
func (t *Terminal) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, value := range p {
		t.writeByteLocked(value)
	}
	return len(p), nil
}

// OutputBytes restituisce una copia dell'output raw.
func (t *Terminal) OutputBytes() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]byte(nil), t.output.Bytes()...)
}

// OutputString restituisce l'output raw come stringa.
func (t *Terminal) OutputString() string {
	return string(t.OutputBytes())
}

// DrainOutput restituisce e svuota il buffer raw.
func (t *Terminal) DrainOutput() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	data := append([]byte(nil), t.output.Bytes()...)
	t.output.Reset()
	return data
}

// Snapshot restituisce una vista coerente dello stato corrente.
func (t *Terminal) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	rows := make([]string, len(t.screen))
	for i, row := range t.screen {
		rows[i] = string(row)
	}
	return Snapshot{
		Width:        t.config.Width,
		Height:       t.config.Height,
		Rows:         rows,
		CursorRow:    t.row,
		CursorCol:    t.col,
		PendingInput: len(t.input),
		OutputBytes:  t.output.Len(),
	}
}

// ScreenString restituisce il contenuto visibile, senza spazi finali di riga.
func (t *Terminal) ScreenString() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	lines := make([]string, len(t.screen))
	for i, row := range t.screen {
		lines[i] = strings.TrimRight(string(row), " ")
	}
	return strings.Join(lines, "\n")
}

// Resize cambia dimensioni allo schermo preservando il contenuto visibile in
// alto a sinistra e clampando il cursore dentro i nuovi limiti.
func (t *Terminal) Resize(width int, height int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if width <= 0 {
		width = DefaultWidth
	}
	if height <= 0 {
		height = DefaultHeight
	}
	next := make([][]byte, height)
	for i := range next {
		next[i] = blankLine(width)
	}
	rows := minInt(height, len(t.screen))
	for r := 0; r < rows; r++ {
		copy(next[r], t.screen[r])
	}
	t.config.Width = width
	t.config.Height = height
	t.screen = next
	if t.row >= height {
		t.row = height - 1
	}
	if t.col >= width {
		t.col = width - 1
	}
}

// Reset svuota input, output, parser ANSI e schermo.
func (t *Terminal) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.input = nil
	t.output.Reset()
	t.esc = nil
	t.row = 0
	t.col = 0
	t.clearScreen()
}

func (t *Terminal) writeByteLocked(value byte) {
	t.output.WriteByte(value)
	if t.config.ANSI && len(t.esc) > 0 {
		t.esc = append(t.esc, value)
		if escapeComplete(t.esc) {
			t.handleEscape(string(t.esc))
			t.esc = nil
		}
		return
	}
	if t.config.ANSI && value == 0x1B {
		t.esc = []byte{value}
		return
	}
	t.writeDisplay(value)
}

func (t *Terminal) writeDisplay(value byte) {
	switch value {
	case '\r':
		t.col = 0
	case '\n':
		t.newLine()
	case '\b':
		if t.col > 0 {
			t.col--
		}
	case '\t':
		target := ((t.col / 8) + 1) * 8
		for t.col < target {
			t.putPrintable(' ')
		}
	default:
		if value >= 0x20 && value <= 0x7E {
			t.putPrintable(value)
		}
	}
}

func (t *Terminal) putPrintable(value byte) {
	t.screen[t.row][t.col] = value
	t.col++
	if t.col >= t.config.Width {
		t.col = 0
		t.newLine()
	}
}

func (t *Terminal) newLine() {
	t.row++
	if t.row < t.config.Height {
		return
	}
	copy(t.screen, t.screen[1:])
	t.screen[t.config.Height-1] = blankLine(t.config.Width)
	t.row = t.config.Height - 1
}

func (t *Terminal) clearScreen() {
	t.screen = make([][]byte, t.config.Height)
	for i := range t.screen {
		t.screen[i] = blankLine(t.config.Width)
	}
	t.row = 0
	t.col = 0
}

func (t *Terminal) clearLine() {
	for i := t.col; i < t.config.Width; i++ {
		t.screen[t.row][i] = ' '
	}
}

func (t *Terminal) handleEscape(seq string) {
	if !strings.HasPrefix(seq, "\x1b[") {
		return
	}
	body := seq[2 : len(seq)-1]
	final := seq[len(seq)-1]
	switch final {
	case 'J':
		t.eraseDisplay(parseCSIInt(body, 0))
	case 'K':
		t.eraseLine(parseCSIInt(body, 0))
	case 'H', 'f':
		t.moveCursor(body)
	case 'A':
		t.moveRelative(-parseCSIInt(body, 1), 0)
	case 'B':
		t.moveRelative(parseCSIInt(body, 1), 0)
	case 'C':
		t.moveRelative(0, parseCSIInt(body, 1))
	case 'D':
		t.moveRelative(0, -parseCSIInt(body, 1))
	case 'm':
		// Attributi colore/stile ignorati: il buffer e' testuale.
	}
}

func (t *Terminal) eraseDisplay(mode int) {
	switch mode {
	case 0:
		t.eraseLine(0)
		for r := t.row + 1; r < t.config.Height; r++ {
			t.screen[r] = blankLine(t.config.Width)
		}
	case 1:
		for r := 0; r < t.row; r++ {
			t.screen[r] = blankLine(t.config.Width)
		}
		for c := 0; c <= t.col && c < t.config.Width; c++ {
			t.screen[t.row][c] = ' '
		}
	case 2:
		t.clearScreen()
	}
}

func (t *Terminal) eraseLine(mode int) {
	switch mode {
	case 0:
		for i := t.col; i < t.config.Width; i++ {
			t.screen[t.row][i] = ' '
		}
	case 1:
		for i := 0; i <= t.col && i < t.config.Width; i++ {
			t.screen[t.row][i] = ' '
		}
	case 2:
		t.screen[t.row] = blankLine(t.config.Width)
	}
}

func (t *Terminal) moveCursor(body string) {
	if body == "" {
		t.row = 0
		t.col = 0
		return
	}
	parts := strings.Split(body, ";")
	if len(parts) > 2 {
		return
	}
	row := parseCSIInt(parts[0], 1) - 1
	col := 0
	if len(parts) == 2 {
		col = parseCSIInt(parts[1], 1) - 1
	}
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	if row >= t.config.Height {
		row = t.config.Height - 1
	}
	if col >= t.config.Width {
		col = t.config.Width - 1
	}
	t.row = row
	t.col = col
}

func (t *Terminal) moveRelative(deltaRow int, deltaCol int) {
	row := t.row + deltaRow
	col := t.col + deltaCol
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	if row >= t.config.Height {
		row = t.config.Height - 1
	}
	if col >= t.config.Width {
		col = t.config.Width - 1
	}
	t.row = row
	t.col = col
}

func parseCSIInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func escapeComplete(seq []byte) bool {
	if len(seq) < 2 {
		return false
	}
	if seq[1] != '[' {
		return true
	}
	return len(seq) >= 3 && seq[len(seq)-1] >= 0x40 && seq[len(seq)-1] <= 0x7E
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func blankLine(width int) []byte {
	line := make([]byte, width)
	for i := range line {
		line[i] = ' '
	}
	return line
}
