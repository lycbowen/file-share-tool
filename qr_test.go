package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintAccessQRCode(t *testing.T) {
	var buf bytes.Buffer
	url := "http://192.168.1.3:8000"

	if err := printAccessQRCode(&buf, url); err != nil {
		t.Fatalf("printAccessQRCode() error = %v", err)
	}

	got := buf.String()
	for _, want := range []string{"Scan to open:", url, ansiBlack, ansiWhite, ansiReset} {
		if !strings.Contains(got, want) {
			t.Fatalf("printAccessQRCode() output missing %q", want)
		}
	}
}

func TestWriteTerminalQRCode(t *testing.T) {
	var buf bytes.Buffer
	writeTerminalQRCode(&buf, [][]bool{
		{true, false},
		{false, true},
	})

	got := buf.String()
	if strings.Count(got, "\n") != 2 {
		t.Fatalf("writeTerminalQRCode() wrote %q, want two lines", got)
	}
	if !strings.Contains(got, ansiBlack) || !strings.Contains(got, ansiWhite) {
		t.Fatalf("writeTerminalQRCode() output missing terminal colors: %q", got)
	}
}
