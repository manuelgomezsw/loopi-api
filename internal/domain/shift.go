package domain

type Shift struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	StoreID      int    `gorm:"column:store_id;not null" json:"store_id"`
	Name         string `gorm:"column:name;not null" json:"name"`
	StartTime    string `gorm:"column:start_time;type:TIME;not null" json:"start_time"`
	EndTime      string `gorm:"column:end_time;type:TIME;not null" json:"end_time"`
	LunchMinutes int    `gorm:"column:lunch_minutes;default:0" json:"lunch_minutes"`
	IsActive     bool   `gorm:"column:is_active;default:true" json:"is_active"`
}
