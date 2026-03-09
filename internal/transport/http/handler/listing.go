package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	categoryapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/category"
	listingapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/listing"
	categorydomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/category"
	listingdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/listing"
	platformmiddleware "github.com/jiazaiwanbi/second-hand-platform/internal/platform/http/middleware"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
	"github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/dto"
)

const listingTimeLayout = "2006-01-02T15:04:05Z07:00"

type ListingHandler struct {
	listingService  *listingapp.Service
	categoryService *categoryapp.Service
}

func NewListingHandler(listingService *listingapp.Service, categoryService *categoryapp.Service) *ListingHandler {
	return &ListingHandler{listingService: listingService, categoryService: categoryService}
}

func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "authentication required"})
		return
	}

	var request dto.CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid request body"})
		return
	}

	item, err := h.listingService.Create(r.Context(), listingapp.CreateInput{
		SellerID:    userID,
		CategoryID:  request.CategoryID,
		Title:       request.Title,
		Description: request.Description,
		PriceCents:  request.PriceCents,
		Publish:     request.Publish,
	})
	if err != nil {
		h.handleListingError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, toListingResponse(item))
}

func (h *ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	filter := listingapp.ListFilter{Page: page, PageSize: pageSize, Keyword: r.URL.Query().Get("keyword")}
	if rawCategoryID := r.URL.Query().Get("category_id"); rawCategoryID != "" {
		categoryID, err := strconv.ParseInt(rawCategoryID, 10, 64)
		if err != nil {
			sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid category_id"})
			return
		}
		filter.CategoryID = &categoryID
	}

	items, total, err := h.listingService.List(r.Context(), filter)
	if err != nil {
		h.handleListingError(w, r, err)
		return
	}

	responses := make([]dto.ListingResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toListingResponse(item))
	}

	sharedhttp.WriteOK(w, dto.ListResponse[dto.ListingResponse]{Page: page, PageSize: pageSize, Total: total, Items: responses})
}

func (h *ListingHandler) Get(w http.ResponseWriter, r *http.Request) {
	listingID, err := parseIDParam(r.PathValue("id"))
	if err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid listing id"})
		return
	}

	item, err := h.listingService.Get(r.Context(), listingID)
	if err != nil {
		h.handleListingError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, toListingResponse(item))
}

func (h *ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "authentication required"})
		return
	}

	listingID, err := parseIDParam(r.PathValue("id"))
	if err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid listing id"})
		return
	}

	var request dto.UpdateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: "invalid request body"})
		return
	}

	var status *listingdomain.Status
	if request.Status != nil {
		parsed := listingdomain.Status(*request.Status)
		status = &parsed
	}

	item, err := h.listingService.Update(r.Context(), listingapp.UpdateInput{
		ListingID:   listingID,
		SellerID:    userID,
		CategoryID:  request.CategoryID,
		Title:       request.Title,
		Description: request.Description,
		PriceCents:  request.PriceCents,
		Status:      status,
	})
	if err != nil {
		h.handleListingError(w, r, err)
		return
	}

	sharedhttp.WriteOK(w, toListingResponse(item))
}

func (h *ListingHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformmiddleware.UserIDFromContext(r.Context())
	if !ok {
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "authentication required"})
		return
	}

	page, pageSize := parsePagination(r)
	items, total, err := h.listingService.ListMine(r.Context(), userID, page, pageSize)
	if err != nil {
		h.handleListingError(w, r, err)
		return
	}

	responses := make([]dto.ListingResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toListingResponse(item))
	}

	sharedhttp.WriteOK(w, dto.ListResponse[dto.ListingResponse]{Page: page, PageSize: pageSize, Total: total, Items: responses})
}

func (h *ListingHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.categoryService.List(r.Context())
	if err != nil {
		sharedhttp.WriteError(w, r, sharedhttp.NewInternalError())
		return
	}

	responses := make([]dto.CategoryResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toCategoryResponse(item))
	}

	sharedhttp.WriteOK(w, responses)
}

func (h *ListingHandler) handleListingError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, listingdomain.ErrListingNotFound):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusNotFound, Code: "NOT_FOUND", Message: "listing not found"})
	case errors.Is(err, listingdomain.ErrForbidden):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusForbidden, Code: "FORBIDDEN", Message: "forbidden"})
	case errors.Is(err, listingdomain.ErrInvalidTitle), errors.Is(err, listingdomain.ErrInvalidDescription), errors.Is(err, listingdomain.ErrInvalidPrice), errors.Is(err, listingdomain.ErrInvalidStatusTransition):
		sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: err.Error()})
	default:
		sharedhttp.WriteError(w, r, sharedhttp.NewInternalError())
	}
}

func toListingResponse(item listingdomain.Listing) dto.ListingResponse {
	return dto.ListingResponse{
		ID:          item.ID,
		SellerID:    item.SellerID,
		CategoryID:  item.CategoryID,
		Title:       item.Title,
		Description: item.Description,
		PriceCents:  item.PriceCents,
		Status:      string(item.Status),
		CreatedAt:   item.CreatedAt.Format(listingTimeLayout),
		UpdatedAt:   item.UpdatedAt.Format(listingTimeLayout),
	}
}

func toCategoryResponse(item categorydomain.Category) dto.CategoryResponse {
	return dto.CategoryResponse{ID: item.ID, Name: item.Name}
}

func parsePagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 20
	if rawPage := r.URL.Query().Get("page"); rawPage != "" {
		if parsed, err := strconv.Atoi(rawPage); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if rawPageSize := r.URL.Query().Get("page_size"); rawPageSize != "" {
		if parsed, err := strconv.Atoi(rawPageSize); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	return page, pageSize
}

func parseIDParam(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
