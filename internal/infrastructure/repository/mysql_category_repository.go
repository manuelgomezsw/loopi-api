package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// mysqlCategoryRepository implements repository.CategoryRepository.
type mysqlCategoryRepository struct {
	db *sql.DB
}

// NewMySQLCategoryRepository creates a new MySQL category repository.
func NewMySQLCategoryRepository(db *sql.DB) repository.CategoryRepository {
	return &mysqlCategoryRepository{db: db}
}

// FindAll retrieves all categories ordered by display_order.
func (r *mysqlCategoryRepository) FindAll(ctx context.Context) ([]*entity.Category, error) {
	query := `
		SELECT c.id, c.name, c.display_order, c.active, c.created_at, c.updated_at,
		       COUNT(i.id) as item_count
		FROM categories c
		LEFT JOIN items i ON c.id = i.category_id AND i.active = 1
		GROUP BY c.id
		ORDER BY c.display_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// FindAllActive retrieves all active categories ordered by display_order.
func (r *mysqlCategoryRepository) FindAllActive(ctx context.Context) ([]*entity.Category, error) {
	query := `
		SELECT c.id, c.name, c.display_order, c.active, c.created_at, c.updated_at,
		       COUNT(i.id) as item_count
		FROM categories c
		LEFT JOIN items i ON c.id = i.category_id AND i.active = 1
		WHERE c.active = 1
		GROUP BY c.id
		ORDER BY c.display_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active categories: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// FindByID retrieves a category by its ID.
func (r *mysqlCategoryRepository) FindByID(ctx context.Context, id uint16) (*entity.Category, error) {
	query := `
		SELECT c.id, c.name, c.display_order, c.active, c.created_at, c.updated_at,
		       COUNT(i.id) as item_count
		FROM categories c
		LEFT JOIN items i ON c.id = i.category_id AND i.active = 1
		WHERE c.id = ?
		GROUP BY c.id
	`

	var cat entity.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cat.ID, &cat.Name, &cat.DisplayOrder, &cat.Active,
		&cat.CreatedAt, &cat.UpdatedAt, &cat.ItemCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find category by id: %w", err)
	}

	return &cat, nil
}

// FindByName retrieves a category by its name.
func (r *mysqlCategoryRepository) FindByName(ctx context.Context, name string) (*entity.Category, error) {
	query := `
		SELECT id, name, display_order, active, created_at, updated_at
		FROM categories
		WHERE name = ?
	`

	var cat entity.Category
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&cat.ID, &cat.Name, &cat.DisplayOrder, &cat.Active,
		&cat.CreatedAt, &cat.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find category by name: %w", err)
	}

	return &cat, nil
}

// Create creates a new category.
func (r *mysqlCategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	query := `
		INSERT INTO categories (name, display_order, active)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, category.Name, category.DisplayOrder, category.Active)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	category.ID = uint16(id)
	return nil
}

// Update updates an existing category.
func (r *mysqlCategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	query := `
		UPDATE categories
		SET name = ?, display_order = ?, active = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, category.Name, category.DisplayOrder, category.Active, category.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// UpdateStatus updates the active status of a category.
func (r *mysqlCategoryRepository) UpdateStatus(ctx context.Context, id uint16, active bool) error {
	query := `UPDATE categories SET active = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, active, id)
	if err != nil {
		return fmt.Errorf("failed to update category status: %w", err)
	}

	return nil
}

// UpdateDisplayOrders updates the display order for multiple categories.
func (r *mysqlCategoryRepository) UpdateDisplayOrders(ctx context.Context, orders map[uint16]int) error {
	if len(orders) == 0 {
		return nil
	}

	// Build batch update query
	var cases []string
	var ids []interface{}
	for id, order := range orders {
		cases = append(cases, fmt.Sprintf("WHEN %d THEN %d", id, order))
		ids = append(ids, id)
	}

	query := fmt.Sprintf(`
		UPDATE categories
		SET display_order = CASE id %s END
		WHERE id IN (?%s)
	`, strings.Join(cases, " "), strings.Repeat(", ?", len(ids)-1))

	_, err := r.db.ExecContext(ctx, query, ids...)
	if err != nil {
		return fmt.Errorf("failed to update display orders: %w", err)
	}

	return nil
}

// GetMaxDisplayOrder returns the maximum display order value.
func (r *mysqlCategoryRepository) GetMaxDisplayOrder(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(MAX(display_order), 0) FROM categories`

	var maxOrder int
	err := r.db.QueryRowContext(ctx, query).Scan(&maxOrder)
	if err != nil {
		return 0, fmt.Errorf("failed to get max display order: %w", err)
	}

	return maxOrder, nil
}

// scanCategories scans multiple category rows.
func (r *mysqlCategoryRepository) scanCategories(rows *sql.Rows) ([]*entity.Category, error) {
	var categories []*entity.Category
	for rows.Next() {
		var cat entity.Category
		err := rows.Scan(
			&cat.ID, &cat.Name, &cat.DisplayOrder, &cat.Active,
			&cat.CreatedAt, &cat.UpdatedAt, &cat.ItemCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &cat)
	}
	return categories, nil
}
