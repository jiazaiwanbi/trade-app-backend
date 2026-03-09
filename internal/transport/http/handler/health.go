package handler

import (
	"errors"
	"net/http"

	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/database"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
)

type healthResponse struct {
	Status string `json:"status"`
}

type ReadyHandler struct {
	checker database.ReadinessChecker
}

func NewReadyHandler(checker database.ReadinessChecker) *ReadyHandler {
	return &ReadyHandler{checker: checker}
}

func Healthz(w http.ResponseWriter, _ *http.Request) {
	sharedhttp.WriteOK(w, healthResponse{Status: "ok"})
}

func (h *ReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.checker == nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{
			StatusCode: http.StatusServiceUnavailable,
			Code:       "SERVICE_UNAVAILABLE",
			Message:    "readiness dependency is not configured",
		})
		return
	}

	if err := database.CheckReadiness(r.Context(), h.checker); err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{
			StatusCode: http.StatusServiceUnavailable,
			Code:       "SERVICE_UNAVAILABLE",
			Message:    "database is not ready",
			Details:    []sharedhttp.ErrorDetail{{Reason: err.Error()}},
		})
		return
	}

	sharedhttp.WriteOK(w, healthResponse{Status: "ready"})
}

var ErrNotReady = errors.New("database is not ready")
