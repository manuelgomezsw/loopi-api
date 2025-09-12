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

type NoveltyHandler struct {
	uc usecase.NoveltyUseCase
}

func NewNoveltyHandler(uc usecase.NoveltyUseCase) *NoveltyHandler {
	return &NoveltyHandler{uc}
}

func (h *NoveltyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.Novelty
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.BadRequest(w, "Invalid request body")
		return
	}
	if err := h.uc.Create(req); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.Created(w, map[string]string{"message": "Novelty registered"})
}

func (h *NoveltyHandler) GetByEmployeeAndMonth(w http.ResponseWriter, r *http.Request) {
	employeeID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	if employeeID == 0 || year == 0 || month == 0 {
		rest.BadRequest(w, "Missing parameters")
		return
	}

	novelties, err := h.uc.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, novelties)
}

// GetByEmployeeAndDateRange retrieves novelties by employee within a custom date range
func (h *NoveltyHandler) GetByEmployeeAndDateRange(w http.ResponseWriter, r *http.Request) {
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

	novelties, err := h.uc.GetByEmployeeAndDateRange(employeeID, from, to)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, novelties)
}

// GetTotalHoursByType retrieves total novelty hours by type for an employee in a month
func (h *NoveltyHandler) GetTotalHoursByType(w http.ResponseWriter, r *http.Request) {
	employeeID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	noveltyType := r.URL.Query().Get("type")

	if employeeID == 0 || year == 0 || month == 0 || noveltyType == "" {
		rest.BadRequest(w, "Missing parameters: employee, year, month, type")
		return
	}

	totalHours, err := h.uc.GetTotalHoursByEmployeeAndType(employeeID, year, month, noveltyType)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, map[string]interface{}{
		"employee_id":  employeeID,
		"year":         year,
		"month":        month,
		"novelty_type": noveltyType,
		"total_hours":  totalHours,
	})
}

// GetTypesSummary retrieves a summary of novelty types for an employee in a month
func (h *NoveltyHandler) GetTypesSummary(w http.ResponseWriter, r *http.Request) {
	employeeID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	if employeeID == 0 || year == 0 || month == 0 {
		rest.BadRequest(w, "Missing parameters: employee, year, month")
		return
	}

	summary, err := h.uc.GetNoveltyTypesSummary(employeeID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
		"summary":     summary,
	})
}
