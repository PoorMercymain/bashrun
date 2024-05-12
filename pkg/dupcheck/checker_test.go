package dupcheck

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckDuplicatesInJSON(t *testing.T) {
	d := json.NewDecoder(strings.NewReader("{\"test\":100}"))
	err := CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("[{\"test\":100}]"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"test\":200}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader(""))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("[]"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("[1, 2]"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("[1, 2, 1]"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("[\"1\", \"2\"]"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("[\"1\", \"2\", \"1\"]"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"t\":[1, 2]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"t\":[1, 2, 1]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":\"100\", \"t\":[1, 2]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":\"100\", \"t\":[1, 2, 1]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":\"100\", \"t\":[\"1\", \"2\"]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":\"100\", \"t\":[\"1\", \"2\", \"1\"]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"t\":[\"1\", \"2\"]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"t\":[\"1\", \"2\", \"1\"]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"1\":100, \"t\":[\"1\", \"2\"]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"test\":11, \"t\":[\"1\", \"2\"]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"t\":11, \"test\":[\"1\", \"2\"]}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)
}
