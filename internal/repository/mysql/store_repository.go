package mysql

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"

	"gorm.io/gorm"
)

// storeRepository implements repository.StoreRepository with improved maintainability
type storeRepository struct {
	*BaseRepository[domain.Store]
	errorHandler *ErrorHandler
}

// NewStoreRepository creates a new store repository with enhanced features
func NewStoreRepository(db *gorm.DB) repository.StoreRepository {
	return &storeRepository{
		BaseRepository: NewBaseRepository[domain.Store](db, "stores"),
		errorHandler:   NewErrorHandler("stores"),
	}
}

// GetAll retrieves all stores (uses base repository)
func (r *storeRepository) GetAll() ([]domain.Store, error) {
	stores, err := r.BaseRepository.GetAll()
	if err != nil {
		return nil, r.errorHandler.HandleError("GetAll", err)
	}
	return stores, nil
}

// GetByID retrieves a store by ID (uses base repository)
func (r *storeRepository) GetByID(id int) (domain.Store, error) {
	store, err := r.BaseRepository.GetByID(id)
	if err != nil {
		if err == ErrNotFound {
			return domain.Store{}, r.errorHandler.HandleNotFound("GetByID", id)
		}
		return domain.Store{}, r.errorHandler.HandleError("GetByID", err, id)
	}
	return *store, nil
}

// GetByFranchiseID retrieves stores by franchise ID using query helper
func (r *storeRepository) GetByFranchiseID(franchiseID int) ([]domain.Store, error) {
	stores, err := FindActiveByFranchise[domain.Store](r.GetDB(), franchiseID)
	if err != nil {
		return nil, r.errorHandler.HandleError("GetByFranchiseID", err, franchiseID)
	}
	return stores, nil
}

// Create creates a new store with validation and error handling
func (r *storeRepository) Create(store *domain.Store) error {
	// Business validation before creation
	if err := r.validateStore(store); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	if err := r.BaseRepository.Create(store); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}
	return nil
}

// Update updates an existing store
func (r *storeRepository) Update(store *domain.Store) error {
	// Validate before update
	if err := r.validateStore(store); err != nil {
		return r.errorHandler.HandleError("Update", err, store.ID)
	}

	if err := r.BaseRepository.Update(store); err != nil {
		return r.errorHandler.HandleError("Update", err, store.ID)
	}
	return nil
}

// Delete removes a store by ID
func (r *storeRepository) Delete(id int) error {
	// Check if store exists before deletion
	exists, err := r.BaseRepository.Exists(id)
	if err != nil {
		return r.errorHandler.HandleError("Delete", err, id)
	}
	if !exists {
		return r.errorHandler.HandleNotFound("Delete", id)
	}

	if err := r.BaseRepository.Delete(id); err != nil {
		return r.errorHandler.HandleError("Delete", err, id)
	}
	return nil
}

// GetActiveStoresByFranchise retrieves only active stores for a franchise
func (r *storeRepository) GetActiveStoresByFranchise(franchiseID int) ([]domain.Store, error) {
	var stores []domain.Store
	err := NewQueryBuilder(r.GetDB()).
		WhereEquals("franchise_id", franchiseID).
		WhereActive().
		WhereNotDeleted().
		OrderBy("name").
		GetDB().
		Find(&stores).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetActiveStoresByFranchise", err, franchiseID)
	}
	return stores, nil
}

// GetStoresWithEmployeeCount retrieves stores with employee count
func (r *storeRepository) GetStoresWithEmployeeCount(franchiseID int) ([]repository.StoreWithEmployeeCount, error) {
	var results []repository.StoreWithEmployeeCount

	err := r.GetDB().
		Table("stores").
		Select("stores.*, COUNT(store_users.user_id) as employee_count").
		Joins("LEFT JOIN store_users ON stores.id = store_users.store_id").
		Where("stores.franchise_id = ? AND stores.is_active = ?", franchiseID, true).
		Group("stores.id").
		Order("stores.name").
		Scan(&results).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetStoresWithEmployeeCount", err, franchiseID)
	}
	return results, nil
}

// validateStore performs business validation
func (r *storeRepository) validateStore(store *domain.Store) error {
	if store.Name == "" {
		return ErrInvalidInput
	}
	if store.FranchiseID <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// GetStoreStatistics retrieves comprehensive store statistics
func (r *storeRepository) GetStoreStatistics(storeID int) (*repository.StoreStatistics, error) {
	var stats repository.StoreStatistics

	// Get basic store info
	store, err := r.GetByID(storeID)
	if err != nil {
		return nil, err
	}
	stats.Store = store

	// Use transaction for consistency
	err = r.BaseRepository.Transaction(func(tx *gorm.DB) error {
		// Get employee count
		if err := tx.Table("store_users").
			Where("store_id = ?", storeID).
			Count(&stats.EmployeeCount).Error; err != nil {
			return err
		}

		// Get shift count
		if err := tx.Table("shifts").
			Where("store_id = ?", storeID).
			Count(&stats.ShiftCount).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, r.errorHandler.HandleError("GetStoreStatistics", err, storeID)
	}

	return &stats, nil
}
