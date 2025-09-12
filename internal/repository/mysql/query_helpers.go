package mysql

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	db *gorm.DB
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{db: db}
}

// WhereEquals adds an equals condition
func (qb *QueryBuilder) WhereEquals(field string, value interface{}) *QueryBuilder {
	qb.db = qb.db.Where(fmt.Sprintf("%s = ?", field), value)
	return qb
}

// WhereIn adds an IN condition
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	qb.db = qb.db.Where(fmt.Sprintf("%s IN ?", field), values)
	return qb
}

// WhereLike adds a LIKE condition
func (qb *QueryBuilder) WhereLike(field string, pattern string) *QueryBuilder {
	qb.db = qb.db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+pattern+"%")
	return qb
}

// WhereActive filters for active records
func (qb *QueryBuilder) WhereActive() *QueryBuilder {
	qb.db = qb.db.Where("is_active = ?", true)
	return qb
}

// WhereNotDeleted filters out soft-deleted records
func (qb *QueryBuilder) WhereNotDeleted() *QueryBuilder {
	qb.db = qb.db.Where("deleted_at IS NULL")
	return qb
}

// WhereDateRange adds a date range condition
func (qb *QueryBuilder) WhereDateRange(field string, from, to time.Time) *QueryBuilder {
	qb.db = qb.db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), from, to)
	return qb
}

// OrderBy adds ordering
func (qb *QueryBuilder) OrderBy(field string, direction ...string) *QueryBuilder {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	qb.db = qb.db.Order(fmt.Sprintf("%s %s", field, dir))
	return qb
}

// Limit adds a limit clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.db = qb.db.Limit(limit)
	return qb
}

// Offset adds an offset clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.db = qb.db.Offset(offset)
	return qb
}

// Preload adds preloading for associations
func (qb *QueryBuilder) Preload(associations ...string) *QueryBuilder {
	for _, assoc := range associations {
		qb.db = qb.db.Preload(assoc)
	}
	return qb
}

// GetDB returns the built query
func (qb *QueryBuilder) GetDB() *gorm.DB {
	return qb.db
}

// Common query patterns

// FindActiveByFranchise finds active records by franchise ID
func FindActiveByFranchise[T any](db *gorm.DB, franchiseID int) ([]T, error) {
	var entities []T
	err := NewQueryBuilder(db).
		WhereEquals("franchise_id", franchiseID).
		WhereActive().
		OrderBy("name").
		GetDB().
		Find(&entities).Error

	return entities, err
}

// FindByStoreAndDateRange finds records by store and date range
func FindByStoreAndDateRange[T any](db *gorm.DB, storeID int, from, to time.Time) ([]T, error) {
	var entities []T
	err := NewQueryBuilder(db).
		WhereEquals("store_id", storeID).
		WhereDateRange("created_at", from, to).
		OrderBy("created_at", "DESC").
		GetDB().
		Find(&entities).Error

	return entities, err
}

// FindByEmployeeAndMonth finds records by employee and month
func FindByEmployeeAndMonth[T any](db *gorm.DB, employeeID, year, month int) ([]T, error) {
	var entities []T
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := NewQueryBuilder(db).
		WhereEquals("employee_id", employeeID).
		WhereDateRange("date", startDate, endDate).
		OrderBy("date").
		GetDB().
		Find(&entities).Error

	return entities, err
}

// FindWithJoin performs a join operation
func FindWithJoin[T any](db *gorm.DB, joinTable, joinCondition string, conditions map[string]interface{}) ([]T, error) {
	var entities []T
	query := db.Joins(fmt.Sprintf("JOIN %s ON %s", joinTable, joinCondition))

	for field, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	err := query.Find(&entities).Error
	return entities, err
}

// PaginationHelper provides pagination utilities
type PaginationHelper struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	Pages    int   `json:"pages"`
}

// NewPaginationHelper creates a new pagination helper
func NewPaginationHelper(page, pageSize int) *PaginationHelper {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return &PaginationHelper{
		Page:     page,
		PageSize: pageSize,
	}
}

// GetOffset calculates the offset for the current page
func (p *PaginationHelper) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// SetTotal sets the total count and calculates pages
func (p *PaginationHelper) SetTotal(total int64) {
	p.Total = total
	p.Pages = int((total + int64(p.PageSize) - 1) / int64(p.PageSize))
}

// FindWithPagination finds records with pagination
func FindWithPagination[T any](db *gorm.DB, pagination *PaginationHelper, conditions map[string]interface{}) ([]T, error) {
	var entities []T
	var total int64

	// Count total records
	countQuery := db.Model(new(T))
	for field, value := range conditions {
		countQuery = countQuery.Where(fmt.Sprintf("%s = ?", field), value)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	pagination.SetTotal(total)

	// Find records with pagination
	query := db.Offset(pagination.GetOffset()).Limit(pagination.PageSize)
	for field, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	err := query.Find(&entities).Error
	return entities, err
}
