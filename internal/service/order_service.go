package service

import (
	"context"
	"fmt"

	"shop/internal/models"
	"shop/internal/repository"
	pkgerr "shop/pkg/errors"
)

type OrderService struct {
	OrderRepo   repository.OrderRepository
	ProductRepo repository.ProductRepository
	StoreRepo   repository.StoreRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	storeRepo repository.StoreRepository,
) *OrderService {
	return &OrderService{
		OrderRepo:   orderRepo,
		ProductRepo: productRepo,
		StoreRepo:   storeRepo,
	}
}

func (s *OrderService) Create(ctx context.Context, customerID int64, items []models.OrderItem) (*models.Order, error) {
	if customerID <= 0 {
		return nil, pkgerr.ErrAccessDenied
	}
	if len(items) == 0 {
		return nil, pkgerr.ErrEmptyOrder
	}

	var (
		totalPrice float64
		storeID    int64
		processed  []models.OrderItem
		deducted   []models.OrderItem
	)

	for _, item := range items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: product_id and quantity must be greater than zero", pkgerr.ErrInvalidOrder)
		}

		product, err := s.ProductRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch product %d: %w", item.ProductID, err)
		}
		if product == nil {
			return nil, fmt.Errorf("%w: product %d not found", pkgerr.ErrProductNotFound, item.ProductID)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("%w: product %s (requested: %d, available: %d)", pkgerr.ErrInsufficientStock, product.Name, item.Quantity, product.Stock)
		}

		if storeID == 0 {
			storeID = product.StoreID
		}

		// Deduct stock
		if err := s.ProductRepo.UpdateStock(ctx, item.ProductID, -item.Quantity); err != nil {
			for _, d := range deducted {
				_ = s.ProductRepo.UpdateStock(ctx, d.ProductID, d.Quantity)
			}
			return nil, fmt.Errorf("failed to update stock for product %d: %w", item.ProductID, err)
		}
		deducted = append(deducted, item)

		itemPrice := product.Price
		totalPrice += itemPrice * float64(item.Quantity)

		processed = append(processed, models.OrderItem{
			ProductID: item.ProductID,
			StoreID:   product.StoreID,
			Quantity:  item.Quantity,
			Price:     itemPrice,
			Product:   product,
		})
	}

	order := &models.Order{
		CustomerID: customerID,
		StoreID:    storeID,
		TotalPrice: totalPrice,
		Status:     models.OrderStatusPending,
		Items:      processed,
	}

	if err := s.OrderRepo.Create(ctx, order); err != nil {
		for _, d := range deducted {
			_ = s.ProductRepo.UpdateStock(ctx, d.ProductID, d.Quantity)
		}
		return nil, fmt.Errorf("failed to create order in repo: %w", err)
	}

	return order, nil
}

func (s *OrderService) GetByCustomerID(ctx context.Context, customerID int64) ([]*models.Order, error) {
	if customerID <= 0 {
		return nil, fmt.Errorf("invalid customer id")
	}

	orders, err := s.OrderRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("OrderService.GetByCustomerID: %w", err)
	}
	if orders == nil {
		orders = []*models.Order{}
	}
	return orders, nil
}

func (s *OrderService) GetByStoreID(ctx context.Context, storeID, userID int64, userRole string) ([]*models.Order, error) {
	if storeID <= 0 {
		return nil, fmt.Errorf("invalid store id")
	}

	if userRole != string(models.RoleAdmin) {
		store, err := s.StoreRepo.GetByID(ctx, storeID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch store: %w", err)
		}
		if store == nil {
			return nil, pkgerr.ErrStoreNotFound
		}
		if store.SellerID != userID {
			return nil, pkgerr.ErrAccessDenied
		}
	}

	orders, err := s.OrderRepo.GetByStoreID(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("OrderService.GetByStoreID: %w", err)
	}
	if orders == nil {
		orders = []*models.Order{}
	}
	return orders, nil
}

func (s *OrderService) GetByID(ctx context.Context, orderID, userID int64, userRole string) (*models.Order, error) {
	if orderID <= 0 {
		return nil, fmt.Errorf("invalid order id")
	}

	order, err := s.OrderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("OrderService.GetByID: %w", err)
	}
	if order == nil {
		return nil, pkgerr.ErrOrderNotFound
	}

	// Permission check: customer who placed order, seller who owns the store, or admin
	if userRole == string(models.RoleAdmin) || order.CustomerID == userID {
		return order, nil
	}

	store, err := s.StoreRepo.GetByID(ctx, order.StoreID)
	if err == nil && store != nil && store.SellerID == userID {
		return order, nil
	}

	return nil, pkgerr.ErrAccessDenied
}