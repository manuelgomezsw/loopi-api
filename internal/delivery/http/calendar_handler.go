package http

import (
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
	var err error
	if month > 0 {
		holidays, err = h.calendarUseCase.GetHolidaysByMonth(year, month)
	} else {
		holidays, err = h.calendarUseCase.GetHolidays(year)
	}

	if err != nil {
		rest.HandleError(w, err)
		return
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

	// Use the new enhanced usecase method
	summary, err := h.calendarUseCase.GetMonthSummary(year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	// Format response
	result := struct {
		Holidays struct {
			Count int      `json:"count"`
			Dates []string `json:"dates"`
		} `json:"holidays"`
		OrdinaryDays int `json:"ordinary_days"`
		Sundays      int `json:"sundays"`
		WorkingDays  int `json:"working_days"`
	}{}

	result.Holidays.Count = len(summary.Holidays)
	result.Holidays.Dates = make([]string, len(summary.Holidays))
	for i, d := range summary.Holidays {
		result.Holidays.Dates[i] = d.Format("2006-01-02")
	}
	result.OrdinaryDays = summary.OrdinaryDays
	result.Sundays = summary.Sundays
	result.WorkingDays = summary.WorkingDays

	rest.OK(w, result)
}

func (h *CalendarHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	if err := h.calendarUseCase.ClearCache(); err != nil {
		rest.HandleError(w, err)
		return
	}
	rest.OK(w, map[string]string{"message": "Calendar cache cleared"})
}

// GetWorkingDays retrieves the number of working days in a month
func (h *CalendarHandler) GetWorkingDays(w http.ResponseWriter, r *http.Request) {
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

	workingDays, err := h.calendarUseCase.GetWorkingDays(year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, map[string]interface{}{
		"year":         year,
		"month":        month,
		"working_days": workingDays,
	})
}

// GetEnhancedSummary retrieves comprehensive calendar summary using the new usecase method
func (h *CalendarHandler) GetEnhancedSummary(w http.ResponseWriter, r *http.Request) {
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

	summary, err := h.calendarUseCase.GetMonthSummary(year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	// Format holidays as strings for JSON response
	formattedHolidays := make([]string, len(summary.Holidays))
	for i, holiday := range summary.Holidays {
		formattedHolidays[i] = holiday.Format("2006-01-02")
	}

	result := struct {
		Year         int      `json:"year"`
		Month        int      `json:"month"`
		Holidays     []string `json:"holidays"`
		OrdinaryDays int      `json:"ordinary_days"`
		Sundays      int      `json:"sundays"`
		WorkingDays  int      `json:"working_days"`
	}{
		Year:         summary.Year,
		Month:        summary.Month,
		Holidays:     formattedHolidays,
		OrdinaryDays: summary.OrdinaryDays,
		Sundays:      summary.Sundays,
		WorkingDays:  summary.WorkingDays,
	}

	rest.OK(w, result)
}
