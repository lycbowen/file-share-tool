package main

import (
	"fmt"
	"io"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	ansiReset = "\x1b[0m"
	ansiBlack = "\x1b[40m"
	ansiWhite = "\x1b[47m"
)

func printAccessQRCode(w io.Writer, url string) error {
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return err
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Scan to open:")
	fmt.Fprintln(w, url)
	fmt.Fprintln(w)
	writeTerminalQRCode(w, qr.Bitmap())
	fmt.Fprintln(w)
	return nil
}

func writeTerminalQRCode(w io.Writer, bitmap [][]bool) {
	for _, row := range bitmap {
		var line strings.Builder
		lastColor := ""
		for _, dark := range row {
			color := ansiWhite
			if dark {
				color = ansiBlack
			}
			if color != lastColor {
				line.WriteString(color)
				lastColor = color
			}
			line.WriteString("  ")
		}
		line.WriteString(ansiReset)
		fmt.Fprintln(w, line.String())
	}
}
