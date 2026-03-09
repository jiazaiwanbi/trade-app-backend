package user

import (
	"context"

	userdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/user"
)

type Repository interface {
	FindByID(context.Context, int64) (userdomain.User, error)
	UpdateProfile(context.Context, int64, string, string) (userdomain.User, error)
}

type Service struct {
	repo Repository
}

type UpdateProfileInput struct {
	Nickname string
	Bio      string
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMe(ctx context.Context, userID int64) (userdomain.User, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *Service) UpdateMe(ctx context.Context, userID int64, input UpdateProfileInput) (userdomain.User, error) {
	current, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return userdomain.User{}, err
	}
	if err := current.UpdateProfile(input.Nickname, input.Bio); err != nil {
		return userdomain.User{}, err
	}
	return s.repo.UpdateProfile(ctx, userID, current.Nickname, current.Bio)
}
