package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
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
		SELECT i.id, i.type, i.name, i.active, i.inventory_frequency, 
		       i.category_id, i.supplier_id, i.cost, i.created_at, i.updated_at,
		       c.name as category_name
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE i.id = ?
	`

	var item entity.Item
	var categoryName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Type,
		&item.Name,
		&item.Active,
		&item.InventoryFrequency,
		&item.CategoryID,
		&item.SupplierID,
		&item.Cost,
		&item.CreatedAt,
		&item.UpdatedAt,
		&categoryName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find item by id: %w", err)
	}

	if categoryName.Valid {
		item.Category = &entity.Category{ID: item.CategoryID, Name: categoryName.String}
	}

	return &item, nil
}

// FindAllActive retrieves all active items ordered by category then name.
func (r *mysqlItemRepository) FindAllActive(ctx context.Context) ([]*entity.Item, error) {
	query := `
		SELECT i.id, i.type, i.name, i.active, i.inventory_frequency, 
		       i.category_id, i.supplier_id, i.cost, i.created_at, i.updated_at,
		       c.name as category_name, c.display_order
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE i.active = 1
		ORDER BY c.display_order ASC, i.name ASC
	`

	return r.queryItemsWithCategory(ctx, query)
}

// FindActiveByType retrieves all active items of a specific type.
func (r *mysqlItemRepository) FindActiveByType(ctx context.Context, itemType entity.ItemType) ([]*entity.Item, error) {
	query := `
		SELECT i.id, i.type, i.name, i.active, i.inventory_frequency, 
		       i.category_id, i.supplier_id, i.cost, i.created_at, i.updated_at,
		       c.name as category_name, c.display_order
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE i.active = 1 AND i.type = ?
		ORDER BY c.display_order ASC, i.name ASC
	`

	return r.queryItemsWithCategory(ctx, query, itemType)
}

// FindActiveByInventoryType retrieves active items based on inventory type.
// Items are ordered by category display_order then by item name.
func (r *mysqlItemRepository) FindActiveByInventoryType(ctx context.Context, inventoryType entity.InventoryType) ([]*entity.Item, error) {
	var whereClause string

	switch inventoryType {
	case entity.InventoryTypeDaily:
		whereClause = "i.active = 1 AND i.inventory_frequency = 'daily'"
	case entity.InventoryTypeWeekly:
		whereClause = "i.active = 1 AND i.inventory_frequency IN ('daily', 'weekly')"
	case entity.InventoryTypeMonthly:
		whereClause = "i.active = 1"
	default:
		return nil, fmt.Errorf("invalid inventory type: %s", inventoryType)
	}

	query := fmt.Sprintf(`
		SELECT i.id, i.type, i.name, i.active, i.inventory_frequency, 
		       i.category_id, i.supplier_id, i.cost, i.created_at, i.updated_at,
		       c.name as category_name, c.display_order
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE %s
		ORDER BY c.display_order ASC, i.name ASC
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*entity.Item
	for rows.Next() {
		var item entity.Item
		var categoryName sql.NullString
		var displayOrder sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Name,
			&item.Active,
			&item.InventoryFrequency,
			&item.CategoryID,
			&item.SupplierID,
			&item.Cost,
			&item.CreatedAt,
			&item.UpdatedAt,
			&categoryName,
			&displayOrder,
		); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		if categoryName.Valid {
			item.Category = &entity.Category{
				ID:           item.CategoryID,
				Name:         categoryName.String,
				DisplayOrder: int(displayOrder.Int64),
			}
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return items, nil
}

// queryItemsWithCategory is a helper function to execute item queries with category info.
func (r *mysqlItemRepository) queryItemsWithCategory(ctx context.Context, query string, args ...interface{}) ([]*entity.Item, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*entity.Item
	for rows.Next() {
		var item entity.Item
		var categoryName sql.NullString
		var displayOrder sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Name,
			&item.Active,
			&item.InventoryFrequency,
			&item.CategoryID,
			&item.SupplierID,
			&item.Cost,
			&item.CreatedAt,
			&item.UpdatedAt,
			&categoryName,
			&displayOrder,
		); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		if categoryName.Valid {
			item.Category = &entity.Category{
				ID:           item.CategoryID,
				Name:         categoryName.String,
				DisplayOrder: int(displayOrder.Int64),
			}
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
	baseQuery := " FROM items i LEFT JOIN categories c ON i.category_id = c.id WHERE 1=1"
	var args []interface{}

	if itemType != nil {
		baseQuery += " AND i.type = ?"
		args = append(args, *itemType)
	}
	if frequency != nil {
		baseQuery += " AND i.inventory_frequency = ?"
		args = append(args, *frequency)
	}
	if active != nil {
		baseQuery += " AND i.active = ?"
		args = append(args, *active)
	}
	if search != "" {
		baseQuery += " AND i.name LIKE ?"
		args = append(args, "%"+search+"%")
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	// Get paginated results
	selectQuery := `SELECT i.id, i.type, i.name, i.active, i.inventory_frequency, 
	                       i.category_id, i.supplier_id, i.cost, i.created_at, i.updated_at,
	                       c.name as category_name` + baseQuery + ` ORDER BY c.display_order, i.name LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*entity.Item
	for rows.Next() {
		var item entity.Item
		var categoryName sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Name,
			&item.Active,
			&item.InventoryFrequency,
			&item.CategoryID,
			&item.SupplierID,
			&item.Cost,
			&item.CreatedAt,
			&item.UpdatedAt,
			&categoryName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan item: %w", err)
		}
		if categoryName.Valid {
			item.Category = &entity.Category{ID: item.CategoryID, Name: categoryName.String}
		}
		items = append(items, &item)
	}

	return items, total, nil
}

// Create creates a new item.
func (r *mysqlItemRepository) Create(ctx context.Context, item *entity.Item) error {
	query := `
		INSERT INTO items (type, name, active, inventory_frequency, category_id, supplier_id, cost)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, item.Type, item.Name, item.Active, item.InventoryFrequency, item.CategoryID, item.SupplierID, item.Cost)
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
		SET type = ?, name = ?, active = ?, inventory_frequency = ?, 
		    category_id = ?, supplier_id = ?, cost = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, item.Type, item.Name, item.Active, item.InventoryFrequency, item.CategoryID, item.SupplierID, item.Cost, datetime.Now(), item.ID)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}

// UpdateStatus updates the active status of an item.
func (r *mysqlItemRepository) UpdateStatus(ctx context.Context, id uint16, active bool) error {
	query := `
		UPDATE items
		SET active = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, active, datetime.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	return nil
}
