package dto

import "time"

type DayType string

const (
	Ordinary DayType = "ordinary"
	Sunday   DayType = "sunday"
	Holiday  DayType = "holiday"
)

type CalendarDay struct {
	Date    time.Time `json:"date"`
	DayType DayType   `json:"type"`
	Week    int       `json:"week"` // útil para agrupación semanal
}
