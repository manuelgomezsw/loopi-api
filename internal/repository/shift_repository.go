package repository

import (
	"loopi-api/internal/domain"
)

// ShiftStatistics represents comprehensive shift statistics
type ShiftStatistics struct {
	TotalShifts    int64              `json:"total_shifts"`
	ActiveShifts   int64              `json:"active_shifts"`
	ShiftsByPeriod []ShiftPeriodCount `json:"shifts_by_period"`
}

// ShiftPeriodCount represents shift count by period
type ShiftPeriodCount struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

type ShiftRepository interface {
	// Basic CRUD operations
	Create(cfg domain.Shift) error
	ListAll() ([]domain.Shift, error)
	ListByStore(storeID int) ([]domain.Shift, error)
	GetByID(id int) (*domain.Shift, error)

	// Enhanced business operations
	GetActiveShiftsByStore(storeID int) ([]domain.Shift, error)
	GetShiftsByPeriod(period string) ([]domain.Shift, error)
	GetShiftStatistics(storeID int) (*ShiftStatistics, error)
}
