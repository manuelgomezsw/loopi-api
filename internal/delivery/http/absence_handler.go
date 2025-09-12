package http

import (
	"encoding/json"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
	"time"
)

type AbsenceHandler struct {
	uc usecase.AbsenceUseCase
}

func NewAbsenceHandler(uc usecase.AbsenceUseCase) *AbsenceHandler {
	return &AbsenceHandler{uc}
}

func (h *AbsenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.Absence
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.BadRequest(w, "Invalid request body")
		return
	}
	if err := h.uc.Create(req); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.Created(w, map[string]string{"message": "Absence registered"})
}

func (h *AbsenceHandler) GetByEmployeeAndMonth(w http.ResponseWriter, r *http.Request) {
	employeeID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	if employeeID == 0 || year == 0 || month == 0 {
		rest.BadRequest(w, "Missing parameters")
		return
	}

	absences, err := h.uc.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, absences)
}

// GetByEmployeeAndDateRange retrieves absences by employee within a custom date range
func (h *AbsenceHandler) GetByEmployeeAndDateRange(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := r.URL.Query().Get("employee")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if employeeIDStr == "" || fromStr == "" || toStr == "" {
		rest.BadRequest(w, "Missing required parameters: employee, from, to")
		return
	}

	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil {
		rest.BadRequest(w, "Invalid employee ID")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		rest.BadRequest(w, "Invalid from date format. Use YYYY-MM-DD")
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		rest.BadRequest(w, "Invalid to date format. Use YYYY-MM-DD")
		return
	}

	absences, err := h.uc.GetByEmployeeAndDateRange(employeeID, from, to)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, absences)
}

// GetTotalHours retrieves total absence hours for an employee in a month
func (h *AbsenceHandler) GetTotalHours(w http.ResponseWriter, r *http.Request) {
	employeeID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	if employeeID == 0 || year == 0 || month == 0 {
		rest.BadRequest(w, "Missing parameters: employee, year, month")
		return
	}

	totalHours, err := h.uc.GetTotalHoursByEmployee(employeeID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
		"total_hours": totalHours,
	})
}
