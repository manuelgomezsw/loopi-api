package usecase

import (
	"loopi-api/internal/calendar"
	"time"
)

type CalendarUseCase interface {
	GetHolidays(year int) []time.Time
	GetHolidaysByMonth(year int, month int) []time.Time
	CountOrdinaryDays(year int, month int) int
	CountSundays(year int, month int) int
}

type calendarUseCase struct{}

func NewCalendarUseCase() CalendarUseCase {
	return &calendarUseCase{}
}

func (c *calendarUseCase) GetHolidays(year int) []time.Time {
	return calendar.GetColombianHolidaysCached(year)
}

func (c *calendarUseCase) GetHolidaysByMonth(year int, month int) []time.Time {
	return calendar.GetColombianHolidaysByMonthCached(year, month)
}

func (c *calendarUseCase) CountOrdinaryDays(year int, month int) int {
	ord, _ := calendar.GetMonthSummaryCached(year, month)
	return ord
}

func (c *calendarUseCase) CountSundays(year int, month int) int {
	_, sun := calendar.GetMonthSummaryCached(year, month)
	return sun
}
