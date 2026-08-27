package repository

import (
	"context"

	"shop/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByToken(ctx context.Context, tokenStr string) (*models.RefreshToken, error)
	DeleteByToken(ctx context.Context, tokenStr string) error
	DeleteByUserID(ctx context.Context, userID int64) error
}

type StoreRepository interface {
	Create(ctx context.Context, store *models.Store) error
	GetByID(ctx context.Context, id int64) (*models.Store, error)
	GetBySellerID(ctx context.Context, sellerID int64) (*models.Store, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetAll(ctx context.Context) ([]*models.Category, error)
	GetByID(ctx context.Context, id int64) (*models.Category, error)
}

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id int64) (*models.Product, error)
	GetByStoreID(ctx context.Context, storeID int64) ([]*models.Product, error)
	GetAll(ctx context.Context) ([]*models.Product, error)
	UpdateStock(ctx context.Context, productID int64, delta int) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id int64) (*models.Order, error)
	GetByCustomerID(ctx context.Context, customerID int64) ([]*models.Order, error)
	GetByStoreID(ctx context.Context, storeID int64) ([]*models.Order, error)
}
