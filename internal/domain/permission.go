package domain

type Permission struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"` // ✅ Campo ID directo con auto-increment

	Name        string `gorm:"size:100;unique;not null" json:"name"`
	Description string `json:"description"`

	RolePermissions []RolePermission `json:"-"`
}
