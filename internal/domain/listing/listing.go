package listing

import (
	"net/url"
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
	maxImages              = 10
)

type Listing struct {
	ID          int64
	SellerID    int64
	CategoryID  *int64
	Title       string
	Description string
	PriceCents  int64
	ImageURLs   []string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func New(sellerID int64, categoryID *int64, title string, description string, priceCents int64, imageURLs []string) (Listing, error) {
	listing := Listing{
		SellerID:    sellerID,
		CategoryID:  categoryID,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		PriceCents:  priceCents,
		ImageURLs:   normalizeImageURLs(imageURLs),
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
	if len(l.ImageURLs) > maxImages {
		return ErrTooManyImages
	}
	for _, imageURL := range l.ImageURLs {
		parsed, err := url.ParseRequestURI(imageURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return ErrInvalidImageURL
		}
	}
	return nil
}

func (l Listing) Update(categoryID *int64, title string, description string, priceCents int64, imageURLs []string) (Listing, error) {
	updated := l
	updated.CategoryID = categoryID
	updated.Title = strings.TrimSpace(title)
	updated.Description = strings.TrimSpace(description)
	updated.PriceCents = priceCents
	updated.ImageURLs = normalizeImageURLs(imageURLs)

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

func normalizeImageURLs(imageURLs []string) []string {
	normalized := make([]string, 0, len(imageURLs))
	for _, imageURL := range imageURLs {
		trimmed := strings.TrimSpace(imageURL)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}
