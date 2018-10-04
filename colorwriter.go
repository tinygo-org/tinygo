package main

import (
	"io"
)

// ANSI escape codes for terminal colors.
const (
	TermColorReset  = "\x1b[0m"
	TermColorYellow = "\x1b[93m"
)

// ColorWriter wraps an io.Writer but adds a prefix and a terminal color.
type ColorWriter struct {
	Out    io.Writer
	Color  string
	Prefix string
	line   []byte
}

// Write implements io.Writer, but with an added prefix and terminal color.
func (w *ColorWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		if c == '\n' {
			w.line = append(w.line, []byte(TermColorReset)...)
			w.line = append(w.line, '\n')
			// Write this line.
			_, err := w.Out.Write(w.line)
			w.line = w.line[:0]
			w.line = append(w.line, []byte(w.Color+w.Prefix)...)
			if err != nil {
				return 0, err
			}
		} else {
			w.line = append(w.line, c)
		}
	}
	return len(p), nil
}
