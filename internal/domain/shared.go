package domain

import "time"

type BaseModel struct {
	ID        uint       `gorm:"primaryKey" json:"-"`
	CreatedAt *time.Time `json:"-"`
	UpdatedAt *time.Time `json:"-"`
}
