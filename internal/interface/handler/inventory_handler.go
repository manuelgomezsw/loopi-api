package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/dto/request"
	"github.com/manuelgomezsw/loopi-api/internal/application/dto/response"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/interface/middleware"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
)

// InventoryHandler handles inventory endpoints.
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

// NewInventoryHandler creates a new inventory handler.
func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// GetSuggestedSchedule handles GET /api/inventories/suggested-schedule.
func (h *InventoryHandler) GetSuggestedSchedule(w http.ResponseWriter, r *http.Request) {
	suggested := h.inventoryService.GetSuggestedSchedule()

	var schedule *string
	if suggested.Schedule != nil {
		s := string(*suggested.Schedule)
		schedule = &s
	}

	resp := response.SuggestedScheduleResponse{
		InventoryType: string(suggested.InventoryType),
		Schedule:      schedule,
		Date:          suggested.Date.Format("2006-01-02"),
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetLatest handles GET /api/inventories/latest.
func (h *InventoryHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	inventory, err := h.inventoryService.GetLatestCompleted(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get latest inventory")
		return
	}

	if inventory == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"inventory": nil})
		return
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := response.InventoryResponse{
		ID:            inventory.ID,
		InventoryDate: inventory.InventoryDate.Format("2006-01-02"),
		InventoryType: string(inventory.InventoryType),
		Schedule:      schedule,
		Status:        string(inventory.Status),
		ResponsibleID: inventory.ResponsibleID,
		StartedAt:     inventory.StartedAt,
		CompletedAt:   inventory.CompletedAt,
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"inventory": resp})
}

// GetInProgress handles GET /api/inventories/in-progress.
func (h *InventoryHandler) GetInProgress(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := middleware.GetEmployeeID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventories, err := h.inventoryService.GetInProgress(r.Context(), employeeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get in-progress inventories")
		return
	}

	items := make([]response.InProgressInventoryResponse, 0, len(inventories))
	for _, inv := range inventories {
		var schedule *string
		if inv.Schedule != nil {
			s := string(*inv.Schedule)
			schedule = &s
		}

		items = append(items, response.InProgressInventoryResponse{
			ID:            inv.ID,
			InventoryDate: inv.InventoryDate.Format("2006-01-02"),
			InventoryType: string(inv.InventoryType),
			Schedule:      schedule,
			StartedAt:     inv.StartedAt,
		})
	}

	resp := response.InProgressInventoriesResponse{
		Inventories: items,
		Count:       len(items),
	}

	respondJSON(w, http.StatusOK, resp)
}

// Create handles POST /api/inventories.
func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	employeeID, ok := middleware.GetEmployeeID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventory, err := h.inventoryService.CreateInventory(
		r.Context(),
		req.GetInventoryType(),
		req.GetSchedule(),
		req.GetDate(),
		employeeID,
	)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			respondError(w, appErr.Code, appErr.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create inventory")
		return
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := response.InventoryResponse{
		ID:            inventory.ID,
		InventoryDate: inventory.InventoryDate.Format("2006-01-02"),
		InventoryType: string(inventory.InventoryType),
		Schedule:      schedule,
		Status:        string(inventory.Status),
		ResponsibleID: inventory.ResponsibleID,
		StartedAt:     inventory.StartedAt,
	}

	respondJSON(w, http.StatusCreated, resp)
}

