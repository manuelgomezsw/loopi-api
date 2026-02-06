package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// AdminHandler handles admin-related HTTP requests.
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// GetDashboard returns dashboard data.
func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Get days parameter, default to 3
	days := 3
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	data, err := h.adminService.GetDashboard(r.Context(), days)
	if err != nil {
		http.Error(w, `{"error":"failed to get dashboard data"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ListInventories returns a paginated list of inventories with filters.
func (h *AdminHandler) ListInventories(w http.ResponseWriter, r *http.Request) {
	filter := service.InventoryFilter{
		Page:     1,
		PageSize: 20,
	}

	// Parse query parameters
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			filter.PageSize = ps
		}
	}
	if dateFromStr := r.URL.Query().Get("date_from"); dateFromStr != "" {
		if t, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateToStr := r.URL.Query().Get("date_to"); dateToStr != "" {
		if t, err := time.Parse("2006-01-02", dateToStr); err == nil {
			filter.DateTo = &t
		}
	}
	if invType := r.URL.Query().Get("inventory_type"); invType != "" {
		it := entity.InventoryType(invType)
		filter.InventoryType = &it
	}
	if employeeIDStr := r.URL.Query().Get("employee_id"); employeeIDStr != "" {
		if id, err := strconv.ParseUint(employeeIDStr, 10, 16); err == nil {
			eid := uint16(id)
			filter.EmployeeID = &eid
		}
	}
	if hasDiscStr := r.URL.Query().Get("has_discrepancies"); hasDiscStr != "" {
		hasDisc := hasDiscStr == "true"
		filter.HasDiscrepancies = &hasDisc
	}

	result, err := h.adminService.ListInventories(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list inventories"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetInventoryDetail returns detailed information about an inventory.
func (h *AdminHandler) GetInventoryDetail(w http.ResponseWriter, r *http.Request) {
	inventoryIDStr := chi.URLParam(r, "inventoryID")
	inventoryID, err := strconv.ParseUint(inventoryIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"invalid inventory ID"}`, http.StatusBadRequest)
		return
	}

	detail, err := h.adminService.GetInventoryDetail(r.Context(), uint32(inventoryID))
	if err != nil {
		http.Error(w, `{"error":"failed to get inventory detail"}`, http.StatusInternalServerError)
		return
	}
	if detail == nil {
		http.Error(w, `{"error":"inventory not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// UpdateInventoryDetailRequest is the request body for updating an inventory detail.
type UpdateInventoryDetailRequest struct {
	RealValue     *uint16 `json:"real_value"`
	StockReceived *uint16 `json:"stock_received"`
	UnitsSold     *uint16 `json:"units_sold"`
}

// UpdateInventoryDetail updates a specific inventory detail.
func (h *AdminHandler) UpdateInventoryDetail(w http.ResponseWriter, r *http.Request) {
	inventoryIDStr := chi.URLParam(r, "inventoryID")
	inventoryID, err := strconv.ParseUint(inventoryIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"invalid inventory ID"}`, http.StatusBadRequest)
		return
	}

	detailIDStr := chi.URLParam(r, "detailID")
	detailID, err := strconv.ParseUint(detailIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"invalid detail ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateInventoryDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.adminService.UpdateInventoryDetail(r.Context(), uint32(inventoryID), uint32(detailID), req.RealValue, req.StockReceived, req.UnitsSold); err != nil {
		http.Error(w, `{"error":"failed to update inventory detail"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
