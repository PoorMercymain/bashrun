package mimecheck

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJSONContentTypeCorrect(t *testing.T) {
	r, err := http.NewRequest("POST", "", bytes.NewReader([]byte("")))
	require.NoError(t, err)

	isCorrect := IsJSONContentTypeCorrect(r)
	require.False(t, isCorrect)

	r.Header.Set("Content-Type", "application/json")
	isCorrect = IsJSONContentTypeCorrect(r)
	require.True(t, isCorrect)

	r.Header.Set("Content-Type", "text/plain")
	isCorrect = IsJSONContentTypeCorrect(r)
	require.False(t, isCorrect)
}
