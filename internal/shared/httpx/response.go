package httpx

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorDetail struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason"`
}

type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Details    []ErrorDetail
}

func (e AppError) Error() string {
	return e.Message
}

func WriteOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, SuccessResponse{
		Code:    "OK",
		Message: "success",
		Data:    data,
	})
}

func WriteError(w http.ResponseWriter, _ *http.Request, err AppError) {
	writeJSON(w, err.StatusCode, ErrorResponse{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	})
}

func TimeoutResponse() string {
	response := ErrorResponse{
		Code:    "TIMEOUT",
		Message: "request timed out",
	}

	encoded, _ := json.Marshal(response)
	return string(encoded)
}

func NewInternalError() AppError {
	return AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    "internal server error",
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
