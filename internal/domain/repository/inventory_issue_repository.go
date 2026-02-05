package repository

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// InventoryIssueRepository defines the interface for inventory issue data access.
type InventoryIssueRepository interface {
	// FindByID retrieves an issue by its ID.
	FindByID(ctx context.Context, id uint32) (*entity.InventoryIssue, error)

	// FindOpenIssues retrieves all open issues.
	FindOpenIssues(ctx context.Context) ([]*entity.InventoryIssue, error)

	// FindByInventoryDetailID retrieves issues for a specific inventory detail.
	FindByInventoryDetailID(ctx context.Context, detailID uint32) ([]*entity.InventoryIssue, error)

	// Create creates a new inventory issue.
	Create(ctx context.Context, issue *entity.InventoryIssue) error

	// CreateBatch creates multiple inventory issues at once.
	CreateBatch(ctx context.Context, issues []*entity.InventoryIssue) error

	// Resolve marks an issue as resolved.
	Resolve(ctx context.Context, id uint32, resolvedBy uint16, notes string) error
}
