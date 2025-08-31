package domain

type WorkConfig struct {
	ID           uint   `gorm:"primaryKey"`
	DiurnalStart string `gorm:"column:diurnal_start"` // "06:00"
	DiurnalEnd   string `gorm:"column:diurnal_end"`   // "21:00"
	IsActive     bool   `gorm:"column:is_active"`
}
