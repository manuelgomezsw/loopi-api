package repository

import (
	"loopi-api/internal/domain"
)

// StoreWithEmployeeCount represents a store with its employee count
type StoreWithEmployeeCount struct {
	domain.Store
	EmployeeCount int `json:"employee_count"`
}

// StoreStatistics represents comprehensive store statistics
type StoreStatistics struct {
	Store         domain.Store `json:"store"`
	EmployeeCount int64        `json:"employee_count"`
	ShiftCount    int64        `json:"shift_count"`
}

type StoreRepository interface {
	// Basic CRUD operations
	GetAll() ([]domain.Store, error)
	GetByID(id int) (domain.Store, error)
	GetByFranchiseID(franchiseID int) ([]domain.Store, error)
	Create(s *domain.Store) error
	Update(s *domain.Store) error
	Delete(id int) error

	// Enhanced business operations
	GetActiveStoresByFranchise(franchiseID int) ([]domain.Store, error)
	GetStoresWithEmployeeCount(franchiseID int) ([]StoreWithEmployeeCount, error)
	GetStoreStatistics(storeID int) (*StoreStatistics, error)
}
