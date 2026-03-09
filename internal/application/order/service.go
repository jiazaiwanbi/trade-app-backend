package order

import (
	"context"

	orderdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/order"
)

type Repository interface {
	Create(ctx context.Context, input CreateInput) (orderdomain.Order, error)
	GetByID(ctx context.Context, id int64) (orderdomain.Order, error)
	UpdateStatus(ctx context.Context, order orderdomain.Order) (orderdomain.Order, error)
	ListMine(ctx context.Context, userID int64, page int, pageSize int) ([]orderdomain.Order, int, error)
}

type CreateInput struct {
	ListingID int64
	BuyerID   int64
}

type ActionInput struct {
	OrderID  int64
	ActorID  int64
	Complete bool
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (orderdomain.Order, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Cancel(ctx context.Context, input ActionInput) (orderdomain.Order, error) {
	current, err := s.repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return orderdomain.Order{}, err
	}
	updated, err := current.Cancel(input.ActorID)
	if err != nil {
		return orderdomain.Order{}, err
	}
	return s.repo.UpdateStatus(ctx, updated)
}

func (s *Service) Complete(ctx context.Context, input ActionInput) (orderdomain.Order, error) {
	current, err := s.repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return orderdomain.Order{}, err
	}
	updated, err := current.Complete(input.ActorID)
	if err != nil {
		return orderdomain.Order{}, err
	}
	return s.repo.UpdateStatus(ctx, updated)
}

func (s *Service) ListMine(ctx context.Context, userID int64, page int, pageSize int) ([]orderdomain.Order, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListMine(ctx, userID, page, pageSize)
}
