package main

import "testing"

func TestSessionWebSocketURL(t *testing.T) {
	tests := []struct {
		name      string
		apiURL    string
		sessionID string
		want      string
	}{
		{
			name:      "http default path",
			apiURL:    "http://127.0.0.1:8080",
			sessionID: "abc",
			want:      "ws://127.0.0.1:8080/sessions/abc/ws",
		},
		{
			name:      "https nested path",
			apiURL:    "https://example.test/api/",
			sessionID: "a b",
			want:      "wss://example.test/api/sessions/a%20b/ws",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sessionWebSocketURL(tt.apiURL, tt.sessionID)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("url=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeLine(t *testing.T) {
	tests := map[string]string{
		"DIR\n":      "DIR\r",
		"HELP\r\n":   "HELP\r",
		"TYPE X.TXT": "TYPE X.TXT\r",
	}
	for input, want := range tests {
		if got := normalizeLine(input); got != want {
			t.Fatalf("normalizeLine(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestParseFlags(t *testing.T) {
	cfg, err := parseFlags([]string{"-api", "http://localhost:8080", "-session", "s1", "-line"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.apiURL != "http://localhost:8080" || cfg.sessionID != "s1" || !cfg.lineMode {
		t.Fatalf("cfg=%+v", cfg)
	}
}
