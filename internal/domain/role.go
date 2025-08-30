package domain

type Role struct {
	BaseModel

	Name        string `gorm:"size:50;unique;not null"`
	Description string
	IsActive    bool `gorm:"default:true"`

	RolePermissions []RolePermission `gorm:"many2many:role_permissions"`
	UserRoles       []UserRole
}

type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`

	Role       Role       `gorm:"foreignKey:RoleID"`
	Permission Permission `gorm:"foreignKey:PermissionID"`
}
