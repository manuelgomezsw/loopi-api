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

// --- Item Handlers ---

// ListItems returns a paginated list of items with filters.
func (h *AdminHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	filter := service.ItemFilter{
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

	result, err := h.adminService.ListItems(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list items"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetItem returns a single item by ID.
func (h *AdminHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "itemID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid item ID"}`, http.StatusBadRequest)
		return
	}

	item, err := h.adminService.GetItem(r.Context(), uint16(id))
	if err != nil {
		http.Error(w, `{"error":"item not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// CreateItemRequest is the request body for creating an item.
type CreateItemRequest struct {
	Type                   string  `json:"type"`
	Name                   string  `json:"name"`
	InventoryFrequency     string  `json:"inventory_frequency"`
	CategoryID             uint16  `json:"category_id"`
	SupplierID             *uint16 `json:"supplier_id"`
	Cost                   uint32  `json:"cost"`
	AddToActiveInventories bool    `json:"add_to_active_inventories"`
}

// CreateItem creates a new item.
func (h *AdminHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	serviceReq := service.CreateItemRequest{
		Type:                   entity.ItemType(req.Type),
		Name:                   req.Name,
		InventoryFrequency:     entity.InventoryFrequency(req.InventoryFrequency),
		CategoryID:             req.CategoryID,
		SupplierID:             req.SupplierID,
		Cost:                   req.Cost,
		AddToActiveInventories: req.AddToActiveInventories,
	}

	item, err := h.adminService.CreateItem(r.Context(), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
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
}

// UpdateItem updates an existing item.
func (h *AdminHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "itemID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid item ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
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
	}

	item, err := h.adminService.UpdateItem(r.Context(), uint16(id), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// UpdateItemStatusRequest is the request body for updating item status.
type UpdateItemStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateItemStatus updates the active status of an item.
func (h *AdminHandler) UpdateItemStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "itemID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid item ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateItemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.adminService.UpdateItemStatus(r.Context(), uint16(id), req.Active); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// GetActiveInventoriesCount returns the count of in-progress inventories.
func (h *AdminHandler) GetActiveInventoriesCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.adminService.GetActiveInventoriesCount(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get active inventories count")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"count": count})
}

// --- Employee Handlers ---

// ListEmployees returns a paginated list of employees with filters.
func (h *AdminHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	filter := service.EmployeeFilter{
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
	if role := r.URL.Query().Get("role"); role != "" {
		r := entity.Role(role)
		filter.Role = &r
	}
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active := activeStr == "true"
		filter.Active = &active
	}
	filter.Search = r.URL.Query().Get("search")

	result, err := h.adminService.ListEmployees(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list employees"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetEmployee returns a single employee by ID.
func (h *AdminHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid employee ID"}`, http.StatusBadRequest)
		return
	}

	employee, err := h.adminService.GetEmployee(r.Context(), uint16(id))
	if err != nil {
		http.Error(w, `{"error":"employee not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}

// CreateEmployeeRequest is the request body for creating an employee.
type CreateEmployeeRequest struct {
	Username       string  `json:"username"`
	Password       string  `json:"password"`
	Name           string  `json:"name"`
	LastName       string  `json:"last_name"`
	DocumentType   *string `json:"document_type"`
	DocumentNumber *string `json:"document_number"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	BirthDate      *string `json:"birth_date"` // Format: YYYY-MM-DD
	Role           string  `json:"role"`
}

// CreateEmployee creates a new employee.
func (h *AdminHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	var birthDate *time.Time
	if req.BirthDate != nil && *req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			http.Error(w, `{"error":"invalid birth_date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
			return
		}
		birthDate = &t
	}

	serviceReq := service.CreateEmployeeRequest{
		Username:       req.Username,
		Password:       req.Password,
		Name:           req.Name,
		LastName:       req.LastName,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		Phone:          req.Phone,
		Email:          req.Email,
		BirthDate:      birthDate,
		Role:           entity.Role(req.Role),
	}

	employee, err := h.adminService.CreateEmployee(r.Context(), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(employee)
}

// UpdateEmployeeRequest is the request body for updating an employee.
type UpdateEmployeeRequest struct {
	Username       string  `json:"username"`
	Name           string  `json:"name"`
	LastName       string  `json:"last_name"`
	DocumentType   *string `json:"document_type"`
	DocumentNumber *string `json:"document_number"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	BirthDate      *string `json:"birth_date"` // Format: YYYY-MM-DD
	Role           string  `json:"role"`
	Active         bool    `json:"active"`
}

// UpdateEmployee updates an existing employee.
func (h *AdminHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid employee ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	var birthDate *time.Time
	if req.BirthDate != nil && *req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			http.Error(w, `{"error":"invalid birth_date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
			return
		}
		birthDate = &t
	}

	serviceReq := service.UpdateEmployeeRequest{
		Username:       req.Username,
		Name:           req.Name,
		LastName:       req.LastName,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		Phone:          req.Phone,
		Email:          req.Email,
		BirthDate:      birthDate,
		Role:           entity.Role(req.Role),
		Active:         req.Active,
	}

	employee, err := h.adminService.UpdateEmployee(r.Context(), uint16(id), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}

// UpdateEmployeeStatusRequest is the request body for updating employee status.
type UpdateEmployeeStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateEmployeeStatus updates the active status of an employee.
func (h *AdminHandler) UpdateEmployeeStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid employee ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateEmployeeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.adminService.UpdateEmployeeStatus(r.Context(), uint16(id), req.Active); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// ResetEmployeePassword resets an employee's password to the default.
func (h *AdminHandler) ResetEmployeePassword(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid employee ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.adminService.ResetEmployeePassword(r.Context(), uint16(id)); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// --- Category Handlers ---

// ListCategories returns all categories ordered by display_order.
func (h *AdminHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.adminService.ListCategories(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list categories"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
		"total":      len(categories),
	})
}

// GetCategory returns a single category by ID.
func (h *AdminHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "categoryID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid category ID"}`, http.StatusBadRequest)
		return
	}

	category, err := h.adminService.GetCategory(r.Context(), uint16(id))
	if err != nil {
		http.Error(w, `{"error":"category not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// CreateCategoryRequest is the request body for creating a category.
type CreateCategoryRequest struct {
	Name string `json:"name"`
}

// CreateCategory creates a new category.
func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	serviceReq := service.CreateCategoryRequest{
		Name: req.Name,
	}

	category, err := h.adminService.CreateCategory(r.Context(), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// UpdateCategoryRequest is the request body for updating a category.
type UpdateCategoryRequest struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// UpdateCategory updates an existing category.
func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "categoryID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid category ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	serviceReq := service.UpdateCategoryRequest{
		Name:   req.Name,
		Active: req.Active,
	}

	category, err := h.adminService.UpdateCategory(r.Context(), uint16(id), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// UpdateCategoryStatusRequest is the request body for updating category status.
type UpdateCategoryStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateCategoryStatus updates the active status of a category.
func (h *AdminHandler) UpdateCategoryStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "categoryID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid category ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateCategoryStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.adminService.UpdateCategoryStatus(r.Context(), uint16(id), req.Active); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// ReorderCategoriesRequest is the request body for reordering categories.
type ReorderCategoriesRequest struct {
	Orders []struct {
		ID           uint16 `json:"id"`
		DisplayOrder int    `json:"display_order"`
	} `json:"orders"`
}

// ReorderCategories updates the display order of multiple categories.
func (h *AdminHandler) ReorderCategories(w http.ResponseWriter, r *http.Request) {
	var req ReorderCategoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	serviceReq := service.ReorderCategoryRequest{
		Orders: make([]service.CategoryOrderItem, len(req.Orders)),
	}
	for i, o := range req.Orders {
		serviceReq.Orders[i] = service.CategoryOrderItem{
			ID:           o.ID,
			DisplayOrder: o.DisplayOrder,
		}
	}

	if err := h.adminService.ReorderCategories(r.Context(), serviceReq); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// --- Supplier Handlers ---

// ListSuppliers returns a paginated list of suppliers with filters.
func (h *AdminHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	filter := service.SupplierFilter{
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
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active := activeStr == "true"
		filter.Active = &active
	}
	filter.Search = r.URL.Query().Get("search")

	result, err := h.adminService.ListSuppliers(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list suppliers"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListAllActiveSuppliers returns all active suppliers for dropdowns.
func (h *AdminHandler) ListAllActiveSuppliers(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.adminService.ListAllActiveSuppliers(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list suppliers"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"suppliers": suppliers,
		"total":     len(suppliers),
	})
}

// GetSupplier returns a single supplier by ID.
func (h *AdminHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "supplierID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid supplier ID"}`, http.StatusBadRequest)
		return
	}

	supplier, err := h.adminService.GetSupplier(r.Context(), uint16(id))
	if err != nil {
		http.Error(w, `{"error":"supplier not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(supplier)
}

// CreateSupplierRequest is the request body for creating a supplier.
type CreateSupplierRequest struct {
	BusinessName string `json:"business_name"`
	TaxID        string `json:"tax_id"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
}

// CreateSupplier creates a new supplier.
func (h *AdminHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	serviceReq := service.CreateSupplierRequest{
		BusinessName: req.BusinessName,
		TaxID:        req.TaxID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
	}

	supplier, err := h.adminService.CreateSupplier(r.Context(), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(supplier)
}

// UpdateSupplierRequest is the request body for updating a supplier.
type UpdateSupplierRequest struct {
	BusinessName string `json:"business_name"`
	TaxID        string `json:"tax_id"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
	Active       bool   `json:"active"`
}

// UpdateSupplier updates an existing supplier.
func (h *AdminHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "supplierID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid supplier ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	serviceReq := service.UpdateSupplierRequest{
		BusinessName: req.BusinessName,
		TaxID:        req.TaxID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Active:       req.Active,
	}

	supplier, err := h.adminService.UpdateSupplier(r.Context(), uint16(id), serviceReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(supplier)
}

// UpdateSupplierStatusRequest is the request body for updating supplier status.
type UpdateSupplierStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateSupplierStatus updates the active status of a supplier.
func (h *AdminHandler) UpdateSupplierStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "supplierID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		http.Error(w, `{"error":"invalid supplier ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateSupplierStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.adminService.UpdateSupplierStatus(r.Context(), uint16(id), req.Active); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}
