package domain

import "time"

type Novelty struct {
	BaseEntity

	EmployeeID int       `json:"employee_id"`
	Date       time.Time `json:"date"`
	Hours      float64   `json:"hours"`
	Type       string    `json:"type"` // "positive" or "negative"
	Comment    string    `json:"comment"`
}
