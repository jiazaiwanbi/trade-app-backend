package order

import "errors"

var ErrOrderNotFound = errors.New("order not found")
var ErrForbidden = errors.New("forbidden")
var ErrInvalidStatusTransition = errors.New("invalid order status transition")
var ErrCannotOrderOwnListing = errors.New("cannot order own listing")
var ErrListingUnavailable = errors.New("listing unavailable")
