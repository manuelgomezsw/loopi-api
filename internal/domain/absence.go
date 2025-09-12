package domain

import "time"

type Absence struct {
	BaseEntity

	EmployeeID int       `json:"employee_id"`
	Date       time.Time `json:"date"`
	Hours      float64   `json:"hours"`
	Reason     string    `json:"reason"`
}
