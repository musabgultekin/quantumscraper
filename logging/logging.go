package logging

import (
	"io"
	"strings"
)

type FilteredWriter struct {
	w io.Writer
}

func (fw *FilteredWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "Unsolicited response received") {
		// Discard unwanted log messages
		return len(p), nil
	}
	return fw.w.Write(p)
}
