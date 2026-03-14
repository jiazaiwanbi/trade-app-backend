package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	orderapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/order"
	orderdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/order"
	platformmiddleware "github.com/jiazaiwanbi/second-hand-platform/internal/platform/http/middleware"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
	"github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/dto"
)

type OrderHandler struct {
	service *orderapp.Service
}

func NewOrderHandler(service *orderapp.Service) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "authentication required"})
		return
	}

	var request dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid request body"})
		return
	}

	item, err := h.service.Create(r.Context(), orderapp.CreateInput{ListingID: request.ListingID, BuyerID: userID})
	if err != nil {
		h.handleOrderError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, toOrderResponse(item))
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.applyAction(w, r, "cancel")
}

func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	h.applyAction(w, r, "pay")
}

func (h *OrderHandler) Ship(w http.ResponseWriter, r *http.Request) {
	h.applyAction(w, r, "ship")
}

func (h *OrderHandler) Receive(w http.ResponseWriter, r *http.Request) {
	h.applyAction(w, r, "receive")
}

func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "authentication required"})
		return
	}

	page, pageSize := parsePagination(r)
	items, total, err := h.service.ListMine(r.Context(), userID, page, pageSize)
	if err != nil {
		h.handleOrderError(w, r, err)
		return
	}

	responses := make([]dto.OrderResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toOrderResponse(item))
	}

	sharedhttp.WriteOK(w, dto.ListResponse[dto.OrderResponse]{Page: page, PageSize: pageSize, Total: total, Items: responses})
}

func (h *OrderHandler) applyAction(w http.ResponseWriter, r *http.Request, action string) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "authentication required"})
		return
	}

	orderID, err := parseIDParam(r.PathValue("id"))
	if err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid order id"})
		return
	}

	input := orderapp.ActionInput{OrderID: orderID, ActorID: userID}
	var item orderdomain.Order
	switch action {
	case "cancel":
		item, err = h.service.Cancel(r.Context(), input)
	case "pay":
		item, err = h.service.Pay(r.Context(), input)
	case "ship":
		item, err = h.service.Ship(r.Context(), input)
	case "receive":
		item, err = h.service.Receive(r.Context(), input)
	default:
		err = orderdomain.ErrInvalidStatusTransition
	}
	if err != nil {
		h.handleOrderError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, toOrderResponse(item))
}

func (h *OrderHandler) handleOrderError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, orderdomain.ErrOrderNotFound):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusNotFound, Code: "NOT_FOUND", Message: "order not found"})
	case errors.Is(err, orderdomain.ErrForbidden):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusForbidden, Code: "FORBIDDEN", Message: "forbidden"})
	case errors.Is(err, orderdomain.ErrCannotOrderOwnListing), errors.Is(err, orderdomain.ErrInvalidStatusTransition), errors.Is(err, orderdomain.ErrListingUnavailable):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: err.Error()})
	default:
		sharedhttp.WriteError(w, r, sharedhttp.NewInternalError())
	}
}

func toOrderResponse(item orderdomain.Order) dto.OrderResponse {
	return dto.OrderResponse{
		ID:        item.ID,
		ListingID: item.ListingID,
		BuyerID:   item.BuyerID,
		SellerID:  item.SellerID,
		Status:    string(item.Status),
		CreatedAt: item.CreatedAt.Format(listingTimeLayout),
		UpdatedAt: item.UpdatedAt.Format(listingTimeLayout),
	}
}
