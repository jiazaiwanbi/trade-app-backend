package dto

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateMeRequest struct {
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

type CreateListingRequest struct {
	CategoryID  *int64   `json:"category_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PriceCents  int64    `json:"price_cents"`
	ImageURLs   []string `json:"image_urls"`
	Publish     bool     `json:"publish"`
}

type UpdateListingRequest struct {
	CategoryID  *int64   `json:"category_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PriceCents  int64    `json:"price_cents"`
	ImageURLs   []string `json:"image_urls"`
	Status      *string  `json:"status"`
}

type CreateOrderRequest struct {
	ListingID int64 `json:"listing_id"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type ListingResponse struct {
	ID          int64    `json:"id"`
	SellerID    int64    `json:"seller_id"`
	CategoryID  *int64   `json:"category_id,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PriceCents  int64    `json:"price_cents"`
	ImageURLs   []string `json:"image_urls"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type CategoryResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type OrderResponse struct {
	ID        int64  `json:"id"`
	ListingID int64  `json:"listing_id"`
	BuyerID   int64  `json:"buyer_id"`
	SellerID  int64  `json:"seller_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListResponse[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Items    []T `json:"items"`
}
