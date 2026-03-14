package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	listingapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/listing"
	listingdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/listing"
)

type ListingRepository struct {
	pool *pgxpool.Pool
}

func NewListingRepository(pool *pgxpool.Pool) *ListingRepository {
	return &ListingRepository{pool: pool}
}

func (r *ListingRepository) Create(ctx context.Context, listing listingdomain.Listing) (listingdomain.Listing, error) {
	query := `
    INSERT INTO listings (seller_id, category_id, title, description, price_cents, image_urls, status)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id, seller_id, category_id, title, description, price_cents, image_urls, status, created_at, updated_at
  `
	created, err := scanListing(r.pool.QueryRow(ctx, query, listing.SellerID, listing.CategoryID, listing.Title, listing.Description, listing.PriceCents, listing.ImageURLs, listing.Status))
	if err != nil {
		return listingdomain.Listing{}, fmt.Errorf("create listing: %w", err)
	}
	return created, nil
}

func (r *ListingRepository) GetByID(ctx context.Context, id int64) (listingdomain.Listing, error) {
	query := `
    SELECT id, seller_id, category_id, title, description, price_cents, image_urls, status, created_at, updated_at
    FROM listings
    WHERE id = $1
  `
	listing, err := scanListing(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return listingdomain.Listing{}, listingdomain.ErrListingNotFound
		}
		return listingdomain.Listing{}, fmt.Errorf("get listing by id: %w", err)
	}
	return listing, nil
}

func (r *ListingRepository) Update(ctx context.Context, listing listingdomain.Listing) (listingdomain.Listing, error) {
	query := `
    UPDATE listings
    SET category_id = $2, title = $3, description = $4, price_cents = $5, image_urls = $6, status = $7, updated_at = NOW()
    WHERE id = $1
    RETURNING id, seller_id, category_id, title, description, price_cents, image_urls, status, created_at, updated_at
  `
	updated, err := scanListing(r.pool.QueryRow(ctx, query, listing.ID, listing.CategoryID, listing.Title, listing.Description, listing.PriceCents, listing.ImageURLs, listing.Status))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return listingdomain.Listing{}, listingdomain.ErrListingNotFound
		}
		return listingdomain.Listing{}, fmt.Errorf("update listing: %w", err)
	}
	return updated, nil
}

func (r *ListingRepository) List(ctx context.Context, filter listingapp.ListFilter) ([]listingdomain.Listing, int, error) {
	where := []string{"status = 'published'"}
	args := []any{}
	argIndex := 1

	if filter.Keyword != "" {
		where = append(where, fmt.Sprintf("LOWER(title) LIKE $%d", argIndex))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(filter.Keyword))+"%")
		argIndex++
	}
	if filter.CategoryID != nil {
		where = append(where, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM listings WHERE " + strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count listings: %w", err)
	}

	query := fmt.Sprintf(`
    SELECT id, seller_id, category_id, title, description, price_cents, image_urls, status, created_at, updated_at
    FROM listings
    WHERE %s
    ORDER BY created_at DESC
    LIMIT $%d OFFSET $%d
  `, strings.Join(where, " AND "), argIndex, argIndex+1)
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	return r.listByQuery(ctx, query, args, total)
}

func (r *ListingRepository) ListBySeller(ctx context.Context, sellerID int64, page int, pageSize int) ([]listingdomain.Listing, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM listings WHERE seller_id = $1`, sellerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count seller listings: %w", err)
	}
	query := `
    SELECT id, seller_id, category_id, title, description, price_cents, image_urls, status, created_at, updated_at
    FROM listings
    WHERE seller_id = $1
    ORDER BY created_at DESC
    LIMIT $2 OFFSET $3
  `
	return r.listByQuery(ctx, query, []any{sellerID, pageSize, (page - 1) * pageSize}, total)
}

func (r *ListingRepository) listByQuery(ctx context.Context, query string, args []any, total int) ([]listingdomain.Listing, int, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list listings: %w", err)
	}
	defer rows.Close()

	items := make([]listingdomain.Listing, 0)
	for rows.Next() {
		item, err := scanListing(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan listing: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate listings: %w", err)
	}
	return items, total, nil
}

func scanListing(row interface{ Scan(dest ...any) error }) (listingdomain.Listing, error) {
	var listing listingdomain.Listing
	var categoryID *int64
	var imageURLs []string
	var status string
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&listing.ID, &listing.SellerID, &categoryID, &listing.Title, &listing.Description, &listing.PriceCents, &imageURLs, &status, &createdAt, &updatedAt); err != nil {
		return listingdomain.Listing{}, err
	}
	listing.CategoryID = categoryID
	listing.ImageURLs = imageURLs
	listing.Status = listingdomain.Status(status)
	listing.CreatedAt = createdAt
	listing.UpdatedAt = updatedAt
	return listing, nil
}
