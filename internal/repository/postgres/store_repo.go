package postgres

import (
	"context"
	"errors"

	"shop/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoreRepo struct {
	db *pgxpool.Pool
}

func NewStoreRepository(db *pgxpool.Pool) *StoreRepo {
	return &StoreRepo{db: db}
}

func (r *StoreRepo) Create(ctx context.Context, store *models.Store) error {
	query := `INSERT INTO stores (seller_id, name, description)
	VALUES ($1, $2, $3) RETURNING id, created_at`

	return r.db.QueryRow(ctx, query, store.SellerID, store.Name, store.Description).
		Scan(&store.ID, &store.CreatedAt)
}

func (r *StoreRepo) GetByID(ctx context.Context, id int64) (*models.Store, error) {
	query := `SELECT id, seller_id, name, description, created_at FROM stores WHERE id = $1`

	var store models.Store
	err := r.db.QueryRow(ctx, query, id).Scan(&store.ID, &store.SellerID, &store.Name, &store.Description, &store.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepo) GetBySellerID(ctx context.Context, sellerID int64) (*models.Store, error) {
	query := `SELECT id, seller_id, name, description, created_at FROM stores WHERE seller_id = $1`

	var store models.Store
	err := r.db.QueryRow(ctx, query, sellerID).Scan(&store.ID, &store.SellerID, &store.Name, &store.Description, &store.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepo) GetAll(ctx context.Context) ([]*models.Store, error) {
	query := `SELECT id, seller_id, name, description, created_at FROM stores ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stores := make([]*models.Store, 0)
	for rows.Next() {
		var store models.Store
		if err := rows.Scan(&store.ID, &store.SellerID, &store.Name, &store.Description, &store.CreatedAt); err != nil {
			return nil, err
		}
		stores = append(stores, &store)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stores, nil
}
