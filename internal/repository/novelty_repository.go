package repository

import "loopi-api/internal/domain"

type NoveltyRepository interface {
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error)
	Create(novelty *domain.Novelty) error
}
