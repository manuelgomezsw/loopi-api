package service

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	invdomain "github.com/manuelgomezsw/loopi-api/internal/domain/inventory"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// AdminService is the facade for all admin operations. It delegates to domain-specific services.
type AdminService struct {
	dashboard *AdminDashboardService
	inventory *AdminInventoryService
	item      *AdminItemService
	employee  *AdminEmployeeService
	category  *AdminCategoryService
	supplier  *AdminSupplierService
}

// NewAdminService creates a new admin service (facade) and its domain services.
func NewAdminService(
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
	employeeRepo repository.EmployeeRepository,
	itemRepo repository.ItemRepository,
	measurementUnitRepo repository.MeasurementUnitRepository,
	categoryRepo repository.CategoryRepository,
	supplierRepo repository.SupplierRepository,
	enricher *invdomain.Enricher,
) *AdminService {
	inventorySvc := NewAdminInventoryService(inventoryRepo, inventoryDetailRepo, employeeRepo, itemRepo, enricher)
	dashboardSvc := NewAdminDashboardService(inventoryRepo, inventoryDetailRepo, enricher, inventorySvc)
	itemSvc := NewAdminItemService(itemRepo, measurementUnitRepo, inventoryRepo, inventoryDetailRepo)
	employeeSvc := NewAdminEmployeeService(employeeRepo)
	categorySvc := NewAdminCategoryService(categoryRepo)
	supplierSvc := NewAdminSupplierService(supplierRepo)

	return &AdminService{
		dashboard: dashboardSvc,
		inventory: inventorySvc,
		item:      itemSvc,
		employee:  employeeSvc,
		category:  categorySvc,
		supplier:  supplierSvc,
	}
}

// GetDashboard delegates to AdminDashboardService.
func (s *AdminService) GetDashboard(ctx context.Context, days int) (*DashboardData, error) {
	return s.dashboard.GetDashboard(ctx, days)
}

// ListInventories delegates to AdminInventoryService.
func (s *AdminService) ListInventories(ctx context.Context, filter InventoryFilter) (*InventoryListResult, error) {
	return s.inventory.ListInventories(ctx, filter)
}

// GetInventoryDetail delegates to AdminInventoryService.
func (s *AdminService) GetInventoryDetail(ctx context.Context, inventoryID uint32) (*InventoryDetailView, error) {
	return s.inventory.GetInventoryDetail(ctx, inventoryID)
}

// UpdateInventoryDetail delegates to AdminInventoryService.
func (s *AdminService) UpdateInventoryDetail(ctx context.Context, inventoryID uint32, detailID uint32, suggestedValue, realValue, stockReceived, unitsSold, shrinkage *uint16) error {
	return s.inventory.UpdateInventoryDetail(ctx, inventoryID, detailID, suggestedValue, realValue, stockReceived, unitsSold, shrinkage)
}

// GetActiveInventoriesCount delegates to AdminInventoryService.
func (s *AdminService) GetActiveInventoriesCount(ctx context.Context) (int, error) {
	return s.inventory.GetActiveInventoriesCount(ctx)
}

// CreateInitialInventory delegates to AdminInventoryService.
func (s *AdminService) CreateInitialInventory(ctx context.Context, responsibleID uint16) (*entity.Inventory, error) {
	return s.inventory.CreateInitialInventory(ctx, responsibleID)
}

// ListItems delegates to AdminItemService.
func (s *AdminService) ListItems(ctx context.Context, filter ItemFilter) (*ItemListResult, error) {
	return s.item.ListItems(ctx, filter)
}

// GetItem delegates to AdminItemService.
func (s *AdminService) GetItem(ctx context.Context, id uint16) (*entity.Item, error) {
	return s.item.GetItem(ctx, id)
}

// CreateItem delegates to AdminItemService.
func (s *AdminService) CreateItem(ctx context.Context, req CreateItemRequest) (*entity.Item, error) {
	return s.item.CreateItem(ctx, req)
}

// UpdateItem delegates to AdminItemService.
func (s *AdminService) UpdateItem(ctx context.Context, id uint16, req UpdateItemRequest) (*entity.Item, error) {
	return s.item.UpdateItem(ctx, id, req)
}

// UpdateItemStatus delegates to AdminItemService.
func (s *AdminService) UpdateItemStatus(ctx context.Context, id uint16, active bool) error {
	return s.item.UpdateItemStatus(ctx, id, active)
}

// ListMeasurementUnits delegates to AdminItemService.
func (s *AdminService) ListMeasurementUnits(ctx context.Context) ([]*entity.MeasurementUnit, error) {
	return s.item.ListMeasurementUnits(ctx)
}

// ListEmployees delegates to AdminEmployeeService.
func (s *AdminService) ListEmployees(ctx context.Context, filter EmployeeFilter) (*EmployeeListResult, error) {
	return s.employee.ListEmployees(ctx, filter)
}

