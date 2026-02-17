package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
)

// ItemHandler handles admin item endpoints (CRUD, measurement units, active count).
type ItemHandler struct {
	item *service.AdminItemService
}

// NewItemHandler creates a new admin item handler.
func NewItemHandler(item *service.AdminItemService) *ItemHandler {
	return &ItemHandler{item: item}
}

// ListItems returns a paginated list of items with filters.
func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	filter := service.ItemFilter{Page: 1, PageSize: 20}

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
	if itemType := r.URL.Query().Get("type"); itemType != "" {
		it := entity.ItemType(itemType)
		filter.Type = &it
	}
	if freq := r.URL.Query().Get("frequency"); freq != "" {
		f := entity.InventoryFrequency(freq)
		filter.Frequency = &f
	}
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active := activeStr == "true"
		filter.Active = &active
	}
	filter.Search = r.URL.Query().Get("search")

	result, err := h.item.ListItems(r.Context(), filter)
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list items")
		return
	}

	response.RespondJSON(w, http.StatusOK, result)
}

// GetItem returns a single item by ID.
func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "itemID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	item, err := h.item.GetItem(r.Context(), uint16(id))
	if err != nil {
		response.RespondError(w, http.StatusNotFound, "item not found")
		return
	}

	response.RespondJSON(w, http.StatusOK, item)
}

// ListMeasurementUnits returns all measurement units for item creation/edition.
func (h *ItemHandler) ListMeasurementUnits(w http.ResponseWriter, r *http.Request) {
	units, err := h.item.ListMeasurementUnits(r.Context())
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list measurement units")
		return
	}
	if units == nil {
		units = []*entity.MeasurementUnit{}
	}
	response.RespondJSON(w, http.StatusOK, units)
}

// CreateItemRequest is the request body for creating an item.
type CreateItemRequest struct {
	Type                   string  `json:"type"`
	Name                   string  `json:"name"`
	InventoryFrequency     string  `json:"inventory_frequency"`
	CategoryID             uint16  `json:"category_id"`
	SupplierID             *uint16 `json:"supplier_id"`
	Cost                   uint32  `json:"cost"`
	MeasurementUnitID      uint16  `json:"measurement_unit_id"`
	AddToActiveInventories bool    `json:"add_to_active_inventories"`
}

// CreateItem creates a new item.
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	serviceReq := service.CreateItemRequest{
		Type:                   entity.ItemType(req.Type),
		Name:                   req.Name,
		InventoryFrequency:     entity.InventoryFrequency(req.InventoryFrequency),
		CategoryID:             req.CategoryID,
		SupplierID:             req.SupplierID,
		Cost:                   req.Cost,
		MeasurementUnitID:      req.MeasurementUnitID,
		AddToActiveInventories: req.AddToActiveInventories,
	}

	item, err := h.item.CreateItem(r.Context(), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusCreated, item)
}

// UpdateItemRequest is the request body for updating an item.
type UpdateItemRequest struct {
	Type               string  `json:"type"`
	Name               string  `json:"name"`
	InventoryFrequency string  `json:"inventory_frequency"`
	Active             bool    `json:"active"`
	CategoryID         uint16  `json:"category_id"`
	SupplierID         *uint16 `json:"supplier_id"`
	Cost               uint32  `json:"cost"`
	MeasurementUnitID  uint16  `json:"measurement_unit_id"`
}

// UpdateItem updates an existing item.
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "itemID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	serviceReq := service.UpdateItemRequest{
		Type:               entity.ItemType(req.Type),
		Name:               req.Name,
		InventoryFrequency: entity.InventoryFrequency(req.InventoryFrequency),
		Active:             req.Active,
		CategoryID:         req.CategoryID,
		SupplierID:         req.SupplierID,
		Cost:               req.Cost,
		MeasurementUnitID:  req.MeasurementUnitID,
	}

	item, err := h.item.UpdateItem(r.Context(), uint16(id), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, item)
}

// UpdateItemStatusRequest is the request body for updating item status.
type UpdateItemStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateItemStatus updates the active status of an item.
func (h *ItemHandler) UpdateItemStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "itemID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	var req UpdateItemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.item.UpdateItemStatus(r.Context(), uint16(id), req.Active); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}
