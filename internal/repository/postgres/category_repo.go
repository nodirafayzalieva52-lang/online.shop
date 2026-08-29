package postgres

import (
	"context"
	"errors"

	"shop/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, category *models.Category) error {
	query := `INSERT INTO categories (name) VALUES ($1) RETURNING id`

	return r.db.QueryRow(ctx, query, category.Name).Scan(&category.ID)
}

func (r *CategoryRepo) GetAll(ctx context.Context) ([]*models.Category, error) {
	query := `SELECT id, name FROM categories ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]*models.Category, 0)
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}
		categories = append(categories, &cat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	query := `SELECT id, name FROM categories WHERE id = $1`

	var cat models.Category
	err := r.db.QueryRow(ctx, query, id).Scan(&cat.ID, &cat.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}
