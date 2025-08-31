package calendar

import (
	"fmt"
	"sync"
	"time"
)

var (
	holidayCache sync.Map // key: "2025" o "2025-04"
	summaryCache sync.Map // key: "2025-04"
)

// GetColombianHolidaysCached Cache de festivos por año
func GetColombianHolidaysCached(year int) []time.Time {
	/*
	    key := fmt.Sprintf("%d", year)
	  	if val, ok := holidayCache.Load(key); ok {
	  		return val.([]time.Time)
	  	}*/
	holidays := GetColombianHolidays(year)
	//holidayCache.Store(key, holidays)
	return holidays
}

// GetColombianHolidaysByMonthCached Cache de festivos por mes
func GetColombianHolidaysByMonthCached(year int, month int) []time.Time {
	/*
	    key := fmt.Sprintf("%d-%02d", year, month)
	  	if val, ok := holidayCache.Load(key); ok {
	  		return val.([]time.Time)
	  	}
	*/
	all := GetColombianHolidaysCached(year)

	var filtered []time.Time
	for _, d := range all {
		if d.Month() == time.Month(month) {
			filtered = append(filtered, d)
		}
	}

	//holidayCache.Store(key, filtered)
	return filtered
}

// GetMonthSummaryCached Cache de resumen mensual: ordinary_days, sundays
func GetMonthSummaryCached(year int, month int) (ordinaryDays int, sundays int) {
	key := fmt.Sprintf("%d-%02d", year, month)
	if val, ok := summaryCache.Load(key); ok {
		s := val.(map[string]int)
		return s["ordinary_days"], s["sundays"]
	}

	daysInMonth := daysIn(year, time.Month(month))
	ord := 0
	sun := 0

	for day := 1; day <= daysInMonth; day++ {
		d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		switch d.Weekday() {
		case time.Sunday:
			sun++
		default:
			ord++
		}
	}

	summaryCache.Store(key, map[string]int{
		"ordinary_days": ord,
		"sundays":       sun,
	})
	return ord, sun
}

// ClearCalendarCache borra todos los datos cacheados de días festivos y resúmenes mensuales.
func ClearCalendarCache() {
	holidayCache = sync.Map{}
	summaryCache = sync.Map{}
}

// daysIn Obtiene la cantidad de días exactos del mes
func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
