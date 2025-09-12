package domain

type Permission struct {
	BaseEntity

	Name        string `gorm:"size:100;unique;not null" json:"name"`
	Description string `json:"description"`

	RolePermissions []RolePermission `json:"-"`
}
