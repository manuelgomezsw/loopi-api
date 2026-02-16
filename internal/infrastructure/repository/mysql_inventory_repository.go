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

// FindInProgressByEmployee retrieves in-progress inventories for an employee.
func (r *mysqlInventoryRepository) FindInProgressByEmployee(ctx context.Context, employeeID uint16) ([]*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE status = 'in_progress' AND responsible_id = ?
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find in-progress inventories: %w", err)
	}
	defer rows.Close()

	var inventories []*entity.Inventory
	for rows.Next() {
		var inv entity.Inventory
		var completedAt sql.NullTime
		var schedule sql.NullString

		err := rows.Scan(
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory: %w", err)
		}

		if completedAt.Valid {
			inv.CompletedAt = &completedAt.Time
		}
		if schedule.Valid {
			s := entity.Schedule(schedule.String)
			inv.Schedule = &s
		}
		inventories = append(inventories, &inv)
	}

	return inventories, nil
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
// If no previous inventory of the same type is found, it falls back to the latest completed initial inventory.
func (r *mysqlInventoryRepository) FindPreviousInventory(ctx context.Context, date time.Time, inventoryType entity.InventoryType, schedule *entity.Schedule) (*entity.Inventory, error) {
	// Initial inventories don't have a previous inventory
	if inventoryType == entity.InventoryTypeInitial {
		return nil, nil
	}

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
			// For closing, get the noon from the same day; if no noon, use opening same day (many sites only do opening + closing)
			dateStr := date.Format("2006-01-02")
			prevInv, prevErr := r.scanInventory(r.db.QueryRowContext(ctx, `
				SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
				FROM inventories
				WHERE inventory_date = ? AND inventory_type = 'daily' AND schedule = 'noon' AND status = 'completed'
			`, dateStr))
			if prevErr != nil {
				return nil, fmt.Errorf("failed to find previous inventory: %w", prevErr)
			}
			if prevInv == nil {
				prevInv, prevErr = r.scanInventory(r.db.QueryRowContext(ctx, `
					SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
					FROM inventories
					WHERE inventory_date = ? AND inventory_type = 'daily' AND schedule = 'opening' AND status = 'completed'
				`, dateStr))
				if prevErr != nil {
					return nil, fmt.Errorf("failed to find previous inventory (opening): %w", prevErr)
				}
			}
			if prevInv == nil {
				prevInv, prevErr = r.findLatestInitialInventory(ctx)
				if prevErr != nil {
					return nil, fmt.Errorf("failed to find initial inventory fallback: %w", prevErr)
				}
			}
			return prevInv, nil

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

	// If no previous inventory of the same type was found, fall back to the latest completed initial inventory
	if inv == nil {
		inv, err = r.findLatestInitialInventory(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to find initial inventory fallback: %w", err)
		}
	}

	return inv, nil
}

// findLatestInitialInventory retrieves the latest completed initial (baseline) inventory.
func (r *mysqlInventoryRepository) findLatestInitialInventory(ctx context.Context) (*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE inventory_type = 'initial' AND status = 'completed'
		ORDER BY completed_at DESC
		LIMIT 1
	`

	inv, err := r.scanInventory(r.db.QueryRowContext(ctx, query))
	if err != nil {
		return nil, fmt.Errorf("failed to find latest initial inventory: %w", err)
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

// FindByIDWithEmployee retrieves an inventory by its ID including employee data.
func (r *mysqlInventoryRepository) FindByIDWithEmployee(ctx context.Context, id uint32) (*entity.Inventory, error) {
	query := `
		SELECT i.id, i.inventory_date, i.inventory_type, i.schedule, i.status, i.responsible_id, i.started_at, i.completed_at, i.created_at,
		       e.id, e.name, e.last_name
		FROM inventories i
		JOIN employees e ON i.responsible_id = e.id
		WHERE i.id = ?
	`

	var inv entity.Inventory
	var emp entity.Employee
	var completedAt sql.NullTime
	var schedule sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inv.ID, &inv.InventoryDate, &inv.InventoryType, &schedule, &inv.Status, &inv.ResponsibleID, &inv.StartedAt, &completedAt, &inv.CreatedAt,
		&emp.ID, &emp.Name, &emp.LastName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find inventory with employee: %w", err)
	}

	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}
	if schedule.Valid {
		s := entity.Schedule(schedule.String)
		inv.Schedule = &s
	}
	inv.Employee = &emp

	return &inv, nil
}

// CountInventoriesByDate returns the number of inventories for the given date.
func (r *mysqlInventoryRepository) CountInventoriesByDate(ctx context.Context, date time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM inventories WHERE inventory_date = ?
	`, date.Format("2006-01-02")).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count inventories by date: %w", err)
	}
	return count, nil
}

