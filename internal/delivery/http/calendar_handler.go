package http

import (
	"loopi-api/internal/calendar"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
	"time"
)

type CalendarHandler struct {
	calendarUseCase usecase.CalendarUseCase
}

func NewCalendarHandler(calendarUseCase usecase.CalendarUseCase) *CalendarHandler {
	return &CalendarHandler{calendarUseCase: calendarUseCase}
}

func (h *CalendarHandler) GetHolidays(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	month := 0 // Si no se especifica, trae todo el año

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			month = parsed
		}
	}

	var holidays []time.Time
	if month > 0 {
		holidays = h.calendarUseCase.GetHolidaysByMonth(year, month)
	} else {
		holidays = h.calendarUseCase.GetHolidays(year)
	}

	result := struct {
		Count int      `json:"count"`
		Dates []string `json:"dates"`
	}{
		Count: len(holidays),
		Dates: make([]string, len(holidays)),
	}

	for i, d := range holidays {
		result.Dates[i] = d.Format("2006-01-02")
	}

	rest.OK(w, result)
}

func (h *CalendarHandler) GetMonthSummary(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	month := int(time.Now().Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			month = parsed
		}
	}

	holidayDates := calendar.GetColombianHolidaysByMonthCached(year, month)
	holidayMap := make(map[string]bool)
	for _, d := range holidayDates {
		holidayMap[d.Format("2006-01-02")] = true
	}

	ordinary := h.calendarUseCase.CountOrdinaryDays(year, month)
	sundays := h.calendarUseCase.CountSundays(year, month)

	// Restar solo los festivos que NO caen domingo
	excludeCount := 0
	for _, hDate := range holidayDates {
		if hDate.Weekday() != time.Sunday {
			excludeCount++
		}
	}
	adjustedOrdinary := ordinary - excludeCount
	if adjustedOrdinary < 0 {
		adjustedOrdinary = 0
	}

	result := struct {
		Holidays struct {
			Count int      `json:"count"`
			Dates []string `json:"dates"`
		} `json:"holidays"`
		OrdinaryDays int `json:"ordinary_days"`
		Sundays      int `json:"sundays"`
	}{}

	result.Holidays.Count = len(holidayDates)
	result.Holidays.Dates = make([]string, len(holidayDates))
	for i, d := range holidayDates {
		result.Holidays.Dates[i] = d.Format("2006-01-02")
	}
	result.OrdinaryDays = adjustedOrdinary
	result.Sundays = sundays

	rest.OK(w, result)
}

func (h *CalendarHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	calendar.ClearCalendarCache()
	rest.OK(w, map[string]string{"message": "Calendar cache cleared"})
}
