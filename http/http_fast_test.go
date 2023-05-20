package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBrotli(t *testing.T) {
	resp, _, err := GetFast("https://httpbin.org/brotli")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
