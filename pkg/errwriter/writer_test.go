package errwriter

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	appErrors "github.com/PoorMercymain/bashrun/errors"
)

type CustomResponseWriter struct {
	StatusCode int
	Body       JSONError
}

func (w *CustomResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	err := json.Unmarshal(b, &w.Body)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func TestWriteHTTPError(t *testing.T) {
	w := &CustomResponseWriter{}

	testError := errors.New("Test Error")

	WriteHTTPError(w, testError, http.StatusInternalServerError, "Prefix")

	require.Equal(t, http.StatusInternalServerError, w.StatusCode)
	require.Equal(t, appErrors.ErrSomethingWentWrong.Error(), w.Body.Err)

	WriteHTTPError(w, testError, http.StatusBadRequest, "Prefix")

	require.Equal(t, http.StatusBadRequest, w.StatusCode)
	require.Equal(t, testError.Error(), w.Body.Err)
}
