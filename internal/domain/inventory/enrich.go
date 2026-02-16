package inventory

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// Enricher fills details' SuggestedValue from the previous inventory's real value per item.
// Single place for "suggested from previous" so InventoryService and AdminService stay in sync.
type Enricher struct {
	invRepo    repository.InventoryRepository
	detailRepo repository.InventoryDetailRepository
}

// NewEnricher builds an Enricher with the given repositories.
func NewEnricher(invRepo repository.InventoryRepository, detailRepo repository.InventoryDetailRepository) *Enricher {
	return &Enricher{invRepo: invRepo, detailRepo: detailRepo}
}

// Enrich overwrites each detail's SuggestedValue with the previous inventory's real value for that item.
// If there is no previous inventory or details cannot be loaded, details are left unchanged.
// Previous = real only; shrinkage does not subtract from next period's suggested.
func (e *Enricher) Enrich(ctx context.Context, inventory *entity.Inventory, details []*entity.InventoryDetail) {
	previousInv, err := e.invRepo.FindPreviousInventory(ctx, inventory.InventoryDate, inventory.InventoryType, inventory.Schedule)
	if err != nil || previousInv == nil {
		return
	}
	prevDetails, err := e.detailRepo.FindByInventoryID(ctx, previousInv.ID)
	if err != nil {
		return
	}
	realByItem := make(map[uint16]uint16)
	for _, d := range prevDetails {
		if d.RealValue != nil {
			realByItem[d.ItemID] = *d.RealValue
		}
	}
	for _, d := range details {
		if v, ok := realByItem[d.ItemID]; ok {
			d.SuggestedValue = &v
		}
	}
}
