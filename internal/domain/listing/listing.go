package listing

import (
	"strings"
	"time"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusReserved  Status = "reserved"
	StatusSold      Status = "sold"
	StatusArchived  Status = "archived"
)

type Listing struct {
	ID          int64
	SellerID    int64
	CategoryID  *int64
	Title       string
	Description string
	PriceCents  int64
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func New(sellerID int64, categoryID *int64, title string, description string, priceCents int64) (Listing, error) {
	listing := Listing{
		SellerID:    sellerID,
		CategoryID:  categoryID,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		PriceCents:  priceCents,
		Status:      StatusDraft,
	}

	if err := listing.Validate(); err != nil {
		return Listing{}, err
	}

	return listing, nil
}

func (l Listing) Validate() error {
	if l.SellerID <= 0 {
		return ErrForbidden
	}
	if len(strings.TrimSpace(l.Title)) < 3 || len(strings.TrimSpace(l.Title)) > 120 {
		return ErrInvalidTitle
	}
	if len(strings.TrimSpace(l.Description)) < 5 || len(strings.TrimSpace(l.Description)) > 2000 {
		return ErrInvalidDescription
	}
	if l.PriceCents <= 0 {
		return ErrInvalidPrice
	}
	return nil
}

func (l Listing) Update(categoryID *int64, title string, description string, priceCents int64) (Listing, error) {
	updated := l
	updated.CategoryID = categoryID
	updated.Title = strings.TrimSpace(title)
	updated.Description = strings.TrimSpace(description)
	updated.PriceCents = priceCents

	if err := updated.Validate(); err != nil {
		return Listing{}, err
	}

	return updated, nil
}

func (l Listing) Publish() (Listing, error) {
	if l.Status != StatusDraft && l.Status != StatusArchived {
		return Listing{}, ErrInvalidStatusTransition
	}
	updated := l
	updated.Status = StatusPublished
	return updated, nil
}

func (l Listing) Reserve() (Listing, error) {
	if l.Status != StatusPublished {
		return Listing{}, ErrListingUnavailable
	}
	updated := l
	updated.Status = StatusReserved
	return updated, nil
}

func (l Listing) ReleaseReservation() (Listing, error) {
	if l.Status != StatusReserved {
		return Listing{}, ErrInvalidStatusTransition
	}
	updated := l
	updated.Status = StatusPublished
	return updated, nil
}

func (l Listing) Archive() (Listing, error) {
	if l.Status != StatusDraft && l.Status != StatusPublished {
		return Listing{}, ErrInvalidStatusTransition
	}
	updated := l
	updated.Status = StatusArchived
	return updated, nil
}

func (l Listing) MarkSold() (Listing, error) {
	if l.Status != StatusPublished && l.Status != StatusReserved {
		return Listing{}, ErrInvalidStatusTransition
	}
	updated := l
	updated.Status = StatusSold
	return updated, nil
}
