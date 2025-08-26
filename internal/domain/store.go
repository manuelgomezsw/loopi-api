package domain

import "time"

type Store struct {
	BaseModel

	ID          uint   `gorm:"primaryKey"`
	FranchiseID uint   `gorm:"not null"`
	Code        string `gorm:"size:3;unique"`
	Name        string `gorm:"size:100;not null"`
	Location    string `gorm:"size:255"`
	Address     string `gorm:"size:255"`
	IsActive    bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Franchise  Franchise
	StoreUsers []StoreUser
}

type StoreUser struct {
	ID        uint `gorm:"primaryKey"`
	StoreID   uint `gorm:"index;not null"`
	UserID    uint `gorm:"index;not null"`
	StartDate *time.Time
	EndDate   *time.Time

	Store Store `gorm:"foreignKey:StoreID"`
	User  User  `gorm:"foreignKey:UserID"`
}
