package domain

type Shift struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	StoreID      int    `gorm:"column:store_id;not null" json:"store_id"`
	Name         string `json:"name"`
	Period       string `json:"period"` // weekly, biweekly, monthly
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	LunchMinutes int    `json:"lunch_minutes"`
	IsActive     bool   `json:"is_active"`
}
