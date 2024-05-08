package errwriter

import (
	"encoding/json"
	"net/http"

	"github.com/PoorMercymain/bashrun/pkg/logger"
	appErrors "github.com/PoorMercymain/bashrun/errors"
)

type JSONError struct {
	Err string `json:"error"`
}

func WriteHTTPError(w http.ResponseWriter, err error, statusCode int, prefix string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err != nil {
		if statusCode == http.StatusInternalServerError {
			logger.Logger().Error(prefix, ": ", err.Error())

			err = appErrors.ErrSomethingWentWrong
		}

		err = json.NewEncoder(w).Encode(JSONError{Err: err.Error()})
		if err != nil {
			logger.Logger().Error(prefix, ": ", err.Error())
		}
	}
}
