package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/dto/request"
	dtoresponse "github.com/manuelgomezsw/loopi-api/internal/application/dto/response"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/interface/middleware"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
)

// InventoryHandler handles employee inventory endpoints (my inventories flow).
type InventoryHandler struct {
	employeeService *service.EmployeeService
}

// NewInventoryHandler creates a new employee inventory handler.
func NewInventoryHandler(employeeService *service.EmployeeService) *InventoryHandler {
	return &InventoryHandler{employeeService: employeeService}
}

// GetSuggestedSchedule handles GET /api/inventories/suggested-schedule.
func (h *InventoryHandler) GetSuggestedSchedule(w http.ResponseWriter, r *http.Request) {
	suggested := h.employeeService.GetSuggestedSchedule()

	var schedule *string
	if suggested.Schedule != nil {
		s := string(*suggested.Schedule)
		schedule = &s
	}

	resp := dtoresponse.SuggestedScheduleResponse{
		InventoryType: string(suggested.InventoryType),
		Schedule:      schedule,
		Date:          suggested.Date.Format("2006-01-02"),
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// GetLatest handles GET /api/inventories/latest.
func (h *InventoryHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	inventory, err := h.employeeService.GetLatestCompleted(r.Context())
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to get latest inventory")
		return
	}

	if inventory == nil {
		response.RespondJSON(w, http.StatusOK, map[string]interface{}{"inventory": nil})
		return
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := dtoresponse.InventoryResponse{
		ID:            inventory.ID,
		InventoryDate: inventory.InventoryDate.Format("2006-01-02"),
		InventoryType: string(inventory.InventoryType),
		Schedule:      schedule,
		Status:        string(inventory.Status),
		ResponsibleID: inventory.ResponsibleID,
		StartedAt:     inventory.StartedAt,
		CompletedAt:   inventory.CompletedAt,
	}

	response.RespondJSON(w, http.StatusOK, map[string]interface{}{"inventory": resp})
}

// GetInProgress handles GET /api/inventories/in-progress.
func (h *InventoryHandler) GetInProgress(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := middleware.GetEmployeeID(r.Context())
	if !ok {
		response.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventories, err := h.employeeService.GetInProgress(r.Context(), employeeID)
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to get in-progress inventories")
		return
	}

	items := make([]dtoresponse.InProgressInventoryResponse, 0, len(inventories))
	for _, inv := range inventories {
		var schedule *string
		if inv.Schedule != nil {
			s := string(*inv.Schedule)
			schedule = &s
		}

		items = append(items, dtoresponse.InProgressInventoryResponse{
			ID:            inv.ID,
			InventoryDate: inv.InventoryDate.Format("2006-01-02"),
			InventoryType: string(inv.InventoryType),
			Schedule:      schedule,
			StartedAt:     inv.StartedAt,
		})
	}

	resp := dtoresponse.InProgressInventoriesResponse{
		Inventories: items,
		Count:       len(items),
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// Create handles POST /api/inventories.
func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	employeeID, ok := middleware.GetEmployeeID(r.Context())
	if !ok {
		response.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventory, err := h.employeeService.CreateInventory(
		r.Context(),
		req.GetInventoryType(),
		req.GetSchedule(),
		req.GetDate(),
		employeeID,
	)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.RespondError(w, appErr.Code, appErr.Message)
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "failed to create inventory")
		return
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := dtoresponse.InventoryResponse{
		ID:            inventory.ID,
		InventoryDate: inventory.InventoryDate.Format("2006-01-02"),
		InventoryType: string(inventory.InventoryType),
		Schedule:      schedule,
		Status:        string(inventory.Status),
		ResponsibleID: inventory.ResponsibleID,
		StartedAt:     inventory.StartedAt,
	}

	response.RespondJSON(w, http.StatusCreated, resp)
}

// GetItems handles GET /api/inventories/:id/items.
func (h *InventoryHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	inventory, details, err := h.employeeService.GetInventoryItems(r.Context(), inventoryID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			response.RespondError(w, http.StatusNotFound, "inventory not found")
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "failed to get inventory items")
		return
	}

	items := make([]dtoresponse.InventoryItemResponse, 0, len(details))
	completedCount := 0
	for _, d := range details {
		item := dtoresponse.InventoryItemResponse{
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

	resp := dtoresponse.InventoryItemsResponse{
		InventoryID:           inventory.ID,
		InventoryType:         string(inventory.InventoryType),
		Schedule:              schedule,
		Date:                  inventory.InventoryDate.Format("2006-01-02"),
		RequiresSales:         inventory.RequiresSalesAndPurchases(),
		RequiresPurchasesOnly: inventory.RequiresPurchasesOnly(),
		TotalItems:            len(items),
		CompletedItems:        completedCount,
		Items:                 items,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// GetDiscrepancies handles GET /api/inventories/:id/discrepancies.
func (h *InventoryHandler) GetDiscrepancies(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	inventory, details, err := h.employeeService.GetDiscrepancies(r.Context(), inventoryID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			response.RespondError(w, http.StatusNotFound, "inventory not found")
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "failed to get discrepancies")
		return
	}

	items := make([]dtoresponse.DiscrepancyItemResponse, 0, len(details))
	for _, d := range details {
		var suggestedVal uint16
		if d.SuggestedValue != nil {
			suggestedVal = *d.SuggestedValue
		}
		expectedAtEnd := h.employeeService.ComputeExpectedAtEnd(d)
		diff := int16(*d.RealValue) - int16(expectedAtEnd)

		item := dtoresponse.DiscrepancyItemResponse{
			ItemID:         d.ItemID,
			SuggestedValue: suggestedVal,
			RealValue:      *d.RealValue,
			Difference:     diff,
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

	resp := dtoresponse.DiscrepanciesResponse{
		InventoryID:           inventory.ID,
		InventoryType:         string(inventory.InventoryType),
		Schedule:              schedule,
		Date:                  inventory.InventoryDate.Format("2006-01-02"),
		RequiresSales:         inventory.RequiresSalesAndPurchases(),
		RequiresPurchasesOnly: inventory.RequiresPurchasesOnly(),
		TotalItems:            len(items),
		HasDiscrepancies:      len(items) > 0,
		Items:                 items,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// SaveDetail handles POST /api/inventories/:id/details.
func (h *InventoryHandler) SaveDetail(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	var req request.SaveInventoryDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	detail, err := h.employeeService.SaveInventoryDetail(
		r.Context(),
		inventoryID,
		req.ItemID,
		req.RealValue,
	)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.RespondError(w, appErr.Code, appErr.Message)
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "failed to save detail")
		return
	}

	resp := dtoresponse.SaveDetailResponse{
		Saved:          true,
		SuggestedValue: detail.SuggestedValue,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// SaveSales handles POST /api/inventories/:id/sales.
func (h *InventoryHandler) SaveSales(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	var req request.SaveSalesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	detail, err := h.employeeService.SaveSalesAndPurchases(
		r.Context(),
		inventoryID,
		req.ItemID,
		req.StockReceived,
		req.UnitsSold,
	)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.RespondError(w, appErr.Code, appErr.Message)
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "failed to save sales")
		return
	}

	resp := dtoresponse.SaveDetailResponse{
		Saved:          true,
		SuggestedValue: detail.SuggestedValue,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// GetSummary handles GET /api/inventories/:id/summary.
func (h *InventoryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	inventory, details, err := h.employeeService.GetInventorySummary(r.Context(), inventoryID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			response.RespondError(w, http.StatusNotFound, "inventory not found")
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "failed to get inventory summary")
		return
	}

	items := make([]dtoresponse.InventorySummaryItem, 0, len(details))
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
		expectedAtEnd := h.employeeService.ComputeExpectedAtEnd(d)
		diff := int16(*d.RealValue) - int16(expectedAtEnd)
		hasDisc := h.employeeService.HasDiscrepancyFromExpectedEnd(d)

		item := dtoresponse.InventorySummaryItem{
			ItemID:         d.ItemID,
			SuggestedValue: suggestedVal,
			RealValue:      *d.RealValue,
			Difference:     diff,
			HasDiscrepancy: hasDisc,
			StockReceived:  d.StockReceived,
			UnitsSold:      d.UnitsSold,
		}
		if d.Item != nil {
			item.Name = d.Item.Name
		}
		if hasDisc {
			issuesCount++
		}
		items = append(items, item)
	}

	var schedule *string
	if inventory.Schedule != nil {
		s := string(*inventory.Schedule)
		schedule = &s
	}

	resp := dtoresponse.InventorySummaryResponse{
		InventoryID:           inventory.ID,
		InventoryType:         string(inventory.InventoryType),
		Schedule:              schedule,
		Date:                  inventory.InventoryDate.Format("2006-01-02"),
		RequiresPurchasesOnly: inventory.RequiresPurchasesOnly(),
		TotalItems:            len(details),
		ItemsWithIssues:       issuesCount,
		Items:                 items,
		CanComplete:           missingCount == 0,
		MissingItems:          missingCount,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// Complete handles POST /api/inventories/:id/complete.
func (h *InventoryHandler) Complete(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := h.getInventoryID(r)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	issuesCreated, err := h.employeeService.CompleteInventory(r.Context(), inventoryID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.RespondError(w, appErr.Code, appErr.Message)
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "failed to complete inventory")
		return
	}

	resp := dtoresponse.CompleteInventoryResponse{
		Completed:     true,
		IssuesCreated: issuesCreated,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

func (h *InventoryHandler) getInventoryID(r *http.Request) (uint32, error) {
	idStr := chi.URLParam(r, "inventoryID")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}
