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

// FindAllWithFilters retrieves items with optional filters and pagination.
func (r *mysqlItemRepository) FindAllWithFilters(ctx context.Context, itemType *entity.ItemType, frequency *entity.InventoryFrequency, active *bool, search string, page, pageSize int) ([]*entity.Item, int, error) {
	baseQuery := " FROM items WHERE 1=1"
	var args []interface{}

	if itemType != nil {
		baseQuery += " AND type = ?"
		args = append(args, *itemType)
	}
	if frequency != nil {
		baseQuery += " AND inventory_frequency = ?"
		args = append(args, *frequency)
	}
	if active != nil {
		baseQuery += " AND active = ?"
		args = append(args, *active)
	}
	if search != "" {
		baseQuery += " AND name LIKE ?"
		args = append(args, "%"+search+"%")
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	// Get paginated results
	selectQuery := "SELECT id, type, name, active, inventory_frequency, created_at, updated_at" + baseQuery + " ORDER BY type, name LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query items: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	return items, total, nil
}

// Create creates a new item.
func (r *mysqlItemRepository) Create(ctx context.Context, item *entity.Item) error {
	query := `
		INSERT INTO items (type, name, active, inventory_frequency)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, item.Type, item.Name, item.Active, item.InventoryFrequency)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	item.ID = uint16(id)
	return nil
}

// Update updates an existing item.
func (r *mysqlItemRepository) Update(ctx context.Context, item *entity.Item) error {
	query := `
		UPDATE items
		SET type = ?, name = ?, active = ?, inventory_frequency = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, item.Type, item.Name, item.Active, item.InventoryFrequency, item.ID)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}

// UpdateStatus updates the active status of an item.
func (r *mysqlItemRepository) UpdateStatus(ctx context.Context, id uint16, active bool) error {
	query := `
		UPDATE items
		SET active = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, active, id)
	if err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	return nil
}
