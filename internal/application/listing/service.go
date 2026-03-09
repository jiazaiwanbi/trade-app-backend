package listing

import (
	"context"

	listingdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/listing"
)

type Repository interface {
	Create(ctx context.Context, listing listingdomain.Listing) (listingdomain.Listing, error)
	GetByID(ctx context.Context, id int64) (listingdomain.Listing, error)
	Update(ctx context.Context, listing listingdomain.Listing) (listingdomain.Listing, error)
	List(ctx context.Context, filter ListFilter) ([]listingdomain.Listing, int, error)
	ListBySeller(ctx context.Context, sellerID int64, page int, pageSize int) ([]listingdomain.Listing, int, error)
}

type ListFilter struct {
	Keyword    string
	CategoryID *int64
	Page       int
	PageSize   int
}

type CreateInput struct {
	SellerID    int64
	CategoryID  *int64
	Title       string
	Description string
	PriceCents  int64
	Publish     bool
}

type UpdateInput struct {
	ListingID   int64
	SellerID    int64
	CategoryID  *int64
	Title       string
	Description string
	PriceCents  int64
	Status      *listingdomain.Status
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (listingdomain.Listing, error) {
	listing, err := listingdomain.New(input.SellerID, input.CategoryID, input.Title, input.Description, input.PriceCents)
	if err != nil {
		return listingdomain.Listing{}, err
	}
	if input.Publish {
		listing, err = listing.Publish()
		if err != nil {
			return listingdomain.Listing{}, err
		}
	}
	return s.repo.Create(ctx, listing)
}

func (s *Service) Get(ctx context.Context, listingID int64) (listingdomain.Listing, error) {
	return s.repo.GetByID(ctx, listingID)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]listingdomain.Listing, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) ListMine(ctx context.Context, sellerID int64, page int, pageSize int) ([]listingdomain.Listing, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListBySeller(ctx, sellerID, page, pageSize)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (listingdomain.Listing, error) {
	current, err := s.repo.GetByID(ctx, input.ListingID)
	if err != nil {
		return listingdomain.Listing{}, err
	}
	if current.SellerID != input.SellerID {
		return listingdomain.Listing{}, listingdomain.ErrForbidden
	}

	updated, err := current.Update(input.CategoryID, input.Title, input.Description, input.PriceCents)
	if err != nil {
		return listingdomain.Listing{}, err
	}
	updated.Status = current.Status

	if input.Status != nil && *input.Status != current.Status {
		switch *input.Status {
		case listingdomain.StatusDraft:
			updated.Status = current.Status
		case listingdomain.StatusPublished:
			updated, err = updated.Publish()
		case listingdomain.StatusArchived:
			updated, err = updated.Archive()
		case listingdomain.StatusSold:
			updated, err = updated.MarkSold()
		default:
			err = listingdomain.ErrInvalidStatusTransition
		}
		if err != nil {
			return listingdomain.Listing{}, err
		}
	}

	return s.repo.Update(ctx, updated)
}
