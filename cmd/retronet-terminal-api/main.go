// Comando retronet-terminal-api: client terminale per retronet-api.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/retronet-labs/retronet-terminal/internal/wsclient"
	"github.com/retronet-labs/retronet-terminal/live"
)

const defaultAPIURL = "http://127.0.0.1:8080"

type runConfig struct {
	apiURL         string
	sessionID      string
	websocketURL   string
	lineMode       bool
	script         string
	scriptWait     time.Duration
	connectTimeout time.Duration
}

type socketMessage struct {
	Type   string `json:"type"`
	Data   string `json:"data,omitempty"`
	State  string `json:"state,omitempty"`
	Closed bool   `json:"closed,omitempty"`
	Error  string `json:"error,omitempty"`
}

type inputMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, true))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, raw bool) int {
	cfg, err := parseFlags(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "errore: %v\n", err)
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.connectTimeout)
	defer cancel()

	wsURL := cfg.websocketURL
	if wsURL == "" {
		sessionID := cfg.sessionID
		if sessionID == "" {
			var err error
			sessionID, err = createSession(ctx, cfg.apiURL)
			if err != nil {
				fmt.Fprintf(stderr, "errore creazione sessione: %v\n", err)
				return 1
			}
			fmt.Fprintf(stderr, "sessione API creata: %s\n", sessionID)
		}
		var err error
		wsURL, err = sessionWebSocketURL(cfg.apiURL, sessionID)
		if err != nil {
			fmt.Fprintf(stderr, "errore URL websocket: %v\n", err)
			return 2
		}
	}

	conn, err := wsclient.Dial(ctx, wsURL)
	if err != nil {
		fmt.Fprintf(stderr, "errore connessione websocket: %v\n", err)
		return 1
	}
	defer conn.Close()

	done := make(chan error, 1)
	go readWebSocket(conn, stdout, stderr, done)

	if cfg.script != "" {
		if err := sendInput(conn, cfg.script); err != nil {
			fmt.Fprintf(stderr, "errore invio script: %v\n", err)
			return 1
		}
		select {
		case err := <-done:
			if err != nil {
				fmt.Fprintf(stderr, "errore websocket: %v\n", err)
				return 1
			}
		case <-time.After(cfg.scriptWait):
		}
		return 0
	}

	if cfg.lineMode || !raw {
		if err := runLineMode(conn, stdin, done); err != nil {
			fmt.Fprintf(stderr, "errore terminale API: %v\n", err)
			return 1
		}
		return 0
	}

	if err := runRawMode(conn, stdin, stdout, done); err != nil {
		fmt.Fprintf(stderr, "errore terminale API: %v\n", err)
		fmt.Fprintln(stderr, "Riprova con -line se la console non supporta raw mode.")
		return 1
	}
	return 0
}

func parseFlags(args []string, stderr io.Writer) (runConfig, error) {
	fs := flag.NewFlagSet("retronet-terminal-api", flag.ContinueOnError)
	if stderr == nil {
		stderr = io.Discard
	}
	fs.SetOutput(stderr)
	cfg := runConfig{
		apiURL:         defaultAPIURL,
		scriptWait:     500 * time.Millisecond,
		connectTimeout: 5 * time.Second,
	}
	fs.StringVar(&cfg.apiURL, "api", cfg.apiURL, "base URL HTTP di retronet-api")
	fs.StringVar(&cfg.sessionID, "session", "", "ID sessione esistente; se vuoto viene creata una sessione")
	fs.StringVar(&cfg.websocketURL, "url", "", "URL websocket completo; se impostato ignora -api e -session")
	fs.BoolVar(&cfg.lineMode, "line", false, "usa input a righe invece del raw mode")
	fs.StringVar(&cfg.script, "script", "", "byte da inviare subito e poi uscire dopo -script-wait")
	fs.DurationVar(&cfg.scriptWait, "script-wait", cfg.scriptWait, "tempo di attesa dopo -script")
	fs.DurationVar(&cfg.connectTimeout, "connect-timeout", cfg.connectTimeout, "timeout connessione e creazione sessione")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if cfg.websocketURL != "" {
		if _, err := url.Parse(cfg.websocketURL); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func createSession(ctx context.Context, apiURL string) (string, error) {
	endpoint, err := joinURL(apiURL, "/sessions")
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(nil))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("POST /sessions HTTP %d", resp.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", err
	}
	if strings.TrimSpace(created.ID) == "" {
		return "", errors.New("session id vuoto")
	}
	return created.ID, nil
}

func sessionWebSocketURL(apiURL string, sessionID string) (string, error) {
	path := "/sessions/" + sessionID + "/ws"
	value, err := joinURL(apiURL, path)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("schema API non valido: %q", u.Scheme)
	}
	return u.String(), nil
}

func joinURL(baseURL string, path string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("URL API non valido: %q", baseURL)
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func readWebSocket(conn *wsclient.Conn, stdout io.Writer, stderr io.Writer, done chan<- error) {
	for {
		text, err := conn.ReadText()
		if err != nil {
			if errors.Is(err, wsclient.ErrCloseFrame) || errors.Is(err, io.EOF) {
				done <- nil
				return
			}
			done <- err
			return
		}
		var msg socketMessage
		if err := json.Unmarshal([]byte(text), &msg); err != nil {
			fmt.Fprintf(stderr, "messaggio websocket non JSON: %s\n", text)
			continue
		}
		switch msg.Type {
		case "output":
			if msg.Data != "" {
				_, _ = io.WriteString(stdout, msg.Data)
			}
		case "error":
			if msg.Error != "" {
				fmt.Fprintf(stderr, "errore API: %s\n", msg.Error)
			}
		}
		if msg.Closed || msg.State == "closed" {
			done <- nil
			return
		}
	}
}

func runLineMode(conn *wsclient.Conn, stdin io.Reader, done <-chan error) error {
	inputErr := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(stdin)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				if sendErr := sendInput(conn, normalizeLine(line)); sendErr != nil {
					inputErr <- sendErr
					return
				}
			}
			if errors.Is(err, io.EOF) {
				inputErr <- nil
				return
			}
			if err != nil {
				inputErr <- err
				return
			}
		}
	}()
	select {
	case err := <-done:
		return err
	case err := <-inputErr:
		return err
	}
}

func runRawMode(conn *wsclient.Conn, stdin io.Reader, stdout io.Writer, done <-chan error) error {
	inputFile, inputOK := stdin.(*os.File)
	outputFile, outputOK := stdout.(*os.File)
	if !inputOK || !outputOK {
		return errors.New("input/output non sono file console")
	}
	restore, err := live.EnterConsoleRaw(inputFile, outputFile)
	if err != nil {
		return err
	}
	defer restore()

	inputErr := make(chan error, 1)
	go func() {
		buf := make([]byte, 64)
		for {
			n, err := stdin.Read(buf)
			if n > 0 {
				if sendErr := sendInput(conn, string(buf[:n])); sendErr != nil {
					inputErr <- sendErr
					return
				}
			}
			if errors.Is(err, io.EOF) {
				inputErr <- nil
				return
			}
			if err != nil {
				inputErr <- err
				return
			}
		}
	}()
	select {
	case err := <-done:
		return err
	case err := <-inputErr:
		return err
	}
}

func sendInput(conn *wsclient.Conn, data string) error {
	return conn.SendJSON(inputMessage{Type: "input", Data: data})
}

func normalizeLine(line string) string {
	line = strings.TrimRight(line, "\r\n")
	return line + "\r"
}
