package http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBrotli(t *testing.T) {
	resp, err := Get("https://httpbin.org/brotli")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
