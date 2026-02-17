package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
)

// SupplierHandler handles admin supplier endpoints (CRUD, list active).
type SupplierHandler struct {
	supplier *service.AdminSupplierService
}

// NewSupplierHandler creates a new admin supplier handler.
func NewSupplierHandler(supplier *service.AdminSupplierService) *SupplierHandler {
	return &SupplierHandler{supplier: supplier}
}

// ListSuppliers returns a paginated list of suppliers with filters.
func (h *SupplierHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	filter := service.SupplierFilter{Page: 1, PageSize: 20}

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

	result, err := h.supplier.ListSuppliers(r.Context(), filter)
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list suppliers")
		return
	}

	response.RespondJSON(w, http.StatusOK, result)
}

// ListAllActiveSuppliers returns all active suppliers for dropdowns.
func (h *SupplierHandler) ListAllActiveSuppliers(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.supplier.ListAllActiveSuppliers(r.Context())
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list suppliers")
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"suppliers": suppliers,
		"total":     len(suppliers),
	})
}

// GetSupplier returns a single supplier by ID.
func (h *SupplierHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "supplierID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid supplier ID")
		return
	}

	supplier, err := h.supplier.GetSupplier(r.Context(), uint16(id))
	if err != nil {
		response.RespondError(w, http.StatusNotFound, "supplier not found")
		return
	}

	response.RespondJSON(w, http.StatusOK, supplier)
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
func (h *SupplierHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	serviceReq := service.CreateSupplierRequest{
		BusinessName: req.BusinessName,
		TaxID:        req.TaxID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
	}

	supplier, err := h.supplier.CreateSupplier(r.Context(), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusCreated, supplier)
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
func (h *SupplierHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "supplierID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid supplier ID")
		return
	}

	var req UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
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

	supplier, err := h.supplier.UpdateSupplier(r.Context(), uint16(id), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, supplier)
}

// UpdateSupplierStatusRequest is the request body for updating supplier status.
type UpdateSupplierStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateSupplierStatus updates the active status of a supplier.
func (h *SupplierHandler) UpdateSupplierStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "supplierID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid supplier ID")
		return
	}

	var req UpdateSupplierStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.supplier.UpdateSupplierStatus(r.Context(), uint16(id), req.Active); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}
