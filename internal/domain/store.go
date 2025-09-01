package domain

import "time"

type Store struct {
	BaseModel

	FranchiseID uint   `json:"franchise_id" gorm:"not null"`
	Code        string `json:"code" gorm:"size:3;unique"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Location    string `json:"location" gorm:"size:255"`
	Address     string `json:"address" gorm:"size:255"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`

	Franchise  Franchise   `json:"-"` // omitido en respuesta JSON para evitar ciclos
	StoreUsers []StoreUser `json:"-"` // omitido también a menos que quieras incluir relaciones
}

type StoreUser struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	StoreID   uint       `json:"store_id" gorm:"index;not null"`
	UserID    uint       `json:"user_id" gorm:"index;not null"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`

	Store Store `json:"-" gorm:"foreignKey:StoreID"` // evitar ciclos
	User  User  `json:"-" gorm:"foreignKey:UserID"`
}
