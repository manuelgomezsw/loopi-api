package http

import (
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type EmployeeHoursHandler struct {
	employeeHoursUseCase usecase.EmployeeHoursUseCase
}

func NewEmployeeHoursHandler(uc usecase.EmployeeHoursUseCase) *EmployeeHoursHandler {
	return &EmployeeHoursHandler{uc}
}

func (h *EmployeeHoursHandler) GetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := chi.URLParam(r, "id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil || employeeID <= 0 {
		rest.BadRequest(w, "Invalid request body")
		return
	}
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

	summary, err := h.employeeHoursUseCase.GetMonthlySummary(employeeID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, summary)
}

// GetDailySummary retrieves daily hours summary for an employee
func (h *EmployeeHoursHandler) GetDailySummary(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := chi.URLParam(r, "id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil || employeeID <= 0 {
		rest.BadRequest(w, "Invalid employee ID")
		return
	}

	// Get year, month, day from query parameters
	year := time.Now().Year()
	month := int(time.Now().Month())
	day := time.Now().Day()

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
	if d := r.URL.Query().Get("day"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			day = parsed
		}
	}

	summary, err := h.employeeHoursUseCase.GetDailySummary(employeeID, year, month, day)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, summary)
}

// GetYearlySummary retrieves yearly hours summary for an employee
func (h *EmployeeHoursHandler) GetYearlySummary(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := chi.URLParam(r, "id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil || employeeID <= 0 {
		rest.BadRequest(w, "Invalid employee ID")
		return
	}

	// Get year from query parameter
	year := time.Now().Year()
	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}

	summary, err := h.employeeHoursUseCase.GetYearlySummary(employeeID, year)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, summary)
}

// GetWorkingDays retrieves the number of working days for an employee in a given month
func (h *EmployeeHoursHandler) GetWorkingDays(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := chi.URLParam(r, "id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil || employeeID <= 0 {
		rest.BadRequest(w, "Invalid employee ID")
		return
	}

	// Get year and month from query parameters
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

	workingDays, err := h.employeeHoursUseCase.CalculateWorkingDays(employeeID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, map[string]interface{}{
		"employee_id":  employeeID,
		"year":         year,
		"month":        month,
		"working_days": workingDays,
	})
}
