package service

import (
	"context"
	"fmt"
	"strings"

	"shop/internal/models"
	"shop/internal/repository"
	pkgerr "shop/pkg/errors"
)

type StoreService struct {
	storeRepo repository.StoreRepository
}

func NewStoreService(storeRepo repository.StoreRepository) *StoreService {
	return &StoreService{
		storeRepo: storeRepo,
	}
}

func (s *StoreService) CreateStore(ctx context.Context, sellerID int64, name, description string) (*models.Store, error) {
	if sellerID <= 0 {
		return nil, fmt.Errorf("invalid seller id")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("store name cannot be empty")
	}

	existing, err := s.storeRepo.GetBySellerID(ctx, sellerID)
	if err != nil {
		return nil, fmt.Errorf("storeRepo.GetBySellerID: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("seller already has a store")
	}

	store := &models.Store{
		SellerID:    sellerID,
		Name:        name,
		Description: description,
	}

	err = s.storeRepo.Create(ctx, store)
	if err != nil {
		return nil, fmt.Errorf("storeRepo.Create: %w", err)
	}

	return store, nil
}

func (s *StoreService) GetByID(ctx context.Context, id int64) (*models.Store, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid store id")
	}

	store, err := s.storeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error fetching store: %w", err)
	}
	if store == nil {
		return nil, pkgerr.ErrStoreNotFound
	}

	return store, nil
}

func (s *StoreService) GetBySellerID(ctx context.Context, sellerID int64) (*models.Store, error) {
	if sellerID <= 0 {
		return nil, fmt.Errorf("invalid seller id")
	}

	store, err := s.storeRepo.GetBySellerID(ctx, sellerID)
	if err != nil {
		return nil, fmt.Errorf("error fetching store for seller: %w", err)
	}
	if store == nil {
		return nil, pkgerr.ErrStoreNotFound
	}

	return store, nil
}

func (s *StoreService) GetAll(ctx context.Context) ([]*models.Store, error) {
	stores, err := s.storeRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching stores: %w", err)
	}
	if stores == nil {
		stores = []*models.Store{}
	}
	return stores, nil
}