// GetEmployee delegates to AdminEmployeeService.
func (s *AdminService) GetEmployee(ctx context.Context, id uint16) (*entity.Employee, error) {
	return s.employee.GetEmployee(ctx, id)
}

// CreateEmployee delegates to AdminEmployeeService.
func (s *AdminService) CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (*entity.Employee, error) {
	return s.employee.CreateEmployee(ctx, req)
}

// UpdateEmployee delegates to AdminEmployeeService.
func (s *AdminService) UpdateEmployee(ctx context.Context, id uint16, req UpdateEmployeeRequest) (*entity.Employee, error) {
	return s.employee.UpdateEmployee(ctx, id, req)
}

// UpdateEmployeeStatus delegates to AdminEmployeeService.
func (s *AdminService) UpdateEmployeeStatus(ctx context.Context, id uint16, active bool) error {
	return s.employee.UpdateEmployeeStatus(ctx, id, active)
}

// ResetEmployeePassword delegates to AdminEmployeeService.
func (s *AdminService) ResetEmployeePassword(ctx context.Context, id uint16) error {
	return s.employee.ResetEmployeePassword(ctx, id)
}

// ListAllActiveEmployees delegates to AdminEmployeeService.
func (s *AdminService) ListAllActiveEmployees(ctx context.Context) ([]*entity.Employee, error) {
	return s.employee.ListAllActiveEmployees(ctx)
}

// ListCategories delegates to AdminCategoryService.
func (s *AdminService) ListCategories(ctx context.Context) ([]*entity.Category, error) {
	return s.category.ListCategories(ctx)
}

// GetCategory delegates to AdminCategoryService.
func (s *AdminService) GetCategory(ctx context.Context, id uint16) (*entity.Category, error) {
	return s.category.GetCategory(ctx, id)
}

// CreateCategory delegates to AdminCategoryService.
func (s *AdminService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*entity.Category, error) {
	return s.category.CreateCategory(ctx, req)
}

// UpdateCategory delegates to AdminCategoryService.
func (s *AdminService) UpdateCategory(ctx context.Context, id uint16, req UpdateCategoryRequest) (*entity.Category, error) {
	return s.category.UpdateCategory(ctx, id, req)
}

// UpdateCategoryStatus delegates to AdminCategoryService.
func (s *AdminService) UpdateCategoryStatus(ctx context.Context, id uint16, active bool) error {
	return s.category.UpdateCategoryStatus(ctx, id, active)
}

// ReorderCategories delegates to AdminCategoryService.
func (s *AdminService) ReorderCategories(ctx context.Context, req ReorderCategoryRequest) error {
	return s.category.ReorderCategories(ctx, req)
}

// ListSuppliers delegates to AdminSupplierService.
func (s *AdminService) ListSuppliers(ctx context.Context, filter SupplierFilter) (*SupplierListResult, error) {
	return s.supplier.ListSuppliers(ctx, filter)
}

// ListAllActiveSuppliers delegates to AdminSupplierService.
func (s *AdminService) ListAllActiveSuppliers(ctx context.Context) ([]*entity.Supplier, error) {
	return s.supplier.ListAllActiveSuppliers(ctx)
}

// GetSupplier delegates to AdminSupplierService.
func (s *AdminService) GetSupplier(ctx context.Context, id uint16) (*entity.Supplier, error) {
	return s.supplier.GetSupplier(ctx, id)
}

// CreateSupplier delegates to AdminSupplierService.
func (s *AdminService) CreateSupplier(ctx context.Context, req CreateSupplierRequest) (*entity.Supplier, error) {
	return s.supplier.CreateSupplier(ctx, req)
}

// UpdateSupplier delegates to AdminSupplierService.
func (s *AdminService) UpdateSupplier(ctx context.Context, id uint16, req UpdateSupplierRequest) (*entity.Supplier, error) {
	return s.supplier.UpdateSupplier(ctx, id, req)
}

// UpdateSupplierStatus delegates to AdminSupplierService.
func (s *AdminService) UpdateSupplierStatus(ctx context.Context, id uint16, active bool) error {
	return s.supplier.UpdateSupplierStatus(ctx, id, active)
}

// Dashboard returns the dashboard service (for handler injection).
func (s *AdminService) Dashboard() *AdminDashboardService { return s.dashboard }

// Inventory returns the admin inventory service (for handler injection).
func (s *AdminService) Inventory() *AdminInventoryService { return s.inventory }

// Item returns the admin item service (for handler injection).
func (s *AdminService) Item() *AdminItemService { return s.item }

// Employee returns the admin employee service (for handler injection).
func (s *AdminService) Employee() *AdminEmployeeService { return s.employee }

// Category returns the admin category service (for handler injection).
func (s *AdminService) Category() *AdminCategoryService { return s.category }

// Supplier returns the admin supplier service (for handler injection).
func (s *AdminService) Supplier() *AdminSupplierService { return s.supplier }
