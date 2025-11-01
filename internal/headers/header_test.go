package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host:   localhost:42069   \r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 28, n) // length of "Host:   localhost:42069   \r\n"
		assert.False(t, done)
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\nSome body data here")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.True(t, done)
		assert.Empty(t, headers) // No headers should be added
	})

	t.Run("No CRLF - need more data", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Empty(t, headers) // No headers should be added
	})

	t.Run("Header keys converted to lowercase", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Content-Type: application/json\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "application/json", headers["content-type"]) // lowercase key
		assert.Equal(t, 32, n)                                       // length of "Content-Type: application/json\r\n"
		assert.False(t, done)
	})

	t.Run("Invalid character in header key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("H©st: localhost:42069\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Empty(t, headers)
	})

	t.Run("Multiple values for same header", func(t *testing.T) {
		headers := NewHeaders()

		// First Accept header
		data := []byte("Accept: text/html\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "text/html", headers["accept"])
		assert.Equal(t, 19, n) // length of "Accept: text/html\r\n"
		assert.False(t, done)

		// Second Accept header - should append
		data = []byte("Accept: application/json\r\n")
		n, done, err = headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "text/html, application/json", headers["accept"])
		assert.Equal(t, 26, n) // length of "Accept: application/json\r\n"
		assert.False(t, done)

		// Third Accept header - should append again
		data = []byte("Accept: application/xml\r\n")
		n, done, err = headers.Parse(data)
		assert.Equal(t, 25, n) // length of "Accept: application/json\r\n"
		require.NoError(t, err)
		assert.Equal(t, "text/html, application/json, application/xml", headers["accept"])
		assert.False(t, done)
	})
}
