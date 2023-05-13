package domains

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDomainLoader(t *testing.T) {
	dl, err := New()
	assert.NoError(t, err, "failed to initialize URLLoader")

	// Check if file was created in tmp directory
	_, err = os.Stat(filepath.Join(os.TempDir(), tmpFile))
	assert.NoError(t, err, "file does not exist in tmp directory")

	// Check if we can read a domain from the file
	domain, err := dl.NextURL()
	assert.NoError(t, err, "failed to read next url")
	assert.NotEmpty(t, domain, "domain should not be empty")

	// Check if EOF error is returned when we reach end of the file
	var lastErr error
	for lastErr == nil {
		_, lastErr = dl.NextURL()
	}
	assert.True(t, errors.Is(lastErr, io.EOF), "EOF error should be returned when reaching end of the file")

	// Check if we can close the file
	err = dl.Close()
	assert.NoError(t, err, "failed to close file")
}
