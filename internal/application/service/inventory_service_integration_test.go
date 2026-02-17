package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	invdomain "github.com/manuelgomezsw/loopi-api/internal/domain/inventory"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// fake inventory repo for GetDiscrepancies / CompleteInventory tests
type fakeInvRepoForDiscrepancy struct {
	inv            *entity.Inventory
	completeCalled bool
}

func (f *fakeInvRepoForDiscrepancy) FindByID(ctx context.Context, id uint32) (*entity.Inventory, error) {
	if f.inv != nil && f.inv.ID == id {
		return f.inv, nil
	}
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindByIDWithEmployee(ctx context.Context, id uint32) (*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindByDateTypeAndSchedule(ctx context.Context, date time.Time, inventoryType entity.InventoryType, schedule *entity.Schedule) (*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindLatestCompleted(ctx context.Context) (*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindInProgressByEmployee(ctx context.Context, employeeID uint16) ([]*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindLatestByType(ctx context.Context, inventoryType entity.InventoryType) (*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindPreviousInventory(ctx context.Context, date time.Time, inventoryType entity.InventoryType, schedule *entity.Schedule) (*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindAllWithFilters(ctx context.Context, dateFrom, dateTo *time.Time, inventoryType *entity.InventoryType, employeeID *uint16, page, pageSize int) ([]*entity.Inventory, int, error) {
	return nil, 0, nil
}
func (f *fakeInvRepoForDiscrepancy) CountInventoriesByDate(ctx context.Context, date time.Time) (int, error) {
	return 0, nil
}
func (f *fakeInvRepoForDiscrepancy) FindCompletedInventoriesByDate(ctx context.Context, date time.Time) ([]*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) FindAllInProgress(ctx context.Context) ([]*entity.Inventory, error) {
	return nil, nil
}
func (f *fakeInvRepoForDiscrepancy) CountInProgress(ctx context.Context) (int, error) {
	return 0, nil
}
func (f *fakeInvRepoForDiscrepancy) Create(ctx context.Context, inventory *entity.Inventory) error {
	return nil
}
func (f *fakeInvRepoForDiscrepancy) Update(ctx context.Context, inventory *entity.Inventory) error {
	return nil
}
func (f *fakeInvRepoForDiscrepancy) Complete(ctx context.Context, id uint32) error {
	f.completeCalled = true
	return nil
}

func ptrU16(v uint16) *uint16 { return &v }

// fake detail repo returning fixed details (with items for FindByInventoryIDWithItems)
type fakeDetailRepoForDiscrepancy struct {
	details []*entity.InventoryDetail
}

func (f *fakeDetailRepoForDiscrepancy) FindByID(ctx context.Context, id uint32) (*entity.InventoryDetail, error) {
	return nil, nil
}
func (f *fakeDetailRepoForDiscrepancy) FindByInventoryID(ctx context.Context, inventoryID uint32) ([]*entity.InventoryDetail, error) {
	return f.details, nil
}
func (f *fakeDetailRepoForDiscrepancy) FindByInventoryIDWithItems(ctx context.Context, inventoryID uint32) ([]*entity.InventoryDetail, error) {
	return f.details, nil
}
func (f *fakeDetailRepoForDiscrepancy) FindByInventoryAndItem(ctx context.Context, inventoryID uint32, itemID uint16) (*entity.InventoryDetail, error) {
	return nil, nil
}
func (f *fakeDetailRepoForDiscrepancy) Create(ctx context.Context, detail *entity.InventoryDetail) error {
	return nil
}
func (f *fakeDetailRepoForDiscrepancy) Update(ctx context.Context, detail *entity.InventoryDetail) error {
	return nil
}
func (f *fakeDetailRepoForDiscrepancy) Upsert(ctx context.Context, detail *entity.InventoryDetail) error {
	return nil
}
func (f *fakeDetailRepoForDiscrepancy) CreateBatch(ctx context.Context, details []*entity.InventoryDetail) error {
	return nil
}

// minimal item repo for service constructor
type fakeItemRepo struct{}

func (fakeItemRepo) FindByID(ctx context.Context, id uint16) (*entity.Item, error) { return nil, nil }
func (fakeItemRepo) FindAllActive(ctx context.Context) ([]*entity.Item, error)     { return nil, nil }
func (fakeItemRepo) FindActiveByType(ctx context.Context, itemType entity.ItemType) ([]*entity.Item, error) {
	return nil, nil
}
func (fakeItemRepo) FindActiveByInventoryType(ctx context.Context, inventoryType entity.InventoryType) ([]*entity.Item, error) {
	return nil, nil
}
func (fakeItemRepo) FindAllWithFilters(ctx context.Context, itemType *entity.ItemType, frequency *entity.InventoryFrequency, active *bool, search string, page, pageSize int) ([]*entity.Item, int, error) {
	return nil, 0, nil
}
func (fakeItemRepo) Create(ctx context.Context, item *entity.Item) error            { return nil }
func (fakeItemRepo) Update(ctx context.Context, item *entity.Item) error            { return nil }
func (fakeItemRepo) UpdateStatus(ctx context.Context, id uint16, active bool) error { return nil }

var _ repository.InventoryRepository = (*fakeInvRepoForDiscrepancy)(nil)
var _ repository.InventoryDetailRepository = (*fakeDetailRepoForDiscrepancy)(nil)
var _ repository.ItemRepository = (*fakeItemRepo)(nil)

// TestGetDiscrepancies_ReturnsOnlyDetailsWithRealNeExpected verifies that GetDiscrepancies
// returns only items where real_value != expected_at_end (domain rule), and that the count
// matches what CompleteInventory would return for the same data.
func TestGetDiscrepancies_ReturnsOnlyDetailsWithRealNeExpected(t *testing.T) {
	// Same fixture as domain consistency test: 2 with discrepancy (items 3 and 4).
	inv := &entity.Inventory{ID: 1, Status: entity.InventoryStatusInProgress, InventoryType: entity.InventoryTypeDaily}
	details := []*entity.InventoryDetail{
		{ItemID: 1, SuggestedValue: ptrU16(10), RealValue: ptrU16(10)},
		{ItemID: 2, SuggestedValue: ptrU16(42), UnitsSold: ptrU16(3), RealValue: ptrU16(39)},
		{ItemID: 3, SuggestedValue: ptrU16(20), StockReceived: ptrU16(5), UnitsSold: ptrU16(2), RealValue: ptrU16(22)},
		{ItemID: 4, SuggestedValue: ptrU16(8), RealValue: ptrU16(12)},
		{ItemID: 5, SuggestedValue: ptrU16(0), RealValue: ptrU16(0)},
	}
	// Ensure items have Name for any code path that might read it
	for i, d := range details {
		d.Item = &entity.Item{ID: d.ItemID, Name: fmt.Sprintf("item-%d", i)}
	}

	invRepo := &fakeInvRepoForDiscrepancy{inv: inv}
	detailRepo := &fakeDetailRepoForDiscrepancy{details: details}
	enricher := invdomain.NewEnricher(invRepo, detailRepo)
	svc := NewInventoryService(invRepo, detailRepo, fakeItemRepo{}, enricher)
	ctx := context.Background()

	gotInv, gotDetails, err := svc.GetDiscrepancies(ctx, 1)
	if err != nil {
		t.Fatalf("GetDiscrepancies: %v", err)
	}
	if gotInv == nil || gotInv.ID != 1 {
		t.Errorf("expected inventory id 1, got %v", gotInv)
	}
	// Expect 2 discrepancies (items 3 and 4)
	if len(gotDetails) != 2 {
		t.Errorf("expected 2 discrepancy details, got %d", len(gotDetails))
	}
	for _, d := range gotDetails {
		expected := invdomain.ExpectedAtEnd(d)
		if !invdomain.HasDiscrepancyFromExpectedEnd(d, expected) {
			t.Errorf("detail item %d: should have discrepancy", d.ItemID)
		}
	}
}

// TestCompleteInventory_DiscrepancyCountMatchesDomain verifies that the discrepancy count
// returned by CompleteInventory is computed with the same domain rule (expected_at_end, has_discrepancy).
func TestCompleteInventory_DiscrepancyCountMatchesDomain(t *testing.T) {
	inv := &entity.Inventory{ID: 1, Status: entity.InventoryStatusInProgress, InventoryType: entity.InventoryTypeDaily}
	details := []*entity.InventoryDetail{
		{ItemID: 1, SuggestedValue: ptrU16(10), RealValue: ptrU16(10)},
		{ItemID: 2, SuggestedValue: ptrU16(42), UnitsSold: ptrU16(3), RealValue: ptrU16(39)},
		{ItemID: 3, SuggestedValue: ptrU16(20), StockReceived: ptrU16(5), UnitsSold: ptrU16(2), RealValue: ptrU16(22)},
		{ItemID: 4, SuggestedValue: ptrU16(8), RealValue: ptrU16(12)},
	}
	invRepo := &fakeInvRepoForDiscrepancy{inv: inv}
	detailRepo := &fakeDetailRepoForDiscrepancy{details: details}
	enricher := invdomain.NewEnricher(invRepo, detailRepo)
	svc := NewInventoryService(invRepo, detailRepo, fakeItemRepo{}, enricher)
	ctx := context.Background()

	count, err := svc.CompleteInventory(ctx, 1)
	if err != nil {
		t.Fatalf("CompleteInventory: %v", err)
	}
	if count != 2 {
		t.Errorf("expected discrepancy count 2, got %d", count)
	}
	if !invRepo.completeCalled {
		t.Error("Complete was not called on repo")
	}
}
