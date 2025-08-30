package domain

type User struct {
	BaseModel

	UserID         int     `gorm:"column:id"`
	FirstName      string  `gorm:"size:100;not null"`
	LastName       string  `gorm:"size:100;not null"`
	DocumentType   string  `gorm:"size:20;not null"`
	DocumentNumber string  `gorm:"size:50;not null"`
	Birthdate      string  `gorm:"not null"`
	Phone          string  `gorm:"size:50;not null"`
	Email          string  `gorm:"size:100;unique;not null"`
	PasswordHash   string  `gorm:"size:255;not null"`
	Position       string  `gorm:"size:100;not null"`
	Salary         float64 `gorm:"not null"`
	IsActive       bool    `gorm:"default:true"`

	UserRoles  []UserRole
	StoreUsers []StoreUser
}

type UserRole struct {
	UserID      int `gorm:"primaryKey"`
	RoleID      int `gorm:"primaryKey"`
	FranchiseID int `gorm:"primaryKey"`

	User      User      `gorm:"foreignKey:UserID"`
	Role      Role      `gorm:"foreignKey:RoleID"`
	Franchise Franchise `gorm:"foreignKey:FranchiseID"`
}
