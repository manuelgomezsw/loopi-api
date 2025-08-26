package domain

type Permission struct {
	BaseModel

	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:100;unique;not null"`
	Description string

	RolePermissions []RolePermission
}
