package admin

import (
	"net/http"
	"strconv"

	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
)

// DashboardHandler handles admin dashboard endpoints.
type DashboardHandler struct {
	dashboard *service.AdminDashboardService
}

// NewDashboardHandler creates a new admin dashboard handler.
func NewDashboardHandler(dashboard *service.AdminDashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

// GetDashboard returns dashboard data.
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	days := 3
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	data, err := h.dashboard.GetDashboard(r.Context(), days)
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to get dashboard data")
		return
	}

	response.RespondJSON(w, http.StatusOK, data)
}
