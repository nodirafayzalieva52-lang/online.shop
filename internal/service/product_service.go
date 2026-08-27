package service

import (
	"context"
	"fmt"

	"shop/internal/models"
	"shop/internal/repository"
	pkgerr "shop/pkg/errors"
)

type ProductService struct {
	ProductRepo repository.ProductRepository
	StoreRepo   repository.StoreRepository
}

func NewProductService(productRepo repository.ProductRepository, storeRepo repository.StoreRepository) *ProductService {
	return &ProductService{
		ProductRepo: productRepo,
		StoreRepo:   storeRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, userID int64, userRole string, storeID, categoryID int64, name, description string, price float64, stock int) (*models.Product, error) {
	if storeID <= 0 {
		return nil, fmt.Errorf("invalid store id")
	}
	if name == "" {
		return nil, fmt.Errorf("product name cannot be empty")
	}
	if price <= 0 {
		return nil, fmt.Errorf("product price must be greater than zero")
	}
	if stock < 0 {
		return nil, fmt.Errorf("product stock cannot be negative")
	}

	store, err := s.StoreRepo.GetByID(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch store: %w", err)
	}
	if store == nil {
		return nil, pkgerr.ErrStoreNotFound
	}

	if userRole != string(models.RoleAdmin) && store.SellerID != userID {
		return nil, pkgerr.ErrAccessDenied
	}

	product := &models.Product{
		StoreID:     storeID,
		CategoryID:  categoryID,
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
	}

	err = s.ProductRepo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("productRepo.Create: %w", err)
	}

	return product, nil
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid product id")
	}

	product, err := s.ProductRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error fetching product: %w", err)
	}
	if product == nil {
		return nil, pkgerr.ErrProductNotFound
	}

	return product, nil
}

func (s *ProductService) GetByStoreID(ctx context.Context, storeID int64) ([]*models.Product, error) {
	if storeID <= 0 {
		return nil, fmt.Errorf("invalid store id")
	}

	products, err := s.ProductRepo.GetByStoreID(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("error fetching store products: %w", err)
	}

	if products == nil {
		products = []*models.Product{}
	}

	return products, nil
}

func (s *ProductService) GetAll(ctx context.Context) ([]*models.Product, error) {
	products, err := s.ProductRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching all products: %w", err)
	}
	if products == nil {
		products = []*models.Product{}
	}
	return products, nil
}
