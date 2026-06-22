package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errBody struct {
	Error APIError `json:"error"`
}

type HTTPError struct {
	Status  int
	Code    string
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

func NewError(status int, code, msg string) *HTTPError {
	return &HTTPError{Status: status, Code: code, Message: msg}
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Err(w http.ResponseWriter, err error) {
	var he *HTTPError
	if errors.As(err, &he) {
		JSON(w, he.Status, errBody{Error: APIError{Code: he.Code, Message: he.Message}})
		return
	}
	JSON(w, http.StatusInternalServerError, errBody{Error: APIError{Code: "internal", Message: err.Error()}})
}

func Decode(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return NewError(http.StatusBadRequest, "invalid_json", "invalid JSON body")
	}
	return nil
}
