package pkg

import "errors"

var (
	ErrEmptyOrder        = errors.New("order must contain at least one item")
	ErrInvalidOrder      = errors.New("invalid order data")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrOrderNotFound     = errors.New("order not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrStoreNotFound     = errors.New("store not found")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient product stock")
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrWeakPassword      = errors.New("password must be at least 6 characters long")
	ErrMultiStoreOrder   = errors.New("all items in an order must belong to the same store")
	ErrInvalidRole       = errors.New("invalid role")
)
