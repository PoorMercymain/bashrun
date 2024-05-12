package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	str := "abc_tmp.txt"
	SetLogFile(str)

	n, err := Logger().Write([]byte(str))
	require.NoError(t, err)
	require.Equal(t, len(str), n)
}
