package domain

type Store struct {
	BaseEntity

	FranchiseID uint   `json:"franchise_id" gorm:"not null"`
	Code        string `json:"code" gorm:"size:3;unique"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Location    string `json:"location" gorm:"size:255"`
	Address     string `json:"address" gorm:"size:255"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`

	Franchise  Franchise   `json:"-"` // omitido en respuesta JSON para evitar ciclos
	StoreUsers []StoreUser `json:"-"` // omitido también a menos que quieras incluir relaciones
}

type StoreUser struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"` // ✅ Campo ID directo con auto-increment

	StoreID uint `json:"store_id" gorm:"index;not null"`
	UserID  uint `json:"user_id" gorm:"index;not null"`

	Store Store `json:"-" gorm:"foreignKey:StoreID"` // evitar ciclos
	User  User  `json:"-" gorm:"foreignKey:UserID"`
}
