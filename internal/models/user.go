package models

import "time"

type User struct {
	BaseModel

	FirstName      string    `gorm:"size:100;not null"`
	LastName       string    `gorm:"size:100;not null"`
	DocumentType   string    `gorm:"size:20;not null"`
	DocumentNumber string    `gorm:"size:50;not null"`
	Birthdate      time.Time `gorm:"not null"`
	Phone          string    `gorm:"size:50;not null"`
	Email          string    `gorm:"size:100;unique;not null"`
	PasswordHash   string    `gorm:"size:255;not null"`
	Position       string    `gorm:"size:100;not null"`
	Salary         float64   `gorm:"not null"`
	IsActive       bool      `gorm:"default:true"`

	UserRoles  []UserRole
	StoreUsers []StoreUser
}

type UserRole struct {
	UserID      uint `gorm:"primaryKey"`
	RoleID      uint `gorm:"primaryKey"`
	FranchiseID uint `gorm:"primaryKey"`

	User      User      `gorm:"foreignKey:UserID"`
	Role      Role      `gorm:"foreignKey:RoleID"`
	Franchise Franchise `gorm:"foreignKey:FranchiseID"`
}
