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
		SELECT id, type, name, active, created_at, updated_at
		FROM items
		WHERE id = ?
	`

	var item entity.Item
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Type,
		&item.Name,
		&item.Active,
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
		SELECT id, type, name, active, created_at, updated_at
		FROM items
		WHERE active = 1
		ORDER BY type, name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find active items: %w", err)
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

// FindActiveByType retrieves all active items of a specific type.
func (r *mysqlItemRepository) FindActiveByType(ctx context.Context, itemType entity.ItemType) ([]*entity.Item, error) {
	query := `
		SELECT id, type, name, active, created_at, updated_at
		FROM items
		WHERE active = 1 AND type = ?
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, itemType)
	if err != nil {
		return nil, fmt.Errorf("failed to find items by type: %w", err)
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
