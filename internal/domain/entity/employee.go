package entity

import "time"

// Role represents the role of an employee.
type Role string

const (
	RoleEmployee Role = "employee"
	RoleAdmin    Role = "admin"
)

// Employee represents a store employee.
type Employee struct {
	ID           uint16    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose password hash
	Name         string    `json:"name"`
	LastName     string    `json:"last_name"`
	Role         Role      `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FullName returns the full name of the employee.
func (e *Employee) FullName() string {
	return e.Name + " " + e.LastName
}
