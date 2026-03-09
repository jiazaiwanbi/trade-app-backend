package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	authapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/auth"
	userdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/user"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
	"github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/dto"
)

type AuthHandler struct {
	service *authapp.Service
}

func NewAuthHandler(service *authapp.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid request body"})
		return
	}

	result, err := h.service.Register(r.Context(), authapp.RegisterInput(req))
	if err != nil {
		handleAuthError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, dto.AuthResponse{AccessToken: result.AccessToken, User: toUserResponse(result.User)})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid request body"})
		return
	}

	result, err := h.service.Login(r.Context(), authapp.LoginInput(req))
	if err != nil {
		handleAuthError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, dto.AuthResponse{AccessToken: result.AccessToken, User: toUserResponse(result.User)})
}

func handleAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, userdomain.ErrInvalidEmail), errors.Is(err, userdomain.ErrInvalidNickname):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: err.Error()})
	case err.Error() == "password must be at least 6 characters":
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: err.Error()})
	case err.Error() == "email already exists":
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusConflict, Code: "CONFLICT", Message: err.Error()})
	default:
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: err.Error()})
	}
}

func toUserResponse(user userdomain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
