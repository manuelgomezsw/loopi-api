package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/dto/request"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
)

// InventoryHandler handles admin inventory endpoints (list, detail, update detail, initial).
type InventoryHandler struct {
	inventory *service.AdminInventoryService
}

// NewInventoryHandler creates a new admin inventory handler.
func NewInventoryHandler(inventory *service.AdminInventoryService) *InventoryHandler {
	return &InventoryHandler{inventory: inventory}
}

// ListInventories returns a paginated list of inventories with filters.
func (h *InventoryHandler) ListInventories(w http.ResponseWriter, r *http.Request) {
	filter := service.InventoryFilter{Page: 1, PageSize: 20}

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

	result, err := h.inventory.ListInventories(r.Context(), filter)
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list inventories")
		return
	}

	response.RespondJSON(w, http.StatusOK, result)
}

// GetInventoryDetail returns detailed information about an inventory.
func (h *InventoryHandler) GetInventoryDetail(w http.ResponseWriter, r *http.Request) {
	inventoryIDStr := chi.URLParam(r, "inventoryID")
	inventoryID, err := strconv.ParseUint(inventoryIDStr, 10, 32)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory ID")
		return
	}

	detail, err := h.inventory.GetInventoryDetail(r.Context(), uint32(inventoryID))
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to get inventory detail")
		return
	}
	if detail == nil {
		response.RespondError(w, http.StatusNotFound, "inventory not found")
		return
	}

	response.RespondJSON(w, http.StatusOK, detail)
}

// UpdateInventoryDetailRequest is the request body for updating an inventory detail.
type UpdateInventoryDetailRequest struct {
	SuggestedValue *uint16 `json:"suggested_value"`
	RealValue      *uint16 `json:"real_value"`
	StockReceived  *uint16 `json:"stock_received"`
	UnitsSold      *uint16 `json:"units_sold"`
	Shrinkage      *uint16 `json:"shrinkage"`
}

// UpdateInventoryDetail updates a specific inventory detail.
func (h *InventoryHandler) UpdateInventoryDetail(w http.ResponseWriter, r *http.Request) {
	inventoryIDStr := chi.URLParam(r, "inventoryID")
	inventoryID, err := strconv.ParseUint(inventoryIDStr, 10, 32)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory ID")
		return
	}

	detailIDStr := chi.URLParam(r, "detailID")
	detailID, err := strconv.ParseUint(detailIDStr, 10, 32)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid detail ID")
		return
	}

	var req UpdateInventoryDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.inventory.UpdateInventoryDetail(r.Context(), uint32(inventoryID), uint32(detailID), req.SuggestedValue, req.RealValue, req.StockReceived, req.UnitsSold, req.Shrinkage); err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to update inventory detail")
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetActiveInventoriesCount returns the count of in-progress inventories.
func (h *InventoryHandler) GetActiveInventoriesCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.inventory.GetActiveInventoriesCount(r.Context())
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to get active inventories count")
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]int{"count": count})
}

// CreateInitialInventory creates a new initial (baseline) inventory.
func (h *InventoryHandler) CreateInitialInventory(w http.ResponseWriter, r *http.Request) {
	var req request.CreateInitialInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	inventory, err := h.inventory.CreateInitialInventory(r.Context(), req.ResponsibleID)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusCreated, map[string]interface{}{"inventory": inventory})
}