// GetItems handles GET /api/inventories/:id/items.
func (h *InventoryHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	inventory, details, err := h.inventoryService.GetInventoryItems(r.Context(), inventoryID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "inventory not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get inventory items")
		return
	}

	// Build response
	items := make([]response.InventoryItemResponse, 0, len(details))
	completedCount := 0
	for _, d := range details {
		item := response.InventoryItemResponse{
			ItemID:         d.ItemID,
			SuggestedValue: d.SuggestedValue,
			RealValue:      d.RealValue,
			StockReceived:  d.StockReceived,
			UnitsSold:      d.UnitsSold,
			IsComplete:     d.IsComplete(),
		}
		if d.Item != nil {
			item.Name = d.Item.Name
			if d.Item.Category != nil {
				item.CategoryName = d.Item.Category.Name
			}
		}
		if d.IsComplete() {
			completedCount++
		}
		items = append(items, item)
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := response.InventoryItemsResponse{
		InventoryID:    inventory.ID,
		InventoryType:  string(inventory.InventoryType),
		Schedule:       schedule,
		Date:           inventory.InventoryDate.Format("2006-01-02"),
		RequiresSales:  inventory.RequiresSalesAndPurchases(),
		TotalItems:     len(items),
		CompletedItems: completedCount,
		Items:          items,
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetDiscrepancies handles GET /api/inventories/:id/discrepancies.
func (h *InventoryHandler) GetDiscrepancies(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	inventory, details, err := h.inventoryService.GetDiscrepancies(r.Context(), inventoryID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "inventory not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get discrepancies")
		return
	}

	// Build response
	items := make([]response.DiscrepancyItemResponse, 0, len(details))
	for _, d := range details {
		var suggestedVal uint16
		if d.SuggestedValue != nil {
			suggestedVal = *d.SuggestedValue
		}

		item := response.DiscrepancyItemResponse{
			ItemID:         d.ItemID,
			SuggestedValue: suggestedVal,
			RealValue:      *d.RealValue,
			Difference:     d.Difference(),
			StockReceived:  d.StockReceived,
			UnitsSold:      d.UnitsSold,
		}
		if d.Item != nil {
			item.Name = d.Item.Name
		}
		items = append(items, item)
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := response.DiscrepanciesResponse{
		InventoryID:      inventory.ID,
		InventoryType:    string(inventory.InventoryType),
		Schedule:         schedule,
		Date:             inventory.InventoryDate.Format("2006-01-02"),
		RequiresSales:    inventory.RequiresSalesAndPurchases(),
		TotalItems:       len(items),
		HasDiscrepancies: len(items) > 0,
		Items:            items,
	}

	respondJSON(w, http.StatusOK, resp)
}

// SaveDetail handles POST /api/inventories/:id/details (physical count only).
func (h *InventoryHandler) SaveDetail(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	var req request.SaveInventoryDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	detail, err := h.inventoryService.SaveInventoryDetail(
		r.Context(),
		inventoryID,
		req.ItemID,
		req.RealValue,
	)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			respondError(w, appErr.Code, appErr.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to save detail")
		return
	}

	resp := response.SaveDetailResponse{
		Saved:          true,
		SuggestedValue: detail.SuggestedValue,
	}

	respondJSON(w, http.StatusOK, resp)
}

// SaveSales handles POST /api/inventories/:id/sales (sales and purchases).
func (h *InventoryHandler) SaveSales(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	var req request.SaveSalesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	detail, err := h.inventoryService.SaveSalesAndPurchases(
		r.Context(),
		inventoryID,
		req.ItemID,
		req.StockReceived,
		req.UnitsSold,
	)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			respondError(w, appErr.Code, appErr.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to save sales")
		return
	}

	resp := response.SaveDetailResponse{
		Saved:          true,
		SuggestedValue: detail.SuggestedValue,
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetSummary handles GET /api/inventories/:id/summary.
func (h *InventoryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	inventory, details, err := h.inventoryService.GetInventorySummary(r.Context(), inventoryID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "inventory not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get inventory summary")
		return
	}

	// Build response
	items := make([]response.InventorySummaryItem, 0, len(details))
	issuesCount := 0
	missingCount := 0

	for _, d := range details {
		if !d.IsComplete() {
			missingCount++
			continue
		}

		var suggestedVal uint16
		if d.SuggestedValue != nil {
			suggestedVal = *d.SuggestedValue
		}

		item := response.InventorySummaryItem{
			ItemID:         d.ItemID,
			SuggestedValue: suggestedVal,
			RealValue:      *d.RealValue,
			Difference:     d.Difference(),
			HasDiscrepancy: d.HasDiscrepancy(),
			StockReceived:  d.StockReceived,
			UnitsSold:      d.UnitsSold,
		}
		if d.Item != nil {
			item.Name = d.Item.Name
		}
		if d.HasDiscrepancy() {
			issuesCount++
		}
		items = append(items, item)
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := response.InventorySummaryResponse{
		InventoryID:     inventory.ID,
		InventoryType:   string(inventory.InventoryType),
		Schedule:        schedule,
		Date:            inventory.InventoryDate.Format("2006-01-02"),
		TotalItems:      len(details),
		ItemsWithIssues: issuesCount,
		Items:           items,
		CanComplete:     missingCount == 0,
		MissingItems:    missingCount,
	}

	respondJSON(w, http.StatusOK, resp)
}

// Complete handles POST /api/inventories/:id/complete.
func (h *InventoryHandler) Complete(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	issuesCreated, err := h.inventoryService.CompleteInventory(r.Context(), inventoryID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			respondError(w, appErr.Code, appErr.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to complete inventory")
		return
	}

	resp := response.CompleteInventoryResponse{
		Completed:     true,
		IssuesCreated: issuesCreated,
	}

	respondJSON(w, http.StatusOK, resp)
}

// getInventoryID extracts the inventory ID from the URL.
func (h *InventoryHandler) getInventoryID(r *http.Request) (uint32, error) {
	idStr := chi.URLParam(r, "inventoryID")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}
