package postgres

import (
	"context"
	"errors"
	"fmt"
	"shop/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	orderQuery := `INSERT INTO orders (customer_id, store_id, total_price, status)
	VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	err = tx.QueryRow(ctx, orderQuery, order.CustomerID, order.StoreID, order.TotalPrice, order.Status).
		Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `INSERT INTO order_items (order_id, product_id, store_id, quantity, price)
	VALUES ($1, $2, $3, $4, $5) RETURNING id`

	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		if order.Items[i].StoreID == 0 {
			order.Items[i].StoreID = order.StoreID
		}
		err = tx.QueryRow(ctx, itemQuery,
			order.Items[i].OrderID,
			order.Items[i].ProductID,
			order.Items[i].StoreID,
			order.Items[i].Quantity,
			order.Items[i].Price,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit order transaction: %w", err)
	}

	return nil
}

func (r *OrderRepository) getOrderItems(ctx context.Context, orderID int64) ([]models.OrderItem, error) {
	query := `SELECT id, order_id, product_id, store_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.StoreID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *OrderRepository) GetByCustomerID(ctx context.Context, customerID int64) ([]*models.Order, error) {
	query := `SELECT id, customer_id, store_id, total_price, status, created_at
	FROM orders WHERE customer_id = $1 ORDER BY id DESC`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.StoreID,
			&o.TotalPrice,
			&o.Status,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	for _, o := range orders {
		items, err := r.getOrderItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
	}

	return orders, nil
}

func (r *OrderRepository) GetByStoreID(ctx context.Context, storeID int64) ([]*models.Order, error) {
	query := `SELECT id, customer_id, store_id, total_price, status, created_at FROM orders WHERE store_id = $1 ORDER BY id DESC`

	rows, err := r.db.Query(ctx, query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.StoreID,
			&o.TotalPrice,
			&o.Status,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	for _, o := range orders {
		items, err := r.getOrderItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
	}

	return orders, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	query := `SELECT id, customer_id, store_id, total_price, status, created_at FROM orders WHERE id = $1`

	var o models.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.ID,
		&o.CustomerID,
		&o.StoreID,
		&o.TotalPrice,
		&o.Status,
		&o.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}
