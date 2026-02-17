package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
)

// mysqlInventoryDetailRepository implements repository.InventoryDetailRepository.
type mysqlInventoryDetailRepository struct {
	db *sql.DB
}

// NewMySQLInventoryDetailRepository creates a new MySQL inventory detail repository.
func NewMySQLInventoryDetailRepository(db *sql.DB) repository.InventoryDetailRepository {
	return &mysqlInventoryDetailRepository{db: db}
}

// FindByID retrieves an inventory detail by its ID.
func (r *mysqlInventoryDetailRepository) FindByID(ctx context.Context, id uint32) (*entity.InventoryDetail, error) {
	query := `
		SELECT id, inventory_id, item_id, suggested_value, real_value, stock_received, units_sold, shrinkage, created_at, updated_at
		FROM inventory_details
		WHERE id = ?
	`

	var d entity.InventoryDetail
	var suggestedValue, realValue, stockReceived, unitsSold, shrinkage sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.InventoryID, &d.ItemID, &suggestedValue, &realValue, &stockReceived, &unitsSold, &shrinkage, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find inventory detail by id: %w", err)
	}

	if suggestedValue.Valid {
		v := uint16(suggestedValue.Int32)
		d.SuggestedValue = &v
	}
	if realValue.Valid {
		v := uint16(realValue.Int32)
		d.RealValue = &v
	}
	if stockReceived.Valid {
		v := uint16(stockReceived.Int32)
		d.StockReceived = &v
	}
	if unitsSold.Valid {
		v := uint16(unitsSold.Int32)
		d.UnitsSold = &v
	}
	if shrinkage.Valid {
		v := uint16(shrinkage.Int32)
		d.Shrinkage = &v
	}

	return &d, nil
}

// FindByInventoryID retrieves all details for an inventory.
func (r *mysqlInventoryDetailRepository) FindByInventoryID(ctx context.Context, inventoryID uint32) ([]*entity.InventoryDetail, error) {
	query := `
		SELECT id, inventory_id, item_id, suggested_value, real_value, stock_received, units_sold, shrinkage, created_at, updated_at
		FROM inventory_details
		WHERE inventory_id = ?
		ORDER BY item_id
	`

	rows, err := r.db.QueryContext(ctx, query, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to find inventory details: %w", err)
	}
	defer rows.Close()

	return r.scanDetails(rows)
}

// FindByInventoryIDWithItems retrieves all details with item information.
// Items are ordered alphabetically by item name.
func (r *mysqlInventoryDetailRepository) FindByInventoryIDWithItems(ctx context.Context, inventoryID uint32) ([]*entity.InventoryDetail, error) {
	query := `
		SELECT 
			d.id, d.inventory_id, d.item_id, d.suggested_value, d.real_value, d.stock_received, d.units_sold, d.shrinkage, d.created_at, d.updated_at,
			i.id, i.type, i.name, i.active, i.inventory_frequency, i.category_id, i.supplier_id, i.cost, i.created_at, i.updated_at,
			c.name as category_name, c.display_order
		FROM inventory_details d
		INNER JOIN items i ON d.item_id = i.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE d.inventory_id = ?
		ORDER BY i.name ASC
	`

	return r.queryDetailsWithItemsAndCategory(ctx, query, inventoryID)
}

// queryDetailsWithItems is a helper to query details joined with items.
func (r *mysqlInventoryDetailRepository) queryDetailsWithItems(ctx context.Context, query string, args ...interface{}) ([]*entity.InventoryDetail, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory details with items: %w", err)
	}
	defer rows.Close()

	var details []*entity.InventoryDetail
	for rows.Next() {
		var d entity.InventoryDetail
		var item entity.Item
		var suggestedValue, realValue, stockReceived, unitsSold sql.NullInt32

		if err := rows.Scan(
			&d.ID, &d.InventoryID, &d.ItemID, &suggestedValue, &realValue, &stockReceived, &unitsSold, &d.CreatedAt, &d.UpdatedAt,
			&item.ID, &item.Type, &item.Name, &item.Active, &item.InventoryFrequency, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan inventory detail with item: %w", err)
		}

		if suggestedValue.Valid {
			v := uint16(suggestedValue.Int32)
			d.SuggestedValue = &v
		}
		if realValue.Valid {
			v := uint16(realValue.Int32)
			d.RealValue = &v
		}
		if stockReceived.Valid {
			v := uint16(stockReceived.Int32)
			d.StockReceived = &v
		}
		if unitsSold.Valid {
			v := uint16(unitsSold.Int32)
			d.UnitsSold = &v
		}

		d.Item = &item
		details = append(details, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory details: %w", err)
	}

	return details, nil
}

