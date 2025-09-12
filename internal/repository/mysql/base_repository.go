package mysql

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// BaseRepository provides common CRUD operations
type BaseRepository[T any] struct {
	db        *gorm.DB
	tableName string
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository[T any](db *gorm.DB, tableName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:        db,
		tableName: tableName,
	}
}

// Create inserts a new record
func (r *BaseRepository[T]) Create(entity *T) error {
	if err := r.db.Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create %s: %w", r.tableName, err)
	}
	return nil
}

// GetByID retrieves a record by ID
func (r *BaseRepository[T]) GetByID(id int) (*T, error) {
	var entity T
	err := r.db.First(&entity, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get %s by ID %d: %w", r.tableName, id, err)
	}

	return &entity, nil
}

// GetAll retrieves all records
func (r *BaseRepository[T]) GetAll() ([]T, error) {
	var entities []T
	err := r.db.Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get all %s: %w", r.tableName, err)
	}

	return entities, nil
}

// Update saves changes to an existing record
func (r *BaseRepository[T]) Update(entity *T) error {
	if err := r.db.Save(entity).Error; err != nil {
		return fmt.Errorf("failed to update %s: %w", r.tableName, err)
	}
	return nil
}

// Delete removes a record by ID
func (r *BaseRepository[T]) Delete(id int) error {
	var entity T
	if err := r.db.Delete(&entity, id).Error; err != nil {
		return fmt.Errorf("failed to delete %s with ID %d: %w", r.tableName, id, err)
	}
	return nil
}

// SoftDelete performs a soft delete (if the model supports it)
func (r *BaseRepository[T]) SoftDelete(id int) error {
	var entity T
	result := r.db.Model(&entity).Where("id = ?", id).Update("deleted_at", "NOW()")

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete %s with ID %d: %w", r.tableName, id, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetWithPreload retrieves a record with preloaded associations
func (r *BaseRepository[T]) GetWithPreload(id int, preloads ...string) (*T, error) {
	var entity T
	query := r.db

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	err := query.First(&entity, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get %s with preloads by ID %d: %w", r.tableName, id, err)
	}

	return &entity, nil
}

// FindBy executes a custom query with conditions
func (r *BaseRepository[T]) FindBy(conditions map[string]interface{}) ([]T, error) {
	var entities []T
	query := r.db

	for field, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	err := query.Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find %s with conditions: %w", r.tableName, err)
	}

	return entities, nil
}

// Count returns the total number of records
func (r *BaseRepository[T]) Count() (int64, error) {
	var count int64
	var entity T

	err := r.db.Model(&entity).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count %s: %w", r.tableName, err)
	}

	return count, nil
}

// Exists checks if a record exists by ID
func (r *BaseRepository[T]) Exists(id int) (bool, error) {
	var count int64
	var entity T

	err := r.db.Model(&entity).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if %s exists: %w", r.tableName, err)
	}

	return count > 0, nil
}

// Paginate retrieves records with pagination
func (r *BaseRepository[T]) Paginate(offset, limit int) ([]T, error) {
	var entities []T
	err := r.db.Offset(offset).Limit(limit).Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to paginate %s: %w", r.tableName, err)
	}

	return entities, nil
}

// Transaction executes a function within a transaction
func (r *BaseRepository[T]) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// GetDB returns the underlying database instance for custom queries
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}
