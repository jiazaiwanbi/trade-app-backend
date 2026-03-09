package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	orderapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/order"
	listingdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/listing"
	orderdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/order"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, input orderapp.CreateInput) (orderdomain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return orderdomain.Order{}, fmt.Errorf("begin create order tx: %w", err)
	}
	defer tx.Rollback(ctx)

	listing, err := getListingForUpdate(ctx, tx, input.ListingID)
	if err != nil {
		return orderdomain.Order{}, err
	}

	order, err := orderdomain.New(listing.ID, input.BuyerID, listing.SellerID)
	if err != nil {
		return orderdomain.Order{}, err
	}

	reserved, err := listing.Reserve()
	if err != nil {
		if errors.Is(err, listingdomain.ErrListingUnavailable) {
			return orderdomain.Order{}, orderdomain.ErrListingUnavailable
		}
		return orderdomain.Order{}, err
	}

	if err := updateListingStatusTx(ctx, tx, reserved.ID, reserved.Status); err != nil {
		return orderdomain.Order{}, err
	}

	created, err := insertOrderTx(ctx, tx, order)
	if err != nil {
		if isOrderUniqueViolation(err) {
			return orderdomain.Order{}, orderdomain.ErrListingUnavailable
		}
		return orderdomain.Order{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return orderdomain.Order{}, fmt.Errorf("commit create order tx: %w", err)
	}
	return created, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (orderdomain.Order, error) {
	order, err := scanOrder(r.pool.QueryRow(ctx, `
    SELECT id, listing_id, buyer_id, seller_id, status, created_at, updated_at
    FROM orders
    WHERE id = $1
  `, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orderdomain.Order{}, orderdomain.ErrOrderNotFound
		}
		return orderdomain.Order{}, fmt.Errorf("get order: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, order orderdomain.Order) (orderdomain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return orderdomain.Order{}, fmt.Errorf("begin update order tx: %w", err)
	}
	defer tx.Rollback(ctx)

	listing, err := getListingForUpdate(ctx, tx, order.ListingID)
	if err != nil {
		return orderdomain.Order{}, err
	}

	switch order.Status {
	case orderdomain.StatusCancelled:
		listing, err = listing.ReleaseReservation()
	case orderdomain.StatusCompleted:
		listing, err = listing.MarkSold()
	default:
		err = orderdomain.ErrInvalidStatusTransition
	}
	if err != nil {
		return orderdomain.Order{}, err
	}

	if err := updateListingStatusTx(ctx, tx, listing.ID, listing.Status); err != nil {
		return orderdomain.Order{}, err
	}

	updated, err := scanOrder(tx.QueryRow(ctx, `
    UPDATE orders
    SET status = $2, updated_at = NOW()
    WHERE id = $1
    RETURNING id, listing_id, buyer_id, seller_id, status, created_at, updated_at
  `, order.ID, order.Status))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orderdomain.Order{}, orderdomain.ErrOrderNotFound
		}
		return orderdomain.Order{}, fmt.Errorf("update order status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return orderdomain.Order{}, fmt.Errorf("commit update order tx: %w", err)
	}
	return updated, nil
}

func (r *OrderRepository) ListMine(ctx context.Context, userID int64, page int, pageSize int) ([]orderdomain.Order, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE buyer_id = $1 OR seller_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count my orders: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
    SELECT id, listing_id, buyer_id, seller_id, status, created_at, updated_at
    FROM orders
    WHERE buyer_id = $1 OR seller_id = $1
    ORDER BY created_at DESC
    LIMIT $2 OFFSET $3
  `, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list my orders: %w", err)
	}
	defer rows.Close()

	items := make([]orderdomain.Order, 0)
	for rows.Next() {
		item, err := scanOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate orders: %w", err)
	}
	return items, total, nil
}

func getListingForUpdate(ctx context.Context, tx pgx.Tx, listingID int64) (listingdomain.Listing, error) {
	listing, err := scanListing(tx.QueryRow(ctx, `
    SELECT id, seller_id, category_id, title, description, price_cents, status, created_at, updated_at
    FROM listings
    WHERE id = $1
    FOR UPDATE
  `, listingID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return listingdomain.Listing{}, listingdomain.ErrListingNotFound
		}
		return listingdomain.Listing{}, fmt.Errorf("get listing for update: %w", err)
	}
	return listing, nil
}

func updateListingStatusTx(ctx context.Context, tx pgx.Tx, listingID int64, status listingdomain.Status) error {
	commandTag, err := tx.Exec(ctx, `UPDATE listings SET status = $2, updated_at = NOW() WHERE id = $1`, listingID, status)
	if err != nil {
		return fmt.Errorf("update listing status: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return listingdomain.ErrListingNotFound
	}
	return nil
}

func insertOrderTx(ctx context.Context, tx pgx.Tx, order orderdomain.Order) (orderdomain.Order, error) {
	return scanOrder(tx.QueryRow(ctx, `
    INSERT INTO orders (listing_id, buyer_id, seller_id, status)
    VALUES ($1, $2, $3, $4)
    RETURNING id, listing_id, buyer_id, seller_id, status, created_at, updated_at
  `, order.ListingID, order.BuyerID, order.SellerID, order.Status))
}

func scanOrder(row interface{ Scan(dest ...any) error }) (orderdomain.Order, error) {
	var order orderdomain.Order
	var status string
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&order.ID, &order.ListingID, &order.BuyerID, &order.SellerID, &status, &createdAt, &updatedAt); err != nil {
		return orderdomain.Order{}, err
	}
	order.Status = orderdomain.Status(status)
	order.CreatedAt = createdAt
	order.UpdatedAt = updatedAt
	return order, nil
}

func isOrderUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