// queryDetailsWithItemsAndCategory is a helper to query details joined with items and categories.
func (r *mysqlInventoryDetailRepository) queryDetailsWithItemsAndCategory(ctx context.Context, query string, args ...interface{}) ([]*entity.InventoryDetail, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory details with items and category: %w", err)
	}
	defer rows.Close()

	var details []*entity.InventoryDetail
	for rows.Next() {
		var d entity.InventoryDetail
		var item entity.Item
		var suggestedValue, realValue, stockReceived, unitsSold, shrinkage sql.NullInt32
		var supplierID sql.NullInt32
		var categoryName sql.NullString
		var displayOrder sql.NullInt32

		if err := rows.Scan(
			&d.ID, &d.InventoryID, &d.ItemID, &suggestedValue, &realValue, &stockReceived, &unitsSold, &shrinkage, &d.CreatedAt, &d.UpdatedAt,
			&item.ID, &item.Type, &item.Name, &item.Active, &item.InventoryFrequency, &item.CategoryID, &supplierID, &item.Cost, &item.CreatedAt, &item.UpdatedAt,
			&categoryName, &displayOrder,
		); err != nil {
			return nil, fmt.Errorf("failed to scan inventory detail with item and category: %w", err)
		}

		if suggestedValue.Valid {
			v := uint16(suggestedValue.Int32)
			d.SuggestedValue = &v
		}
		if realValue.Valid {
			v := uint16(realValue.Int32)
			d.RealValue = &v
		}
		if stockReceived.Valid {
			v := uint16(stockReceived.Int32)
			d.StockReceived = &v
		}
		if unitsSold.Valid {
			v := uint16(unitsSold.Int32)
			d.UnitsSold = &v
		}
		if shrinkage.Valid {
			v := uint16(shrinkage.Int32)
			d.Shrinkage = &v
		}
		if supplierID.Valid {
			v := uint16(supplierID.Int32)
			item.SupplierID = &v
		}

		// Populate category info
		if categoryName.Valid {
			item.Category = &entity.Category{
				ID:           item.CategoryID,
				Name:         categoryName.String,
				DisplayOrder: int(displayOrder.Int32),
			}
		}

		d.Item = &item
		details = append(details, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory details: %w", err)
	}

	return details, nil
}

// FindByInventoryAndItem retrieves a specific detail by inventory and item.
func (r *mysqlInventoryDetailRepository) FindByInventoryAndItem(ctx context.Context, inventoryID uint32, itemID uint16) (*entity.InventoryDetail, error) {
	query := `
		SELECT id, inventory_id, item_id, suggested_value, real_value, stock_received, units_sold, shrinkage, created_at, updated_at
		FROM inventory_details
		WHERE inventory_id = ? AND item_id = ?
	`

	var d entity.InventoryDetail
	var suggestedValue, realValue, stockReceived, unitsSold, shrinkage sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, inventoryID, itemID).Scan(
		&d.ID, &d.InventoryID, &d.ItemID, &suggestedValue, &realValue, &stockReceived, &unitsSold, &shrinkage, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find inventory detail: %w", err)
	}

	if suggestedValue.Valid {
		v := uint16(suggestedValue.Int32)
		d.SuggestedValue = &v
	}
	if realValue.Valid {
		v := uint16(realValue.Int32)
		d.RealValue = &v
	}
	if stockReceived.Valid {
		v := uint16(stockReceived.Int32)
		d.StockReceived = &v
	}
	if unitsSold.Valid {
		v := uint16(unitsSold.Int32)
		d.UnitsSold = &v
	}
	if shrinkage.Valid {
		v := uint16(shrinkage.Int32)
		d.Shrinkage = &v
	}

	return &d, nil
}

