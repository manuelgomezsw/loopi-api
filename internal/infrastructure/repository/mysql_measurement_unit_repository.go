package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// mysqlMeasurementUnitRepository implements repository.MeasurementUnitRepository.
type mysqlMeasurementUnitRepository struct {
	db *sql.DB
}

// NewMySQLMeasurementUnitRepository creates a new MySQL measurement unit repository.
func NewMySQLMeasurementUnitRepository(db *sql.DB) repository.MeasurementUnitRepository {
	return &mysqlMeasurementUnitRepository{db: db}
}

// FindAll retrieves all measurement units ordered by name (alphabetical).
func (r *mysqlMeasurementUnitRepository) FindAll(ctx context.Context) ([]*entity.MeasurementUnit, error) {
	query := `
		SELECT id, code, name
		FROM measurement_units
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query measurement units: %w", err)
	}
	defer rows.Close()

	var units []*entity.MeasurementUnit
	for rows.Next() {
		var u entity.MeasurementUnit
		if err := rows.Scan(&u.ID, &u.Code, &u.Name); err != nil {
			return nil, fmt.Errorf("failed to scan measurement unit: %w", err)
		}
		units = append(units, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating measurement units: %w", err)
	}

	return units, nil
}

// FindByID retrieves a measurement unit by its ID.
func (r *mysqlMeasurementUnitRepository) FindByID(ctx context.Context, id uint16) (*entity.MeasurementUnit, error) {
	query := `SELECT id, code, name FROM measurement_units WHERE id = ?`

	var u entity.MeasurementUnit
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Code, &u.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find measurement unit by id: %w", err)
	}

	return &u, nil
}
