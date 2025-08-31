package utils

import (
	"time"
)

type DayType string

const (
	Ordinary DayType = "ordinary"
	Sunday   DayType = "sunday"
	Holiday  DayType = "holiday"
)

type CalendarDay struct {
	Date    time.Time `json:"date"`
	DayType DayType   `json:"type"`
	Week    int       `json:"week"`
}

// HolidaysToMap Mapa de festivos por fecha "YYYY-MM-DD" → true
func HolidaysToMap(holidays []time.Time) map[string]bool {
	m := make(map[string]bool)
	for _, h := range holidays {
		m[h.Format("2006-01-02")] = true
	}
	return m
}

// BuildCalendarDays Construye los días del mes clasificados por tipo
func BuildCalendarDays(year int, month int, holidays map[string]bool) []CalendarDay {
	var days []CalendarDay
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)

	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		dayKey := d.Format("2006-01-02")
		var dtype DayType

		switch {
		case d.Weekday() == time.Sunday:
			dtype = Sunday
		case holidays[dayKey]:
			dtype = Holiday
		default:
			dtype = Ordinary
		}

		week := ((d.Day() - 1) / 7) + 1

		days = append(days, CalendarDay{
			Date:    d,
			DayType: dtype,
			Week:    week,
		})
	}
	return days
}
