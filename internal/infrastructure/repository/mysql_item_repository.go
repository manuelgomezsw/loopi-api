package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// mysqlItemRepository implements repository.ItemRepository.
type mysqlItemRepository struct {
	db *sql.DB
}

// NewMySQLItemRepository creates a new MySQL item repository.
func NewMySQLItemRepository(db *sql.DB) repository.ItemRepository {
	return &mysqlItemRepository{db: db}
}

// FindByID retrieves an item by its ID.
func (r *mysqlItemRepository) FindByID(ctx context.Context, id uint16) (*entity.Item, error) {
	query := `
		SELECT id, type, name, active, inventory_frequency, created_at, updated_at
		FROM items
		WHERE id = ?
	`

	var item entity.Item
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Type,
		&item.Name,
		&item.Active,
		&item.InventoryFrequency,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find item by id: %w", err)
	}

	return &item, nil
}

// FindAllActive retrieves all active items.
func (r *mysqlItemRepository) FindAllActive(ctx context.Context) ([]*entity.Item, error) {
	query := `
		SELECT id, type, name, active, inventory_frequency, created_at, updated_at
		FROM items
		WHERE active = 1
		ORDER BY type, name
	`

	return r.queryItems(ctx, query)
}

// FindActiveByType retrieves all active items of a specific type.
func (r *mysqlItemRepository) FindActiveByType(ctx context.Context, itemType entity.ItemType) ([]*entity.Item, error) {
	query := `
		SELECT id, type, name, active, inventory_frequency, created_at, updated_at
		FROM items
		WHERE active = 1 AND type = ?
		ORDER BY name
	`

	return r.queryItems(ctx, query, itemType)
}

// FindActiveByInventoryType retrieves active items based on inventory type.
func (r *mysqlItemRepository) FindActiveByInventoryType(ctx context.Context, inventoryType entity.InventoryType) ([]*entity.Item, error) {
	var query string

	switch inventoryType {
	case entity.InventoryTypeDaily:
		query = `
			SELECT id, type, name, active, inventory_frequency, created_at, updated_at
			FROM items
			WHERE active = 1 AND inventory_frequency = 'daily'
			ORDER BY type, name
		`
	case entity.InventoryTypeWeekly:
		query = `
			SELECT id, type, name, active, inventory_frequency, created_at, updated_at
			FROM items
			WHERE active = 1 AND inventory_frequency IN ('daily', 'weekly')
			ORDER BY type, name
		`
	case entity.InventoryTypeMonthly:
		query = `
			SELECT id, type, name, active, inventory_frequency, created_at, updated_at
			FROM items
			WHERE active = 1
			ORDER BY type, name
		`
	default:
		return nil, fmt.Errorf("invalid inventory type: %s", inventoryType)
	}

	return r.queryItems(ctx, query)
}

// queryItems is a helper function to execute item queries.
func (r *mysqlItemRepository) queryItems(ctx context.Context, query string, args ...interface{}) ([]*entity.Item, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*entity.Item
	for rows.Next() {
		var item entity.Item
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Name,
			&item.Active,
			&item.InventoryFrequency,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return items, nil
}
