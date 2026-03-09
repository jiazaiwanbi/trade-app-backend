package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	userapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/user"
	userdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/user"
	platformmiddleware "github.com/jiazaiwanbi/second-hand-platform/internal/platform/http/middleware"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
	"github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/dto"
)

type UserHandler struct {
	service *userapp.Service
}

func NewUserHandler(service *userapp.Service) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "unauthorized"})
		return
	}

	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		handleUserError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, toUserResponse(user))
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "unauthorized"})
		return
	}

	var req dto.UpdateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid request body"})
		return
	}

	updated, err := h.service.UpdateMe(r.Context(), userID, userapp.UpdateProfileInput(req))
	if err != nil {
		handleUserError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, toUserResponse(updated))
}

func handleUserError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, userdomain.ErrUserNotFound):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusNotFound, Code: "NOT_FOUND", Message: err.Error()})
	case errors.Is(err, userdomain.ErrInvalidNickname):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: err.Error()})
	default:
		sharedhttp.WriteError(w, r, sharedhttp.NewInternalError())
	}
}
