package order

import "time"

type Status string

const (
	StatusPendingPayment  Status = "pending_payment"
	StatusPendingShipment Status = "pending_shipment"
	StatusShipped         Status = "shipped"
	StatusCompleted       Status = "completed"
	StatusCancelled       Status = "cancelled"
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
	return Order{ListingID: listingID, BuyerID: buyerID, SellerID: sellerID, Status: StatusPendingPayment}, nil
}

func (o Order) Cancel(actorID int64) (Order, error) {
	if actorID != o.BuyerID && actorID != o.SellerID {
		return Order{}, ErrForbidden
	}
	if o.Status != StatusPendingPayment {
		return Order{}, ErrInvalidStatusTransition
	}
	updated := o
	updated.Status = StatusCancelled
	return updated, nil
}

func (o Order) Pay(actorID int64) (Order, error) {
	if actorID != o.BuyerID {
		return Order{}, ErrForbidden
	}
	if o.Status != StatusPendingPayment {
		return Order{}, ErrInvalidStatusTransition
	}
	updated := o
	updated.Status = StatusPendingShipment
	return updated, nil
}

func (o Order) Ship(actorID int64) (Order, error) {
	if actorID != o.SellerID {
		return Order{}, ErrForbidden
	}
	if o.Status != StatusPendingShipment {
		return Order{}, ErrInvalidStatusTransition
	}
	updated := o
	updated.Status = StatusShipped
	return updated, nil
}

func (o Order) Receive(actorID int64) (Order, error) {
	if actorID != o.BuyerID {
		return Order{}, ErrForbidden
	}
	if o.Status != StatusShipped {
		return Order{}, ErrInvalidStatusTransition
	}
	updated := o
	updated.Status = StatusCompleted
	return updated, nil
}
