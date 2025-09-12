package domain

type WorkConfig struct {
	BaseEntity

	DiurnalStart string `gorm:"column:diurnal_start" json:"diurnal_start"` // "06:00"
	DiurnalEnd   string `gorm:"column:diurnal_end" json:"diurnal_end"`     // "21:00"
	IsActive     bool   `gorm:"column:is_active" json:"is_active"`
}
