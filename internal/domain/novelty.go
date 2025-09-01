package domain

import "time"

type Novelty struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	EmployeeID int       `json:"employee_id"`
	Date       time.Time `json:"date"`
	Hours      float64   `json:"hours"`
	Type       string    `json:"type"` // "positive" or "negative"
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
