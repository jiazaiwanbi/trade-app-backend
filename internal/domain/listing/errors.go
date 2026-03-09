package listing

import "errors"

var ErrListingNotFound = errors.New("listing not found")
var ErrForbidden = errors.New("forbidden")
var ErrInvalidTitle = errors.New("invalid title")
var ErrInvalidDescription = errors.New("invalid description")
var ErrInvalidPrice = errors.New("invalid price")
var ErrInvalidStatusTransition = errors.New("invalid status transition")
var ErrListingUnavailable = errors.New("listing unavailable")
