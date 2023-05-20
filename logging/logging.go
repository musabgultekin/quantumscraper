package logging

import (
	"io"
	"strings"
)

type FilteredWriter struct {
	io.Writer
}

func (fw *FilteredWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "Unsolicited response received on idle HTTP channel starting with") {
		// Discard unwanted log messages
		return len(p), nil
	}
	return fw.Writer.Write(p)
}