// Create creates a new inventory detail.
func (r *mysqlInventoryDetailRepository) Create(ctx context.Context, detail *entity.InventoryDetail) error {
	query := `
		INSERT INTO inventory_details (inventory_id, item_id, suggested_value, real_value, stock_received, units_sold, shrinkage)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		detail.InventoryID,
		detail.ItemID,
		detail.SuggestedValue,
		detail.RealValue,
		detail.StockReceived,
		detail.UnitsSold,
		detail.Shrinkage,
	)
	if err != nil {
		return fmt.Errorf("failed to create inventory detail: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	detail.ID = uint32(id)
	return nil
}

// Update updates an existing inventory detail.
func (r *mysqlInventoryDetailRepository) Update(ctx context.Context, detail *entity.InventoryDetail) error {
	query := `
		UPDATE inventory_details
		SET suggested_value = ?, real_value = ?, stock_received = ?, units_sold = ?, shrinkage = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		detail.SuggestedValue,
		detail.RealValue,
		detail.StockReceived,
		detail.UnitsSold,
		detail.Shrinkage,
		datetime.Now(),
		detail.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update inventory detail: %w", err)
	}

	return nil
}

// Upsert creates or updates an inventory detail.
func (r *mysqlInventoryDetailRepository) Upsert(ctx context.Context, detail *entity.InventoryDetail) error {
	now := datetime.Now()
	query := `
		INSERT INTO inventory_details (inventory_id, item_id, suggested_value, real_value, stock_received, units_sold, shrinkage, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			suggested_value = VALUES(suggested_value),
			real_value = VALUES(real_value),
			stock_received = VALUES(stock_received),
			units_sold = VALUES(units_sold),
			shrinkage = VALUES(shrinkage),
			updated_at = VALUES(updated_at)
	`

	result, err := r.db.ExecContext(ctx, query,
		detail.InventoryID,
		detail.ItemID,
		detail.SuggestedValue,
		detail.RealValue,
		detail.StockReceived,
		detail.UnitsSold,
		detail.Shrinkage,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert inventory detail: %w", err)
	}

	// Get the ID if it was an insert
	if detail.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil && id > 0 {
			detail.ID = uint32(id)
		}
	}

	return nil
}

// CreateBatch creates multiple inventory details at once.
func (r *mysqlInventoryDetailRepository) CreateBatch(ctx context.Context, details []*entity.InventoryDetail) error {
	if len(details) == 0 {
		return nil
	}

	query := `
		INSERT INTO inventory_details (inventory_id, item_id, suggested_value, real_value, stock_received, units_sold, shrinkage)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, detail := range details {
		result, err := stmt.ExecContext(ctx,
			detail.InventoryID,
			detail.ItemID,
			detail.SuggestedValue,
			detail.RealValue,
			detail.StockReceived,
			detail.UnitsSold,
			detail.Shrinkage,
		)
		if err != nil {
			return fmt.Errorf("failed to insert inventory detail: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		detail.ID = uint32(id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// scanDetails is a helper function to scan inventory detail rows.
func (r *mysqlInventoryDetailRepository) scanDetails(rows *sql.Rows) ([]*entity.InventoryDetail, error) {
	var details []*entity.InventoryDetail
	for rows.Next() {
		var d entity.InventoryDetail
		var suggestedValue, realValue, stockReceived, unitsSold, shrinkage sql.NullInt32

		if err := rows.Scan(
			&d.ID, &d.InventoryID, &d.ItemID, &suggestedValue, &realValue, &stockReceived, &unitsSold, &shrinkage, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan inventory detail: %w", err)
		}

		if suggestedValue.Valid {
			v := uint16(suggestedValue.Int32)
			d.SuggestedValue = &v
		}
		if realValue.Valid {
			v := uint16(realValue.Int32)
			d.RealValue = &v
		}
		if stockReceived.Valid {
			v := uint16(stockReceived.Int32)
			d.StockReceived = &v
		}
		if unitsSold.Valid {
			v := uint16(unitsSold.Int32)
			d.UnitsSold = &v
		}
		if shrinkage.Valid {
			v := uint16(shrinkage.Int32)
			d.Shrinkage = &v
		}

		details = append(details, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory details: %w", err)
	}

	return details, nil
}
