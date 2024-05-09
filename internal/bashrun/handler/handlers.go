package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	appErrors "github.com/PoorMercymain/bashrun/errors"
	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
	"github.com/PoorMercymain/bashrun/pkg/errwriter"
	"github.com/PoorMercymain/bashrun/pkg/logger"
	"github.com/PoorMercymain/bashrun/pkg/reqval"
)

type bashrunHandlers struct {
	srv domain.BashrunService
}

func New(srv domain.BashrunService) *bashrunHandlers {
	return &bashrunHandlers{srv: srv}
}

func (h *bashrunHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	const logPrefix = "handlers.Ping"
	defer r.Body.Close()

	err := h.srv.Ping(r.Context())
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logPrefix)
	}
}

func (h *bashrunHandlers) CreateCommand(w http.ResponseWriter, r *http.Request) {
	const logPrefix = "handlers.CreateCommand"
	defer r.Body.Close()

	err := reqval.ValidateJSONRequest(r)
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logPrefix)
		return
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var command domain.CommandFromUser
	if err = d.Decode(&command); err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logPrefix)
		return
	}

	if command.Command == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrEmptyCommand, http.StatusBadRequest, logPrefix)
		return
	}

	var commandID domain.ID
	commandID.ID, err = h.srv.CreateCommand(r.Context(), command.Command)
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logPrefix)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	if err = json.NewEncoder(w).Encode(commandID); err != nil {
		logger.Logger().Error(logPrefix, ": ", err.Error())
	}
}

func (h *bashrunHandlers) ListCommands(w http.ResponseWriter, r *http.Request) {
	const logPrefix = "handlers.ListCommands"
	defer r.Body.Close()

	strLimit := r.URL.Query().Get("limit")
	strOffset := r.URL.Query().Get("offset")

	if strLimit == "" {
		strLimit = "15"
	}

	if strOffset == "" {
		strOffset = "0"
	}

	limit, err := strconv.Atoi(strLimit)
	if err != nil || (limit < 1 || limit > 50) {
		errwriter.WriteHTTPError(w, appErrors.ErrWrongLimit, http.StatusBadRequest, logPrefix)
		return
	}

	offset, err := strconv.Atoi(strOffset)
	if err != nil || (offset < 0) {
		errwriter.WriteHTTPError(w, appErrors.ErrWrongOffset, http.StatusBadRequest, logPrefix)
		return
	}

	commands, err := h.srv.ListCommands(r.Context(), limit, offset)
	if err != nil {
		if errors.Is(err, appErrors.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logPrefix)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(commands); err != nil {
		logger.Logger().Error(logPrefix, ": ", err.Error())
	}
}

func (h *bashrunHandlers) StopCommand(w http.ResponseWriter, r *http.Request) {
	const logPrefix = "handlers.StopCommand"
	defer r.Body.Close()

	id, err := strconv.Atoi(r.PathValue("command_id"))
	if r.PathValue("command_id") != "" && (err != nil || id < 1) {
		errwriter.WriteHTTPError(w, appErrors.ErrWrongID, http.StatusBadRequest, logPrefix)
		return
	} else if r.PathValue("command_id") == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrEmptyID, http.StatusBadRequest, logPrefix)
		return
	}

	err = h.srv.StopCommand(r.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.ErrNoRows) {
			errwriter.WriteHTTPError(w, appErrors.ErrCommandNotFound, http.StatusNotFound, logPrefix)
			return
		}

		if errors.Is(err, appErrors.ErrCommandNotRunning) {
			errwriter.WriteHTTPError(w, appErrors.ErrCommandNotRunning, http.StatusBadRequest, logPrefix)
			return
		}

		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logPrefix)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *bashrunHandlers) ReadCommand(w http.ResponseWriter, r *http.Request) {
	const logPrefix = "handlers.ReadCommand"
	defer r.Body.Close()

	id, err := strconv.Atoi(r.PathValue("command_id"))
	if r.PathValue("command_id") != "" && (err != nil || id < 1) {
		errwriter.WriteHTTPError(w, appErrors.ErrWrongID, http.StatusBadRequest, logPrefix)
		return
	} else if r.PathValue("command_id") == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrEmptyID, http.StatusBadRequest, logPrefix)
		return
	}

	command, err := h.srv.ReadCommand(r.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.ErrNoRows) {
			errwriter.WriteHTTPError(w, appErrors.ErrCommandNotFound, http.StatusNotFound, logPrefix)
			return
		}

		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logPrefix)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(command); err != nil {
		logger.Logger().Error(logPrefix, ": ", err.Error())
	}
}
