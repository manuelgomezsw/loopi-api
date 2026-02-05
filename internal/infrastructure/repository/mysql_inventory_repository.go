package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// mysqlInventoryRepository implements repository.InventoryRepository.
type mysqlInventoryRepository struct {
	db *sql.DB
}

// NewMySQLInventoryRepository creates a new MySQL inventory repository.
func NewMySQLInventoryRepository(db *sql.DB) repository.InventoryRepository {
	return &mysqlInventoryRepository{db: db}
}

// scanInventory scans a row into an Inventory entity.
func (r *mysqlInventoryRepository) scanInventory(row interface{ Scan(...interface{}) error }) (*entity.Inventory, error) {
	var inv entity.Inventory
	var completedAt sql.NullTime
	var schedule sql.NullString

	err := row.Scan(
		&inv.ID,
		&inv.InventoryDate,
		&inv.InventoryType,
		&schedule,
		&inv.Status,
		&inv.ResponsibleID,
		&inv.StartedAt,
		&completedAt,
		&inv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}
	if schedule.Valid {
		s := entity.Schedule(schedule.String)
		inv.Schedule = &s
	}

	return &inv, nil
}

// FindByID retrieves an inventory by its ID.
func (r *mysqlInventoryRepository) FindByID(ctx context.Context, id uint32) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE id = ?
	`

	inv, err := r.scanInventory(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("failed to find inventory by id: %w", err)
	}
	return inv, nil
}

// FindByDateTypeAndSchedule retrieves an inventory by date, type and schedule.
func (r *mysqlInventoryRepository) FindByDateTypeAndSchedule(ctx context.Context, date time.Time, inventoryType entity.InventoryType, schedule *entity.Schedule) (*entity.Inventory, error) {
	var query string
	var args []interface{}

	if schedule == nil {
		query = `
			SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE inventory_date = ? AND inventory_type = ? AND schedule IS NULL
		`
		args = []interface{}{date.Format("2006-01-02"), inventoryType}
	} else {
		query = `
			SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE inventory_date = ? AND inventory_type = ? AND schedule = ?
		`
		args = []interface{}{date.Format("2006-01-02"), inventoryType, *schedule}
	}

	inv, err := r.scanInventory(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("failed to find inventory by date, type and schedule: %w", err)
	}
	return inv, nil
}

// FindLatestCompleted retrieves the most recent completed inventory.
func (r *mysqlInventoryRepository) FindLatestCompleted(ctx context.Context) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE status = 'completed'
		ORDER BY completed_at DESC
		LIMIT 1
	`

	inv, err := r.scanInventory(r.db.QueryRowContext(ctx, query))
	if err != nil {
		return nil, fmt.Errorf("failed to find latest completed inventory: %w", err)
	}
	return inv, nil
}

// FindLatestByType retrieves the most recent completed inventory for a specific type.
func (r *mysqlInventoryRepository) FindLatestByType(ctx context.Context, inventoryType entity.InventoryType) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE inventory_type = ? AND status = 'completed'
		ORDER BY inventory_date DESC, completed_at DESC
		LIMIT 1
	`

	inv, err := r.scanInventory(r.db.QueryRowContext(ctx, query, inventoryType))
	if err != nil {
		return nil, fmt.Errorf("failed to find latest inventory by type: %w", err)
	}
	return inv, nil
}

// FindPreviousInventory retrieves the previous inventory for calculating suggested values.
func (r *mysqlInventoryRepository) FindPreviousInventory(ctx context.Context, date time.Time, inventoryType entity.InventoryType, schedule *entity.Schedule) (*entity.Inventory, error) {
	var query string
	var args []interface{}

	switch inventoryType {
	case entity.InventoryTypeDaily:
		if schedule == nil {
			return nil, fmt.Errorf("schedule is required for daily inventory")
		}

		switch *schedule {
		case entity.ScheduleOpening:
			// For opening, get the closing from the previous day
			previousDay := date.AddDate(0, 0, -1)
			query = `
				SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
				FROM inventories
				WHERE inventory_date = ? AND inventory_type = 'daily' AND schedule = 'closing' AND status = 'completed'
			`
			args = []interface{}{previousDay.Format("2006-01-02")}

		case entity.ScheduleNoon:
			// For noon, get the opening from the same day
			query = `
				SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
				FROM inventories
				WHERE inventory_date = ? AND inventory_type = 'daily' AND schedule = 'opening' AND status = 'completed'
			`
			args = []interface{}{date.Format("2006-01-02")}

		case entity.ScheduleClosing:
			// For closing, get the noon from the same day
			query = `
				SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
				FROM inventories
				WHERE inventory_date = ? AND inventory_type = 'daily' AND schedule = 'noon' AND status = 'completed'
			`
			args = []interface{}{date.Format("2006-01-02")}

		default:
			return nil, fmt.Errorf("unknown schedule: %s", *schedule)
		}

	case entity.InventoryTypeWeekly:
		// For weekly, get the last weekly inventory
		query = `
			SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE inventory_type = 'weekly' AND status = 'completed' AND inventory_date < ?
			ORDER BY inventory_date DESC
			LIMIT 1
		`
		args = []interface{}{date.Format("2006-01-02")}

	case entity.InventoryTypeMonthly:
		// For monthly, get the last monthly inventory
		query = `
			SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE inventory_type = 'monthly' AND status = 'completed' AND inventory_date < ?
			ORDER BY inventory_date DESC
			LIMIT 1
		`
		args = []interface{}{date.Format("2006-01-02")}

	default:
		return nil, fmt.Errorf("unknown inventory type: %s", inventoryType)
	}

	inv, err := r.scanInventory(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("failed to find previous inventory: %w", err)
	}
	return inv, nil
}

// Create creates a new inventory.
func (r *mysqlInventoryRepository) Create(ctx context.Context, inventory *entity.Inventory) error {
	query := `
		INSERT INTO inventories (inventory_date, inventory_type, schedule, status, responsible_id, started_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var schedule interface{}
	if inventory.Schedule != nil {
		schedule = *inventory.Schedule
	}

	result, err := r.db.ExecContext(ctx, query,
		inventory.InventoryDate.Format("2006-01-02"),
		inventory.InventoryType,
		schedule,
		inventory.Status,
		inventory.ResponsibleID,
		inventory.StartedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	inventory.ID = uint32(id)
	return nil
}

// Update updates an existing inventory.
func (r *mysqlInventoryRepository) Update(ctx context.Context, inventory *entity.Inventory) error {
	query := `
		UPDATE inventories
		SET inventory_date = ?, inventory_type = ?, schedule = ?, status = ?, responsible_id = ?, completed_at = ?
		WHERE id = ?
	`

	var schedule interface{}
	if inventory.Schedule != nil {
		schedule = *inventory.Schedule
	}

	_, err := r.db.ExecContext(ctx, query,
		inventory.InventoryDate.Format("2006-01-02"),
		inventory.InventoryType,
		schedule,
		inventory.Status,
		inventory.ResponsibleID,
		inventory.CompletedAt,
		inventory.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// Complete marks an inventory as completed.
func (r *mysqlInventoryRepository) Complete(ctx context.Context, id uint32) error {
	query := `
		UPDATE inventories
		SET status = 'completed', completed_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to complete inventory: %w", err)
	}

	return nil
}
