package postgres

import (
	"context"
	"errors"

	"shop/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, product *models.Product) error {
	query := `INSERT INTO products (store_id, category_id, name, description, price, stock)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		product.StoreID,
		product.CategoryID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `SELECT id, store_id, category_id, name, description, price, stock, created_at, updated_at
	FROM products WHERE id = $1`

	var p models.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.StoreID, &p.CategoryID, &p.Name,
		&p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByStoreID(ctx context.Context, storeID int64) ([]*models.Product, error) {
	query := `SELECT id, store_id, category_id, name, description, price, stock, created_at, updated_at
	FROM products WHERE store_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.StoreID, &p.CategoryID, &p.Name,
			&p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, nil
}

func (r *ProductRepo) GetAll(ctx context.Context) ([]*models.Product, error) {
	query := `SELECT id, store_id, category_id, name, description, price, stock, created_at, updated_at
	FROM products ORDER BY id DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID,
			&p.StoreID,
			&p.CategoryID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, nil
}

func (r *ProductRepo) UpdateStock(ctx context.Context, productID int64, delta int) error {
	query := `UPDATE products SET stock = stock + $1, updated_at = NOW() WHERE id = $2 AND stock + $1 >= 0`

	result, err := r.db.Exec(ctx, query, delta, productID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("insufficient stock or product not found")
	}
	return nil
}
