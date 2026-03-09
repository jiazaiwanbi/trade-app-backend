package auth

import (
	"context"
	"fmt"
	"strings"

	authdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/auth"
	userdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/user"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/auth"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(context.Context, userdomain.User) (userdomain.User, error)
	FindByEmail(context.Context, string) (userdomain.User, error)
}

type Service struct {
	repo         UserRepository
	tokenManager *auth.TokenManager
}

type RegisterInput struct {
	Email    string
	Password string
	Nickname string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	AccessToken string          `json:"access_token"`
	User        userdomain.User `json:"user"`
}

func NewService(repo UserRepository, tokenManager *auth.TokenManager) *Service {
	return &Service{repo: repo, tokenManager: tokenManager}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}

	user, err := userdomain.New(input.Email, passwordHash, input.Nickname)
	if err != nil {
		return AuthResult{}, err
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}

	token, err := s.tokenManager.Generate(created.ID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResult{AccessToken: token, User: created}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	user, err := s.repo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		return AuthResult{}, authdomain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthResult{}, authdomain.ErrInvalidCredentials
	}

	token, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResult{AccessToken: token, User: user}, nil
}

func hashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}
