package http

import (
	"encoding/json"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/usecase"
	"loopi-api/internal/usecase/dto"
	"net/http"
	"strconv"
)

type ShiftProjectionHandler struct {
	shiftProjectionUseCase usecase.ShiftProjectionUseCase
}

func NewShiftProjectionHandler(shiftProjectionUseCase usecase.ShiftProjectionUseCase) *ShiftProjectionHandler {
	return &ShiftProjectionHandler{shiftProjectionUseCase: shiftProjectionUseCase}
}

// Preview Proyecta la cantidad de días ordinarios, horas extras, dominicales y festivas
func (h *ShiftProjectionHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var req dto.ShiftProjectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.BadRequest(w, "Invalid input")
		return
	}

	result, err := h.shiftProjectionUseCase.PreviewHours(req)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, result)
}

// GetProjectionSummary retrieves comprehensive shift projection summary
func (h *ShiftProjectionHandler) GetProjectionSummary(w http.ResponseWriter, r *http.Request) {
	// Get parameters from query
	shiftIDStr := r.URL.Query().Get("shift_id")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if shiftIDStr == "" || yearStr == "" || monthStr == "" {
		rest.BadRequest(w, "Missing required parameters: shift_id, year, month")
		return
	}

	shiftID, err := strconv.Atoi(shiftIDStr)
	if err != nil {
		rest.BadRequest(w, "Invalid shift_id")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		rest.BadRequest(w, "Invalid year")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		rest.BadRequest(w, "Invalid month")
		return
	}

	summary, err := h.shiftProjectionUseCase.GetShiftProjectionSummary(shiftID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, summary)
}

// GetProjectedDays calculates projected working days for a shift
func (h *ShiftProjectionHandler) GetProjectedDays(w http.ResponseWriter, r *http.Request) {
	// Get parameters from query
	shiftIDStr := r.URL.Query().Get("shift_id")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if shiftIDStr == "" || yearStr == "" || monthStr == "" {
		rest.BadRequest(w, "Missing required parameters: shift_id, year, month")
		return
	}

	shiftID, err := strconv.Atoi(shiftIDStr)
	if err != nil {
		rest.BadRequest(w, "Invalid shift_id")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		rest.BadRequest(w, "Invalid year")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		rest.BadRequest(w, "Invalid month")
		return
	}

	projectedDays, err := h.shiftProjectionUseCase.CalculateProjectedDays(shiftID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, map[string]interface{}{
		"shift_id":       shiftID,
		"year":           year,
		"month":          month,
		"projected_days": projectedDays,
	})
}
