package domain

import "time"

type Franchise struct {
	BaseModel

	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	IsActive  bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Stores    []Store
	UserRoles []UserRole
}