// FindCompletedInventoriesByDate returns completed inventories for the given date (minimal fields).
func (r *mysqlInventoryRepository) FindCompletedInventoriesByDate(ctx context.Context, date time.Time) ([]*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE inventory_date = ? AND status = 'completed'
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to find completed inventories by date: %w", err)
	}
	defer rows.Close()

	var inventories []*entity.Inventory
	for rows.Next() {
		var inv entity.Inventory
		var completedAt sql.NullTime
		var schedule sql.NullString
		if err := rows.Scan(
			&inv.ID, &inv.InventoryDate, &inv.InventoryType, &schedule, &inv.Status, &inv.ResponsibleID, &inv.StartedAt, &completedAt, &inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan inventory: %w", err)
		}
		if completedAt.Valid {
			inv.CompletedAt = &completedAt.Time
		}
		if schedule.Valid {
			s := entity.Schedule(schedule.String)
			inv.Schedule = &s
		}
		inventories = append(inventories, &inv)
	}
	return inventories, rows.Err()
}

// FindAllWithFilters retrieves inventories with optional filters and pagination (data only; no discrepancy logic).
func (r *mysqlInventoryRepository) FindAllWithFilters(ctx context.Context, dateFrom, dateTo *time.Time, inventoryType *entity.InventoryType, employeeID *uint16, page, pageSize int) ([]*entity.Inventory, int, error) {
	baseQuery := `
		FROM inventories i
		JOIN employees e ON i.responsible_id = e.id
		WHERE 1=1
	`

	var args []interface{}
	var conditions string

	if dateFrom != nil {
		conditions += " AND i.inventory_date >= ?"
		args = append(args, dateFrom.Format("2006-01-02"))
	}
	if dateTo != nil {
		conditions += " AND i.inventory_date <= ?"
		args = append(args, dateTo.Format("2006-01-02"))
	}
	if inventoryType != nil {
		conditions += " AND i.inventory_type = ?"
		args = append(args, *inventoryType)
	}
	if employeeID != nil {
		conditions += " AND i.responsible_id = ?"
		args = append(args, *employeeID)
	}

	countQuery := "SELECT COUNT(*) " + baseQuery + conditions
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count inventories: %w", err)
	}

	selectQuery := `
		SELECT i.id, i.inventory_date, i.inventory_type, i.schedule, i.status, i.responsible_id, i.started_at, i.completed_at, i.created_at,
		       e.id, e.name, e.last_name
	` + baseQuery + conditions + `
		ORDER BY i.inventory_date DESC, i.started_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query inventories: %w", err)
	}
	defer rows.Close()

	var inventories []*entity.Inventory
	for rows.Next() {
		var inv entity.Inventory
		var emp entity.Employee
		var completedAt sql.NullTime
		var schedule sql.NullString

		err := rows.Scan(
			&inv.ID, &inv.InventoryDate, &inv.InventoryType, &schedule, &inv.Status, &inv.ResponsibleID, &inv.StartedAt, &completedAt, &inv.CreatedAt,
			&emp.ID, &emp.Name, &emp.LastName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan inventory: %w", err)
		}

		if completedAt.Valid {
			inv.CompletedAt = &completedAt.Time
		}
		if schedule.Valid {
			s := entity.Schedule(schedule.String)
			inv.Schedule = &s
		}
		inv.Employee = &emp
		inventories = append(inventories, &inv)
	}

	return inventories, total, nil
}

// FindAllInProgress retrieves all in-progress inventories.
func (r *mysqlInventoryRepository) FindAllInProgress(ctx context.Context) ([]*entity.Inventory, error) {
	query := `
		SELECT id, inventory_date, inventory_type, schedule, status, responsible_id, started_at, completed_at, created_at
		FROM inventories
		WHERE status = 'in_progress'
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find in-progress inventories: %w", err)
	}
	defer rows.Close()

	var inventories []*entity.Inventory
	for rows.Next() {
		var inv entity.Inventory
		var completedAt sql.NullTime
		var schedule sql.NullString

		err := rows.Scan(
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory: %w", err)
		}

		if completedAt.Valid {
			inv.CompletedAt = &completedAt.Time
		}
		if schedule.Valid {
			s := entity.Schedule(schedule.String)
			inv.Schedule = &s
		}
		inventories = append(inventories, &inv)
	}

	return inventories, nil
}

// CountInProgress returns the count of in-progress inventories.
func (r *mysqlInventoryRepository) CountInProgress(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM inventories WHERE status = 'in_progress'`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count in-progress inventories: %w", err)
	}

	return count, nil
}
