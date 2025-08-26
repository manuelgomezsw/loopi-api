package domain

type Role struct {
	BaseModel

	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:50;unique;not null"`
	Description string
	IsActive    bool `gorm:"default:true"`

	RolePermissions []RolePermission
	UserRoles       []UserRole
}

type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`

	Role       Role       `gorm:"foreignKey:RoleID"`
	Permission Permission `gorm:"foreignKey:PermissionID"`
}
