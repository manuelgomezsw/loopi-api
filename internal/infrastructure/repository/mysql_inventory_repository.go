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

// FindByID retrieves an inventory by its ID.
func (r *mysqlInventoryRepository) FindByID(ctx context.Context, id uint32) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE id = ?
	`

	var inv entity.Inventory
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inv.ID,
		&inv.InventoryDate,
		&inv.Schedule,
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
		return nil, fmt.Errorf("failed to find inventory by id: %w", err)
	}

	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}

	return &inv, nil
}

// FindByDateAndSchedule retrieves an inventory by date and schedule.
func (r *mysqlInventoryRepository) FindByDateAndSchedule(ctx context.Context, date time.Time, schedule entity.Schedule) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE inventory_date = ? AND schedule = ?
	`

	var inv entity.Inventory
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, date.Format("2006-01-02"), schedule).Scan(
		&inv.ID,
		&inv.InventoryDate,
		&inv.Schedule,
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
		return nil, fmt.Errorf("failed to find inventory by date and schedule: %w", err)
	}

	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}

	return &inv, nil
}

// FindLatestCompleted retrieves the most recent completed inventory.
func (r *mysqlInventoryRepository) FindLatestCompleted(ctx context.Context) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE status = 'completed'
		ORDER BY completed_at DESC
		LIMIT 1
	`

	var inv entity.Inventory
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query).Scan(
		&inv.ID,
		&inv.InventoryDate,
		&inv.Schedule,
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
		return nil, fmt.Errorf("failed to find latest completed inventory: %w", err)
	}

	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}

	return &inv, nil
}

// FindLatestBySchedule retrieves the most recent inventory for a specific schedule.
func (r *mysqlInventoryRepository) FindLatestBySchedule(ctx context.Context, schedule entity.Schedule) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE schedule = ? AND status = 'completed'
		ORDER BY inventory_date DESC, completed_at DESC
		LIMIT 1
	`

	var inv entity.Inventory
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, schedule).Scan(
		&inv.ID,
		&inv.InventoryDate,
		&inv.Schedule,
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
		return nil, fmt.Errorf("failed to find latest inventory by schedule: %w", err)
	}

	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}

	return &inv, nil
}

// FindPreviousSchedule retrieves the previous schedule's inventory for calculating suggested values.
func (r *mysqlInventoryRepository) FindPreviousSchedule(ctx context.Context, date time.Time, schedule entity.Schedule) (*entity.Inventory, error) {
	var query string
	var args []interface{}

	switch schedule {
	case entity.ScheduleOpening:
		// For opening, get the closing from the previous day
		previousDay := date.AddDate(0, 0, -1)
		query = `
			SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE inventory_date = ? AND schedule = 'closing' AND status = 'completed'
		`
		args = []interface{}{previousDay.Format("2006-01-02")}

	case entity.ScheduleNoon:
		// For noon, get the opening from the same day
		query = `
			SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE inventory_date = ? AND schedule = 'opening' AND status = 'completed'
		`
		args = []interface{}{date.Format("2006-01-02")}

	case entity.ScheduleClosing:
		// For closing, get the noon from the same day
		query = `
			SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE inventory_date = ? AND schedule = 'noon' AND status = 'completed'
		`
		args = []interface{}{date.Format("2006-01-02")}

	case entity.ScheduleWeekly:
		// For weekly, get the last weekly inventory
		query = `
			SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE schedule = 'weekly' AND status = 'completed' AND inventory_date < ?
			ORDER BY inventory_date DESC
			LIMIT 1
		`
		args = []interface{}{date.Format("2006-01-02")}

	case entity.ScheduleMonthly:
		// For monthly, get the last monthly inventory
		query = `
			SELECT id, inventory_date, schedule, status, responsible_id, started_at, completed_at, created_at
			FROM inventories
			WHERE schedule = 'monthly' AND status = 'completed' AND inventory_date < ?
			ORDER BY inventory_date DESC
			LIMIT 1
		`
		args = []interface{}{date.Format("2006-01-02")}

	default:
		return nil, fmt.Errorf("unknown schedule: %s", schedule)
	}

	var inv entity.Inventory
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&inv.ID,
		&inv.InventoryDate,
		&inv.Schedule,
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
		return nil, fmt.Errorf("failed to find previous schedule inventory: %w", err)
	}

	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}

	return &inv, nil
}

// Create creates a new inventory.
func (r *mysqlInventoryRepository) Create(ctx context.Context, inventory *entity.Inventory) error {
	query := `
		INSERT INTO inventories (inventory_date, schedule, status, responsible_id, started_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		inventory.InventoryDate.Format("2006-01-02"),
		inventory.Schedule,
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
		SET inventory_date = ?, schedule = ?, status = ?, responsible_id = ?, completed_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		inventory.InventoryDate.Format("2006-01-02"),
		inventory.Schedule,
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
