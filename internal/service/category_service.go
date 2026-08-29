package service

import (
	"context"
	"fmt"
	"strings"

	"shop/internal/models"
	"shop/internal/repository"
	pkgerr "shop/pkg/errors"
)

type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, name string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("category name cannot be empty")
	}
	category := &models.Category{Name: name}
	err := s.categoryRepo.Create(ctx, category)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid category id")
	}
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, pkgerr.ErrCategoryNotFound
	}
	return category, nil
}

func (s *CategoryService) GetAll(ctx context.Context) ([]*models.Category, error) {
	return s.categoryRepo.GetAll(ctx)
}