package order

import "time"

type Status string

const (
	StatusCreated   Status = "created"
	StatusCancelled Status = "cancelled"
	StatusCompleted Status = "completed"
)

type Order struct {
	ID        int64
	ListingID int64
	BuyerID   int64
	SellerID  int64
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(listingID int64, buyerID int64, sellerID int64) (Order, error) {
	if listingID <= 0 || buyerID <= 0 || sellerID <= 0 {
		return Order{}, ErrForbidden
	}
	if buyerID == sellerID {
		return Order{}, ErrCannotOrderOwnListing
	}
	return Order{ListingID: listingID, BuyerID: buyerID, SellerID: sellerID, Status: StatusCreated}, nil
}

func (o Order) Cancel(actorID int64) (Order, error) {
	if actorID != o.BuyerID && actorID != o.SellerID {
		return Order{}, ErrForbidden
	}
	if o.Status != StatusCreated {
		return Order{}, ErrInvalidStatusTransition
	}
	updated := o
	updated.Status = StatusCancelled
	return updated, nil
}

func (o Order) Complete(actorID int64) (Order, error) {
	if actorID != o.BuyerID && actorID != o.SellerID {
		return Order{}, ErrForbidden
	}
	if o.Status != StatusCreated {
		return Order{}, ErrInvalidStatusTransition
	}
	updated := o
	updated.Status = StatusCompleted
	return updated, nil
}
