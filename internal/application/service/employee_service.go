package service

import (
	"context"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// EmployeeService is the facade for employee (operativo) operations.
// It delegates to domain-specific services; today only inventory, ready for future growth.
type EmployeeService struct {
	inventory *InventoryService
}

// NewEmployeeService creates a new employee service.
func NewEmployeeService(inventory *InventoryService) *EmployeeService {
	return &EmployeeService{inventory: inventory}
}

// GetSuggestedSchedule delegates to InventoryService.
func (s *EmployeeService) GetSuggestedSchedule() SuggestedInventory {
	return s.inventory.GetSuggestedSchedule()
}

// GetLatestCompleted delegates to InventoryService.
func (s *EmployeeService) GetLatestCompleted(ctx context.Context) (*entity.Inventory, error) {
	return s.inventory.GetLatestCompleted(ctx)
}

// GetInProgress delegates to InventoryService.
func (s *EmployeeService) GetInProgress(ctx context.Context, employeeID uint16) ([]*entity.Inventory, error) {
	return s.inventory.GetInProgress(ctx, employeeID)
}

// CreateInventory delegates to InventoryService.
func (s *EmployeeService) CreateInventory(ctx context.Context, inventoryType entity.InventoryType, schedule *entity.Schedule, date time.Time, responsibleID uint16) (*entity.Inventory, error) {
	return s.inventory.CreateInventory(ctx, inventoryType, schedule, date, responsibleID)
}

// GetInventoryItems delegates to InventoryService.
func (s *EmployeeService) GetInventoryItems(ctx context.Context, inventoryID uint32) (*entity.Inventory, []*entity.InventoryDetail, error) {
	return s.inventory.GetInventoryItems(ctx, inventoryID)
}

// ComputeExpectedAtEnd delegates to InventoryService.
func (s *EmployeeService) ComputeExpectedAtEnd(d *entity.InventoryDetail) uint16 {
	return s.inventory.ComputeExpectedAtEnd(d)
}

// HasDiscrepancyFromExpectedEnd delegates to InventoryService.
func (s *EmployeeService) HasDiscrepancyFromExpectedEnd(d *entity.InventoryDetail) bool {
	return s.inventory.HasDiscrepancyFromExpectedEnd(d)
}

// GetDiscrepancies delegates to InventoryService.
func (s *EmployeeService) GetDiscrepancies(ctx context.Context, inventoryID uint32) (*entity.Inventory, []*entity.InventoryDetail, error) {
	return s.inventory.GetDiscrepancies(ctx, inventoryID)
}

// SaveInventoryDetail delegates to InventoryService.
func (s *EmployeeService) SaveInventoryDetail(ctx context.Context, inventoryID uint32, itemID uint16, realValue uint16) (*entity.InventoryDetail, error) {
	return s.inventory.SaveInventoryDetail(ctx, inventoryID, itemID, realValue)
}

// SaveSalesAndPurchases delegates to InventoryService.
func (s *EmployeeService) SaveSalesAndPurchases(ctx context.Context, inventoryID uint32, itemID uint16, stockReceived, unitsSold *uint16) (*entity.InventoryDetail, error) {
	return s.inventory.SaveSalesAndPurchases(ctx, inventoryID, itemID, stockReceived, unitsSold)
}

// GetInventorySummary delegates to InventoryService.
func (s *EmployeeService) GetInventorySummary(ctx context.Context, inventoryID uint32) (*entity.Inventory, []*entity.InventoryDetail, error) {
	return s.inventory.GetInventorySummary(ctx, inventoryID)
}

// CompleteInventory delegates to InventoryService.
func (s *EmployeeService) CompleteInventory(ctx context.Context, inventoryID uint32) (int, error) {
	return s.inventory.CompleteInventory(ctx, inventoryID)
}
