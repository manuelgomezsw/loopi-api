package domain

type Permission struct {
	BaseModel

	Name        string `gorm:"size:100;unique;not null"`
	Description string

	RolePermissions []RolePermission
}
