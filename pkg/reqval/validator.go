package reqval

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	appErrors "github.com/PoorMercymain/bashrun/errors"
	"github.com/PoorMercymain/bashrun/pkg/dupcheck"
	"github.com/PoorMercymain/bashrun/pkg/mimecheck"
)

func ValidateJSONRequest(r *http.Request) error {
	if !mimecheck.IsJSONContentTypeCorrect(r) {
		return appErrors.ErrWrongMIME
	}

	bytesToCheck, err := io.ReadAll(r.Body)
	if err != nil {
		return appErrors.ErrSomethingWentWrong
	}

	reader := bytes.NewReader(bytes.Clone(bytesToCheck))

	err = dupcheck.CheckDuplicatesInJSON(json.NewDecoder(reader), nil)
	if err != nil {
		return appErrors.ErrWrongJSON
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bytesToCheck))

	return nil
}
