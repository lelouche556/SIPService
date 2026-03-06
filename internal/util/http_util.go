package util

import (
	"encoding/json"
	"errors"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteErr(w http.ResponseWriter, code int, err error) {
	WriteJSON(w, code, map[string]string{"error": err.Error()})
}

func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict), errors.Is(err, ErrExecutionInProgress):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidState), errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
