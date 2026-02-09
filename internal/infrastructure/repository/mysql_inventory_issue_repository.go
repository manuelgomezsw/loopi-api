package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
)

// mysqlInventoryIssueRepository implements repository.InventoryIssueRepository.
type mysqlInventoryIssueRepository struct {
	db *sql.DB
}

// NewMySQLInventoryIssueRepository creates a new MySQL inventory issue repository.
func NewMySQLInventoryIssueRepository(db *sql.DB) repository.InventoryIssueRepository {
	return &mysqlInventoryIssueRepository{db: db}
}

// FindByID retrieves an issue by its ID.
func (r *mysqlInventoryIssueRepository) FindByID(ctx context.Context, id uint32) (*entity.InventoryIssue, error) {
	query := `
		SELECT id, inventory_detail_id, type, expected_value, actual_value, difference, status, resolution_notes, resolved_by, resolved_at, created_at
		FROM inventory_issues
		WHERE id = ?
	`

	issue, err := r.scanIssue(r.db.QueryRowContext(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find issue by id: %w", err)
	}

	return issue, nil
}

// FindOpenIssues retrieves all open issues.
func (r *mysqlInventoryIssueRepository) FindOpenIssues(ctx context.Context) ([]*entity.InventoryIssue, error) {
	query := `
		SELECT id, inventory_detail_id, type, expected_value, actual_value, difference, status, resolution_notes, resolved_by, resolved_at, created_at
		FROM inventory_issues
		WHERE status = 'open'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find open issues: %w", err)
	}
	defer rows.Close()

	return r.scanIssues(rows)
}

// FindByInventoryDetailID retrieves issues for a specific inventory detail.
func (r *mysqlInventoryIssueRepository) FindByInventoryDetailID(ctx context.Context, detailID uint32) ([]*entity.InventoryIssue, error) {
	query := `
		SELECT id, inventory_detail_id, type, expected_value, actual_value, difference, status, resolution_notes, resolved_by, resolved_at, created_at
		FROM inventory_issues
		WHERE inventory_detail_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, detailID)
	if err != nil {
		return nil, fmt.Errorf("failed to find issues by detail id: %w", err)
	}
	defer rows.Close()

	return r.scanIssues(rows)
}

// Create creates a new inventory issue.
func (r *mysqlInventoryIssueRepository) Create(ctx context.Context, issue *entity.InventoryIssue) error {
	query := `
		INSERT INTO inventory_issues (inventory_detail_id, type, expected_value, actual_value, difference, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		issue.InventoryDetailID,
		issue.Type,
		issue.ExpectedValue,
		issue.ActualValue,
		issue.Difference,
		issue.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	issue.ID = uint32(id)
	return nil
}

// CreateBatch creates multiple inventory issues at once.
func (r *mysqlInventoryIssueRepository) CreateBatch(ctx context.Context, issues []*entity.InventoryIssue) error {
	if len(issues) == 0 {
		return nil
	}

	query := `
		INSERT INTO inventory_issues (inventory_detail_id, type, expected_value, actual_value, difference, status)
		VALUES (?, ?, ?, ?, ?, ?)
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

	for _, issue := range issues {
		result, err := stmt.ExecContext(ctx,
			issue.InventoryDetailID,
			issue.Type,
			issue.ExpectedValue,
			issue.ActualValue,
			issue.Difference,
			issue.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert issue: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		issue.ID = uint32(id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Resolve marks an issue as resolved.
func (r *mysqlInventoryIssueRepository) Resolve(ctx context.Context, id uint32, resolvedBy uint16, notes string) error {
	query := `
		UPDATE inventory_issues
		SET status = 'resolved', resolution_notes = ?, resolved_by = ?, resolved_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, notes, resolvedBy, datetime.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	return nil
}

// scanIssue scans a single issue from a row.
func (r *mysqlInventoryIssueRepository) scanIssue(row *sql.Row) (*entity.InventoryIssue, error) {
	var issue entity.InventoryIssue
	var expectedValue, actualValue sql.NullInt32
	var difference sql.NullInt32
	var resolutionNotes sql.NullString
	var resolvedBy sql.NullInt32
	var resolvedAt sql.NullTime

	err := row.Scan(
		&issue.ID,
		&issue.InventoryDetailID,
		&issue.Type,
		&expectedValue,
		&actualValue,
		&difference,
		&issue.Status,
		&resolutionNotes,
		&resolvedBy,
		&resolvedAt,
		&issue.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if expectedValue.Valid {
		v := uint16(expectedValue.Int32)
		issue.ExpectedValue = &v
	}
	if actualValue.Valid {
		v := uint16(actualValue.Int32)
		issue.ActualValue = &v
	}
	if difference.Valid {
		v := int16(difference.Int32)
		issue.Difference = &v
	}
	if resolutionNotes.Valid {
		issue.ResolutionNotes = &resolutionNotes.String
	}
	if resolvedBy.Valid {
		v := uint16(resolvedBy.Int32)
		issue.ResolvedBy = &v
	}
	if resolvedAt.Valid {
		issue.ResolvedAt = &resolvedAt.Time
	}

	return &issue, nil
}

// scanIssues scans multiple issues from rows.
func (r *mysqlInventoryIssueRepository) scanIssues(rows *sql.Rows) ([]*entity.InventoryIssue, error) {
	var issues []*entity.InventoryIssue
	for rows.Next() {
		var issue entity.InventoryIssue
		var expectedValue, actualValue sql.NullInt32
		var difference sql.NullInt32
		var resolutionNotes sql.NullString
		var resolvedBy sql.NullInt32
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&issue.ID,
			&issue.InventoryDetailID,
			&issue.Type,
			&expectedValue,
			&actualValue,
			&difference,
			&issue.Status,
			&resolutionNotes,
			&resolvedBy,
			&resolvedAt,
			&issue.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		if expectedValue.Valid {
			v := uint16(expectedValue.Int32)
			issue.ExpectedValue = &v
		}
		if actualValue.Valid {
			v := uint16(actualValue.Int32)
			issue.ActualValue = &v
		}
		if difference.Valid {
			v := int16(difference.Int32)
			issue.Difference = &v
		}
		if resolutionNotes.Valid {
			issue.ResolutionNotes = &resolutionNotes.String
		}
		if resolvedBy.Valid {
			v := uint16(resolvedBy.Int32)
			issue.ResolvedBy = &v
		}
		if resolvedAt.Valid {
			issue.ResolvedAt = &resolvedAt.Time
		}

		issues = append(issues, &issue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	return issues, nil
}
