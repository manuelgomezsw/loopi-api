package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// mysqlSupplierRepository implements repository.SupplierRepository.
type mysqlSupplierRepository struct {
	db *sql.DB
}

// NewMySQLSupplierRepository creates a new MySQL supplier repository.
func NewMySQLSupplierRepository(db *sql.DB) repository.SupplierRepository {
	return &mysqlSupplierRepository{db: db}
}

// FindAll retrieves all suppliers ordered by business_name.
func (r *mysqlSupplierRepository) FindAll(ctx context.Context) ([]*entity.Supplier, error) {
	query := `
		SELECT s.id, s.business_name, s.tax_id, s.contact_name, s.contact_phone, 
		       s.contact_email, s.active, s.created_at, s.updated_at,
		       COUNT(i.id) as item_count
		FROM suppliers s
		LEFT JOIN items i ON s.id = i.supplier_id AND i.active = 1
		GROUP BY s.id
		ORDER BY s.business_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query suppliers: %w", err)
	}
	defer rows.Close()

	return r.scanSuppliers(rows)
}

// FindAllActive retrieves all active suppliers ordered by business_name.
func (r *mysqlSupplierRepository) FindAllActive(ctx context.Context) ([]*entity.Supplier, error) {
	query := `
		SELECT s.id, s.business_name, s.tax_id, s.contact_name, s.contact_phone, 
		       s.contact_email, s.active, s.created_at, s.updated_at,
		       COUNT(i.id) as item_count
		FROM suppliers s
		LEFT JOIN items i ON s.id = i.supplier_id AND i.active = 1
		WHERE s.active = 1
		GROUP BY s.id
		ORDER BY s.business_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active suppliers: %w", err)
	}
	defer rows.Close()

	return r.scanSuppliers(rows)
}

// FindAllWithFilters retrieves suppliers with pagination and filters.
func (r *mysqlSupplierRepository) FindAllWithFilters(ctx context.Context, active *bool, search string, page, pageSize int) ([]*entity.Supplier, int, error) {
	var conditions []string
	var args []interface{}

	if active != nil {
		conditions = append(conditions, "s.active = ?")
		args = append(args, *active)
	}

	if search != "" {
		conditions = append(conditions, "(s.business_name LIKE ? OR s.tax_id LIKE ? OR s.contact_name LIKE ?)")
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(DISTINCT s.id) FROM suppliers s %s`, whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count suppliers: %w", err)
	}

	// Data query
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT s.id, s.business_name, s.tax_id, s.contact_name, s.contact_phone, 
		       s.contact_email, s.active, s.created_at, s.updated_at,
		       COUNT(i.id) as item_count
		FROM suppliers s
		LEFT JOIN items i ON s.id = i.supplier_id AND i.active = 1
		%s
		GROUP BY s.id
		ORDER BY s.business_name ASC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query suppliers: %w", err)
	}
	defer rows.Close()

	suppliers, err := r.scanSuppliers(rows)
	if err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

// FindByID retrieves a supplier by its ID.
func (r *mysqlSupplierRepository) FindByID(ctx context.Context, id uint16) (*entity.Supplier, error) {
	query := `
		SELECT s.id, s.business_name, s.tax_id, s.contact_name, s.contact_phone, 
		       s.contact_email, s.active, s.created_at, s.updated_at,
		       COUNT(i.id) as item_count
		FROM suppliers s
		LEFT JOIN items i ON s.id = i.supplier_id AND i.active = 1
		WHERE s.id = ?
		GROUP BY s.id
	`

	var sup entity.Supplier
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sup.ID, &sup.BusinessName, &sup.TaxID, &sup.ContactName, &sup.ContactPhone,
		&sup.ContactEmail, &sup.Active, &sup.CreatedAt, &sup.UpdatedAt, &sup.ItemCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find supplier by id: %w", err)
	}

	return &sup, nil
}

// FindByTaxID retrieves a supplier by its tax ID (NIT).
func (r *mysqlSupplierRepository) FindByTaxID(ctx context.Context, taxID string) (*entity.Supplier, error) {
	query := `
		SELECT id, business_name, tax_id, contact_name, contact_phone, 
		       contact_email, active, created_at, updated_at
		FROM suppliers
		WHERE tax_id = ?
	`

	var sup entity.Supplier
	err := r.db.QueryRowContext(ctx, query, taxID).Scan(
		&sup.ID, &sup.BusinessName, &sup.TaxID, &sup.ContactName, &sup.ContactPhone,
		&sup.ContactEmail, &sup.Active, &sup.CreatedAt, &sup.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find supplier by tax_id: %w", err)
	}

	return &sup, nil
}

// Create creates a new supplier.
func (r *mysqlSupplierRepository) Create(ctx context.Context, supplier *entity.Supplier) error {
	query := `
		INSERT INTO suppliers (business_name, tax_id, contact_name, contact_phone, contact_email, active)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		supplier.BusinessName, supplier.TaxID, supplier.ContactName,
		supplier.ContactPhone, supplier.ContactEmail, supplier.Active,
	)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	supplier.ID = uint16(id)
	return nil
}

// Update updates an existing supplier.
func (r *mysqlSupplierRepository) Update(ctx context.Context, supplier *entity.Supplier) error {
	query := `
		UPDATE suppliers
		SET business_name = ?, tax_id = ?, contact_name = ?, contact_phone = ?, 
		    contact_email = ?, active = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		supplier.BusinessName, supplier.TaxID, supplier.ContactName,
		supplier.ContactPhone, supplier.ContactEmail, supplier.Active, supplier.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	return nil
}

// UpdateStatus updates the active status of a supplier.
func (r *mysqlSupplierRepository) UpdateStatus(ctx context.Context, id uint16, active bool) error {
	query := `UPDATE suppliers SET active = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, active, id)
	if err != nil {
		return fmt.Errorf("failed to update supplier status: %w", err)
	}

	return nil
}

// scanSuppliers scans multiple supplier rows.
func (r *mysqlSupplierRepository) scanSuppliers(rows *sql.Rows) ([]*entity.Supplier, error) {
	var suppliers []*entity.Supplier
	for rows.Next() {
		var sup entity.Supplier
		err := rows.Scan(
			&sup.ID, &sup.BusinessName, &sup.TaxID, &sup.ContactName, &sup.ContactPhone,
			&sup.ContactEmail, &sup.Active, &sup.CreatedAt, &sup.UpdatedAt, &sup.ItemCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		suppliers = append(suppliers, &sup)
	}
	return suppliers, nil
}
