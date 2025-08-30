package domain

type Franchise struct {
	BaseModel

	Name     string `gorm:"size:100;not null"`
	IsActive bool   `gorm:"default:true"`

	Stores    []Store
	UserRoles []UserRole
}
