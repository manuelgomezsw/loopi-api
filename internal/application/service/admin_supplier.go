package service

import (
	"context"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// AdminSupplierService handles admin supplier operations (CRUD and list active).
type AdminSupplierService struct {
	supplierRepo repository.SupplierRepository
}

// NewAdminSupplierService creates a new admin supplier service.
func NewAdminSupplierService(supplierRepo repository.SupplierRepository) *AdminSupplierService {
	return &AdminSupplierService{supplierRepo: supplierRepo}
}

// ListSuppliers retrieves suppliers with filters and pagination.
func (s *AdminSupplierService) ListSuppliers(ctx context.Context, filter SupplierFilter) (*SupplierListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	suppliers, total, err := s.supplierRepo.FindAllWithFilters(ctx, filter.Active, filter.Search, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}

	totalPages := total / filter.PageSize
	if total%filter.PageSize > 0 {
		totalPages++
	}

	return &SupplierListResult{
		Suppliers:  suppliers,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListAllActiveSuppliers retrieves all active suppliers for dropdowns.
func (s *AdminSupplierService) ListAllActiveSuppliers(ctx context.Context) ([]*entity.Supplier, error) {
	suppliers, err := s.supplierRepo.FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active suppliers: %w", err)
	}
	return suppliers, nil
}

// GetSupplier retrieves a single supplier by ID.
func (s *AdminSupplierService) GetSupplier(ctx context.Context, id uint16) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	if supplier == nil {
		return nil, fmt.Errorf("supplier not found")
	}
	return supplier, nil
}

// CreateSupplier creates a new supplier.
func (s *AdminSupplierService) CreateSupplier(ctx context.Context, req CreateSupplierRequest) (*entity.Supplier, error) {
	if req.BusinessName == "" {
		return nil, fmt.Errorf("business_name is required")
	}
	if req.TaxID == "" {
		return nil, fmt.Errorf("tax_id is required")
	}

	existing, err := s.supplierRepo.FindByTaxID(ctx, req.TaxID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tax_id: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("tax_id already exists")
	}

	supplier := &entity.Supplier{
		BusinessName: req.BusinessName,
		TaxID:        req.TaxID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Active:       true,
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	return supplier, nil
}

// UpdateSupplier updates an existing supplier.
func (s *AdminSupplierService) UpdateSupplier(ctx context.Context, id uint16, req UpdateSupplierRequest) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	if supplier == nil {
		return nil, fmt.Errorf("supplier not found")
	}

	if req.BusinessName == "" {
		return nil, fmt.Errorf("business_name is required")
	}
	if req.TaxID == "" {
		return nil, fmt.Errorf("tax_id is required")
	}

	existing, err := s.supplierRepo.FindByTaxID(ctx, req.TaxID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tax_id: %w", err)
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("tax_id already exists")
	}

	supplier.BusinessName = req.BusinessName
	supplier.TaxID = req.TaxID
	supplier.ContactName = req.ContactName
	supplier.ContactPhone = req.ContactPhone
	supplier.ContactEmail = req.ContactEmail
	supplier.Active = req.Active

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	return supplier, nil
}

// UpdateSupplierStatus updates the active status of a supplier.
func (s *AdminSupplierService) UpdateSupplierStatus(ctx context.Context, id uint16, active bool) error {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get supplier: %w", err)
	}
	if supplier == nil {
		return fmt.Errorf("supplier not found")
	}

	if err := s.supplierRepo.UpdateStatus(ctx, id, active); err != nil {
		return fmt.Errorf("failed to update supplier status: %w", err)
	}

	return nil
}
