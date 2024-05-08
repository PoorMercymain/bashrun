package handler

import (
	"net/http"

	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
	"github.com/PoorMercymain/bashrun/pkg/errwriter"
)

type bashrunHandlers struct {
	srv domain.BashrunService
}

func New(srv domain.BashrunService) *bashrunHandlers {
	return &bashrunHandlers{srv: srv}
}

func (h *bashrunHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	const logPrefix = "handlers.Ping"

	err := h.srv.Ping(r.Context())
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logPrefix)
	}
}
