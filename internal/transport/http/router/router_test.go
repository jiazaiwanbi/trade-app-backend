package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/auth"
	categoryapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/category"
	listingapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/listing"
	orderapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/order"
	userapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/user"
	authdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/auth"
	categorydomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/category"
	listingdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/listing"
	orderdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/order"
	userdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/user"
	platformauth "github.com/jiazaiwanbi/second-hand-platform/internal/platform/auth"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/logger"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
	"github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/handler"
)

type memoryUserRepo struct {
	users  map[int64]userdomain.User
	byMail map[string]int64
	nextID int64
}

type memoryListingRepo struct {
	listings map[int64]listingdomain.Listing
	nextID   int64
}

type memoryOrderRepo struct {
	orders   map[int64]orderdomain.Order
	nextID   int64
	listings *memoryListingRepo
}

type memoryCategoryRepo struct {
	items []categorydomain.Category
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{users: map[int64]userdomain.User{}, byMail: map[string]int64{}, nextID: 1}
}

func newMemoryListingRepo() *memoryListingRepo {
	return &memoryListingRepo{listings: map[int64]listingdomain.Listing{}, nextID: 1}
}

func newMemoryOrderRepo(listings *memoryListingRepo) *memoryOrderRepo {
	return &memoryOrderRepo{orders: map[int64]orderdomain.Order{}, nextID: 1, listings: listings}
}

func newMemoryCategoryRepo() *memoryCategoryRepo {
	return &memoryCategoryRepo{items: []categorydomain.Category{{ID: 1, Name: "electronics"}, {ID: 2, Name: "books"}}}
}

func (r *memoryUserRepo) Create(_ context.Context, user userdomain.User) (userdomain.User, error) {
	if _, exists := r.byMail[user.Email]; exists {
		return userdomain.User{}, authdomain.ErrEmailAlreadyExists
	}
	user.ID = r.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	user.Bio = ""
	r.nextID++
	r.users[user.ID] = user
	r.byMail[user.Email] = user.ID
	return user, nil
}

func (r *memoryUserRepo) FindByEmail(_ context.Context, email string) (userdomain.User, error) {
	id, ok := r.byMail[email]
	if !ok {
		return userdomain.User{}, userdomain.ErrUserNotFound
	}
	return r.users[id], nil
}

func (r *memoryUserRepo) FindByID(_ context.Context, id int64) (userdomain.User, error) {
	user, ok := r.users[id]
	if !ok {
		return userdomain.User{}, userdomain.ErrUserNotFound
	}
	return user, nil
}

func (r *memoryUserRepo) UpdateProfile(_ context.Context, id int64, nickname string, bio string) (userdomain.User, error) {
	user, ok := r.users[id]
	if !ok {
		return userdomain.User{}, userdomain.ErrUserNotFound
	}
	user.Nickname = nickname
	user.Bio = bio
	user.UpdatedAt = time.Now()
	r.users[id] = user
	return user, nil
}

func (r *memoryListingRepo) Create(_ context.Context, listing listingdomain.Listing) (listingdomain.Listing, error) {
	listing.ID = r.nextID
	listing.CreatedAt = time.Now()
	listing.UpdatedAt = listing.CreatedAt
	r.nextID++
	r.listings[listing.ID] = listing
	return listing, nil
}

func (r *memoryListingRepo) GetByID(_ context.Context, id int64) (listingdomain.Listing, error) {
	item, ok := r.listings[id]
	if !ok {
		return listingdomain.Listing{}, listingdomain.ErrListingNotFound
	}
	return item, nil
}

func (r *memoryListingRepo) Update(_ context.Context, listing listingdomain.Listing) (listingdomain.Listing, error) {
	if _, ok := r.listings[listing.ID]; !ok {
		return listingdomain.Listing{}, listingdomain.ErrListingNotFound
	}
	listing.UpdatedAt = time.Now()
	r.listings[listing.ID] = listing
	return listing, nil
}

func (r *memoryListingRepo) List(_ context.Context, filter listingapp.ListFilter) ([]listingdomain.Listing, int, error) {
	items := make([]listingdomain.Listing, 0)
	for _, item := range r.listings {
		if item.Status != listingdomain.StatusPublished {
			continue
		}
		if filter.CategoryID != nil && (item.CategoryID == nil || *item.CategoryID != *filter.CategoryID) {
			continue
		}
		if filter.Keyword != "" && !bytes.Contains(bytes.ToLower([]byte(item.Title)), bytes.ToLower([]byte(filter.Keyword))) {
			continue
		}
		items = append(items, item)
	}
	return items, len(items), nil
}

