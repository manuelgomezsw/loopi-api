package service

import (
	"context"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	invdomain "github.com/manuelgomezsw/loopi-api/internal/domain/inventory"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
	"golang.org/x/sync/errgroup"
)

// InventoryDiscrepancyCounter is used by the dashboard to count items with discrepancy per inventory.
type InventoryDiscrepancyCounter interface {
	CountItemsWithDiscrepancy(ctx context.Context, inv *entity.Inventory) int
}

// AdminDashboardService handles dashboard data (stats and recent discrepancies).
type AdminDashboardService struct {
	inventoryRepo       repository.InventoryRepository
	inventoryDetailRepo repository.InventoryDetailRepository
	enricher            *invdomain.Enricher
	discrepancyCounter  InventoryDiscrepancyCounter
}

// NewAdminDashboardService creates a new admin dashboard service.
func NewAdminDashboardService(
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
	enricher *invdomain.Enricher,
	discrepancyCounter InventoryDiscrepancyCounter,
) *AdminDashboardService {
	return &AdminDashboardService{
		inventoryRepo:       inventoryRepo,
		inventoryDetailRepo: inventoryDetailRepo,
		enricher:            enricher,
		discrepancyCounter:  discrepancyCounter,
	}
}

// GetDashboard returns dashboard data (stats only).
// The three independent repo calls are executed in parallel.
func (s *AdminDashboardService) GetDashboard(ctx context.Context, days int) (*DashboardData, error) {
	today := datetime.Today()

	var todayCount int
	var pending int
	var completedToday []*entity.Inventory

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		todayCount, err = s.inventoryRepo.CountInventoriesByDate(gctx, today)
		if err != nil {
			return fmt.Errorf("failed to count today inventories: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		pending, err = s.inventoryRepo.CountInProgress(gctx)
		if err != nil {
			return fmt.Errorf("failed to count pending: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		completedToday, err = s.inventoryRepo.FindCompletedInventoriesByDate(gctx, today)
		if err != nil {
			return fmt.Errorf("failed to find completed inventories by date: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	withDiscrepancies := 0
	for _, inv := range completedToday {
		if s.discrepancyCounter.CountItemsWithDiscrepancy(ctx, inv) > 0 {
			withDiscrepancies++
		}
	}
	withoutDiscrepancies := len(completedToday) - withDiscrepancies
	if withoutDiscrepancies < 0 {
		withoutDiscrepancies = 0
	}

	return &DashboardData{
		Stats: DashboardStats{
			TodayInventories:     todayCount,
			WithDiscrepancies:    withDiscrepancies,
			WithoutDiscrepancies: withoutDiscrepancies,
			PendingInventories:   pending,
		},
	}, nil
}
