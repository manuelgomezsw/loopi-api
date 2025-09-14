package domain

type Role struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"` // ✅ Campo ID directo con auto-increment

	Name        string `gorm:"size:50;unique;not null" json:"name"`
	Description string `json:"description"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	RolePermissions []RolePermission `gorm:"many2many:role_permissions" json:"-"`
	UserRoles       []UserRole       `json:"-"`
}

type RolePermission struct {
	RoleID       int `gorm:"primaryKey"`
	PermissionID int `gorm:"primaryKey"`

	Role       Role       `gorm:"foreignKey:RoleID"`
	Permission Permission `gorm:"foreignKey:PermissionID"`
}

// ✅ Sin timestamps - coincide con la tabla role_permissions que solo tiene role_id y permission_id