func (r *memoryListingRepo) ListBySeller(_ context.Context, sellerID int64, _ int, _ int) ([]listingdomain.Listing, int, error) {
	items := make([]listingdomain.Listing, 0)
	for _, item := range r.listings {
		if item.SellerID == sellerID {
			items = append(items, item)
		}
	}
	return items, len(items), nil
}

func (r *memoryOrderRepo) Create(_ context.Context, input orderapp.CreateInput) (orderdomain.Order, error) {
	listing, ok := r.listings.listings[input.ListingID]
	if !ok {
		return orderdomain.Order{}, listingdomain.ErrListingNotFound
	}
	order, err := orderdomain.New(listing.ID, input.BuyerID, listing.SellerID)
	if err != nil {
		return orderdomain.Order{}, err
	}
	reserved, err := listing.Reserve()
	if err != nil {
		return orderdomain.Order{}, orderdomain.ErrListingUnavailable
	}
	r.listings.listings[listing.ID] = reserved
	order.ID = r.nextID
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt
	r.nextID++
	r.orders[order.ID] = order
	return order, nil
}

func (r *memoryOrderRepo) GetByID(_ context.Context, id int64) (orderdomain.Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return orderdomain.Order{}, orderdomain.ErrOrderNotFound
	}
	return order, nil
}

func (r *memoryOrderRepo) UpdateStatus(_ context.Context, order orderdomain.Order) (orderdomain.Order, error) {
	current, ok := r.orders[order.ID]
	if !ok {
		return orderdomain.Order{}, orderdomain.ErrOrderNotFound
	}
	listing := r.listings.listings[current.ListingID]
	var updatedListing listingdomain.Listing
	var err error
	switch order.Status {
	case orderdomain.StatusCancelled:
		updatedListing, err = listing.ReleaseReservation()
	case orderdomain.StatusCompleted:
		updatedListing, err = listing.MarkSold()
	default:
		err = orderdomain.ErrInvalidStatusTransition
	}
	if err != nil {
		return orderdomain.Order{}, err
	}
	r.listings.listings[current.ListingID] = updatedListing
	current.Status = order.Status
	current.UpdatedAt = time.Now()
	r.orders[current.ID] = current
	return current, nil
}

func (r *memoryOrderRepo) ListMine(_ context.Context, userID int64, _ int, _ int) ([]orderdomain.Order, int, error) {
	items := make([]orderdomain.Order, 0)
	for _, item := range r.orders {
		if item.BuyerID == userID || item.SellerID == userID {
			items = append(items, item)
		}
	}
	return items, len(items), nil
}

func (r *memoryCategoryRepo) List(_ context.Context) ([]categorydomain.Category, error) {
	return r.items, nil
}

type stubReadyHandler struct{}

func (stubReadyHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	sharedhttp.WriteOK(w, map[string]string{"status": "ready"})
}

func buildTestRouter(t *testing.T) http.Handler {
	t.Helper()
	userRepo := newMemoryUserRepo()
	listingRepo := newMemoryListingRepo()
	orderRepo := newMemoryOrderRepo(listingRepo)
	categoryRepo := newMemoryCategoryRepo()
	tokenManager := platformauth.NewTokenManager("test-secret", time.Hour)
	authHandler := handler.NewAuthHandler(authapp.NewService(userRepo, tokenManager))
	userHandler := handler.NewUserHandler(userapp.NewService(userRepo))
	listingHandler := handler.NewListingHandler(listingapp.NewService(listingRepo), categoryapp.NewService(categoryRepo))
	orderHandler := handler.NewOrderHandler(orderapp.NewService(orderRepo))
	return New(logger.New(), time.Second, tokenManager, stubReadyHandler{}, authHandler, userHandler, listingHandler, orderHandler)
}

