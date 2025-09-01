package domain

type Franchise struct {
	BaseModel

	Name     string `json:"name" gorm:"size:100;not null"`
	IsActive bool   `json:"is_active" gorm:"default:true"`

	Stores    []Store    `json:"-"` // omitido en JSON por defecto
	UserRoles []UserRole `json:"-"`
}
