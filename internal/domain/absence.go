package domain

import "time"

type Absence struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	EmployeeID int       `json:"employee_id"`
	Date       time.Time `json:"date"`
	Hours      float64   `json:"hours"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