func registerAndToken(t *testing.T, router http.Handler, email string) string {
	t.Helper()
	registerBody := []byte(`{"email":"` + email + `","password":"secret12","nickname":"tester"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusOK {
		t.Fatalf("register expected 200, got %d", registerRec.Code)
	}

	var registerResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(registerRec.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	return registerResp.Data.AccessToken
}

func TestRegisterLoginAndMeFlow(t *testing.T) {
	t.Parallel()
	router := buildTestRouter(t)
	token := registerAndToken(t, router, "test@example.com")

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me expected 200, got %d", meRec.Code)
	}
}

func TestUnauthorizedMe(t *testing.T) {
	t.Parallel()
	router := buildTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	t.Parallel()
	router := buildTestRouter(t)
	_ = registerAndToken(t, router, "test@example.com")

	loginBody := []byte(`{"email":"test@example.com","password":"wrongpass"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", loginRec.Code)
	}
}

func TestCreateListAndForbidOtherSellerUpdate(t *testing.T) {
	t.Parallel()
	router := buildTestRouter(t)
	sellerToken := registerAndToken(t, router, "seller@example.com")
	otherToken := registerAndToken(t, router, "other@example.com")

	createBody := []byte(`{"category_id":1,"title":"MacBook Air","description":"Good condition laptop","price_cents":599900,"publish":true}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/listings", bytes.NewReader(createBody))
	createReq.Header.Set("Authorization", "Bearer "+sellerToken)
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create listing expected 200, got %d body=%s", createRec.Code, createRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/listings?keyword=macbook", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list listings expected 200, got %d", listRec.Code)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/listings/1", nil)
	detailRec := httptest.NewRecorder()
	router.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("get listing expected 200, got %d", detailRec.Code)
	}

	updateBody := []byte(`{"category_id":1,"title":"Hijacked","description":"Nope nope","price_cents":100,"status":"archived"}`)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/listings/1", bytes.NewReader(updateBody))
	updateReq.Header.Set("Authorization", "Bearer "+otherToken)
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusForbidden {
		t.Fatalf("forbidden update expected 403, got %d body=%s", updateRec.Code, updateRec.Body.String())
	}
}

func TestOrderCreateCompleteAndListingStateFlow(t *testing.T) {
	t.Parallel()
	router := buildTestRouter(t)
	sellerToken := registerAndToken(t, router, "seller2@example.com")
	buyerToken := registerAndToken(t, router, "buyer2@example.com")

	createListingBody := []byte(`{"category_id":1,"title":"iPad Pro","description":"tablet in good condition","price_cents":320000,"publish":true}`)
	createListingReq := httptest.NewRequest(http.MethodPost, "/api/v1/listings", bytes.NewReader(createListingBody))
	createListingReq.Header.Set("Authorization", "Bearer "+sellerToken)
	createListingReq.Header.Set("Content-Type", "application/json")
	createListingRec := httptest.NewRecorder()
	router.ServeHTTP(createListingRec, createListingReq)
	if createListingRec.Code != http.StatusOK {
		t.Fatalf("create listing expected 200, got %d body=%s", createListingRec.Code, createListingRec.Body.String())
	}

	createOrderBody := []byte(`{"listing_id":1}`)
	createOrderReq := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(createOrderBody))
	createOrderReq.Header.Set("Authorization", "Bearer "+buyerToken)
	createOrderReq.Header.Set("Content-Type", "application/json")
	createOrderRec := httptest.NewRecorder()
	router.ServeHTTP(createOrderRec, createOrderReq)
	if createOrderRec.Code != http.StatusOK {
		t.Fatalf("create order expected 200, got %d body=%s", createOrderRec.Code, createOrderRec.Body.String())
	}

	publicListReq := httptest.NewRequest(http.MethodGet, "/api/v1/listings?keyword=ipad", nil)
	publicListRec := httptest.NewRecorder()
	router.ServeHTTP(publicListRec, publicListReq)
	if publicListRec.Code != http.StatusOK {
		t.Fatalf("public list expected 200, got %d", publicListRec.Code)
	}

	completeReq := httptest.NewRequest(http.MethodPost, "/api/v1/orders/1/complete", nil)
	completeReq.Header.Set("Authorization", "Bearer "+buyerToken)
	completeRec := httptest.NewRecorder()
	router.ServeHTTP(completeRec, completeReq)
	if completeRec.Code != http.StatusOK {
		t.Fatalf("complete order expected 200, got %d body=%s", completeRec.Code, completeRec.Body.String())
	}

	myOrdersReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/orders", nil)
	myOrdersReq.Header.Set("Authorization", "Bearer "+buyerToken)
	myOrdersRec := httptest.NewRecorder()
	router.ServeHTTP(myOrdersRec, myOrdersReq)
	if myOrdersRec.Code != http.StatusOK {
		t.Fatalf("my orders expected 200, got %d", myOrdersRec.Code)
	}
}
