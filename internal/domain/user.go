package domain

type User struct {
	BaseEntity

	FirstName      string  `gorm:"size:100;not null" json:"first_name"`
	LastName       string  `gorm:"size:100;not null" json:"last_name"`
	DocumentType   string  `gorm:"size:20;not null" json:"document_type"`
	DocumentNumber string  `gorm:"size:50;not null" json:"document_number"`
	Birthdate      string  `gorm:"not null" json:"birthdate"`
	Phone          string  `gorm:"size:50;not null" json:"phone"`
	Email          string  `gorm:"size:100;unique;not null" json:"email"`
	PasswordHash   string  `gorm:"size:255;not null" json:"-"`
	Position       string  `gorm:"size:100;not null" json:"position"`
	Salary         float64 `gorm:"not null" json:"salary"`
	IsActive       bool    `gorm:"default:true" json:"is_active"`

	UserRoles  []UserRole  `json:"-"`
	StoreUsers []StoreUser `json:"-"`
}

type UserRole struct {
	UserID      int `gorm:"primaryKey"`
	RoleID      int `gorm:"primaryKey"`
	FranchiseID int `gorm:"primaryKey"`

	User      User      `gorm:"foreignKey:UserID"`
	Role      Role      `gorm:"foreignKey:RoleID"`
	Franchise Franchise `gorm:"foreignKey:FranchiseID"`
}

// ✅ Sin timestamps - coincide con la tabla user_roles que solo tiene las 3 claves foráneas
