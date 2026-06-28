package wsclient

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDialSendAndReadText(t *testing.T) {
	received := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			t.Errorf("Upgrade=%q", r.Header.Get("Upgrade"))
		}
		h, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("hijack non supportato")
		}
		conn, rw, err := h.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		key := r.Header.Get("Sec-WebSocket-Key")
		_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
		_, _ = rw.WriteString("Upgrade: websocket\r\n")
		_, _ = rw.WriteString("Connection: Upgrade\r\n")
		_, _ = rw.WriteString("Sec-WebSocket-Accept: " + AcceptKey(key) + "\r\n\r\n")
		if err := rw.Flush(); err != nil {
			t.Fatal(err)
		}
		opcode, payload, masked, err := readServerFrame(rw.Reader)
		if err != nil {
			t.Fatal(err)
		}
		if opcode != 0x1 {
			t.Fatalf("opcode=%d", opcode)
		}
		if !masked {
			t.Fatal("frame client non mascherato")
		}
		received <- string(payload)
		if err := writeServerText(conn, `{"type":"output","data":"OK"}`); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, err := Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.SendText(`{"type":"input","data":"A"}`); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != `{"type":"input","data":"A"}` {
		t.Fatalf("received=%q", got)
	}
	text, err := conn.ReadText()
	if err != nil {
		t.Fatal(err)
	}
	if text != `{"type":"output","data":"OK"}` {
		t.Fatalf("text=%q", text)
	}
}

func TestAcceptKey(t *testing.T) {
	const key = "dGhlIHNhbXBsZSBub25jZQ=="
	const want = "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	if got := AcceptKey(key); got != want {
		t.Fatalf("AcceptKey=%q, want %q", got, want)
	}
}

func readServerFrame(r *bufio.Reader) (byte, []byte, bool, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, false, err
	}
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7F)
	switch length {
	case 126:
		var b [2]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, nil, false, err
		}
		length = uint64(binary.BigEndian.Uint16(b[:]))
	case 127:
		var b [8]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, nil, false, err
		}
		length = binary.BigEndian.Uint64(b[:])
	}
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(r, mask[:]); err != nil {
			return 0, nil, false, err
		}
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, false, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	return opcode, payload, masked, nil
}

func writeServerText(conn net.Conn, payload string) error {
	data := []byte(payload)
	header := []byte{0x81}
	if len(data) < 126 {
		header = append(header, byte(len(data)))
	} else {
		header = append(header, 126, byte(len(data)>>8), byte(len(data)))
	}
	if _, err := conn.Write(header); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}
