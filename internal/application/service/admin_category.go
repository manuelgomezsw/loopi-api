package service

import (
	"context"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// AdminCategoryService handles admin category operations (CRUD and reorder).
type AdminCategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewAdminCategoryService creates a new admin category service.
func NewAdminCategoryService(categoryRepo repository.CategoryRepository) *AdminCategoryService {
	return &AdminCategoryService{categoryRepo: categoryRepo}
}

// ListCategories retrieves all categories.
func (s *AdminCategoryService) ListCategories(ctx context.Context) ([]*entity.Category, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

// GetCategory retrieves a single category by ID.
func (s *AdminCategoryService) GetCategory(ctx context.Context, id uint16) (*entity.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}
	return category, nil
}

// CreateCategory creates a new category.
func (s *AdminCategoryService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*entity.Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	existing, err := s.categoryRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check category name: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("category name already exists")
	}

	maxOrder, err := s.categoryRepo.GetMaxDisplayOrder(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get max display order: %w", err)
	}

	category := &entity.Category{
		Name:         req.Name,
		DisplayOrder: maxOrder + 1,
		Active:       true,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// UpdateCategory updates an existing category.
func (s *AdminCategoryService) UpdateCategory(ctx context.Context, id uint16, req UpdateCategoryRequest) (*entity.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	existing, err := s.categoryRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check category name: %w", err)
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("category name already exists")
	}

	category.Name = req.Name
	category.Active = req.Active

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// UpdateCategoryStatus updates the active status of a category.
func (s *AdminCategoryService) UpdateCategoryStatus(ctx context.Context, id uint16, active bool) error {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return fmt.Errorf("category not found")
	}

	if err := s.categoryRepo.UpdateStatus(ctx, id, active); err != nil {
		return fmt.Errorf("failed to update category status: %w", err)
	}

	return nil
}

// ReorderCategories updates the display order of multiple categories.
func (s *AdminCategoryService) ReorderCategories(ctx context.Context, req ReorderCategoryRequest) error {
	if len(req.Orders) == 0 {
		return nil
	}

	orders := make(map[uint16]int)
	for _, item := range req.Orders {
		orders[item.ID] = item.DisplayOrder
	}

	if err := s.categoryRepo.UpdateDisplayOrders(ctx, orders); err != nil {
		return fmt.Errorf("failed to reorder categories: %w", err)
	}

	return nil
}
