package category

import (
	"context"

	categorydomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/category"
)

type Repository interface {
	List(ctx context.Context) ([]categorydomain.Category, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]categorydomain.Category, error) {
	return s.repo.List(ctx)
}
