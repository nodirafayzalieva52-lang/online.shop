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

const productSelectCols = `id, store_id, category_id, name, description, price, stock, COALESCE(image_url, ''), created_at, updated_at`

func scanProduct(scan func(dest ...any) error) (*models.Product, error) {
	var p models.Product
	var catID *int64
	if err := scan(
		&p.ID,
		&p.StoreID,
		&catID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.Stock,
		&p.ImageURL,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if catID != nil {
		p.CategoryID = *catID
	}
	return &p, nil
}

func nullableCategoryID(id int64) *int64 {
	if id <= 0 {
		return nil
	}
	return &id
}

func (r *ProductRepo) Create(ctx context.Context, product *models.Product) error {
	query := `INSERT INTO products (store_id, category_id, name, description, price, stock, image_url)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		product.StoreID,
		nullableCategoryID(product.CategoryID),
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.ImageURL,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `SELECT ` + productSelectCols + ` FROM products WHERE id = $1`

	p, err := scanProduct(r.db.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) GetByStoreID(ctx context.Context, storeID int64) ([]*models.Product, error) {
	query := `SELECT ` + productSelectCols + ` FROM products WHERE store_id = $1 ORDER BY created_at DESC`
	return r.list(ctx, query, storeID)
}

func (r *ProductRepo) GetAll(ctx context.Context) ([]*models.Product, error) {
	query := `SELECT ` + productSelectCols + ` FROM products ORDER BY id DESC`
	return r.list(ctx, query)
}

func (r *ProductRepo) list(ctx context.Context, query string, args ...any) ([]*models.Product, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*models.Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows.Scan)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepo) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET category_id = $1, name = $2, description = $3, price = $4, stock = $5, image_url = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at`

	return r.db.QueryRow(ctx, query,
		nullableCategoryID(product.CategoryID),
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.ImageURL,
		product.ID,
	).Scan(&product.UpdatedAt)
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
