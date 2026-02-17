package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
)

// CategoryHandler handles admin category endpoints (CRUD, reorder).
type CategoryHandler struct {
	category *service.AdminCategoryService
}

// NewCategoryHandler creates a new admin category handler.
func NewCategoryHandler(category *service.AdminCategoryService) *CategoryHandler {
	return &CategoryHandler{category: category}
}

// ListCategories returns all categories ordered by display_order.
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.category.ListCategories(r.Context())
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
		"total":      len(categories),
	})
}

// GetCategory returns a single category by ID.
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "categoryID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	category, err := h.category.GetCategory(r.Context(), uint16(id))
	if err != nil {
		response.RespondError(w, http.StatusNotFound, "category not found")
		return
	}

	response.RespondJSON(w, http.StatusOK, category)
}

// CreateCategoryRequest is the request body for creating a category.
type CreateCategoryRequest struct {
	Name string `json:"name"`
}

// CreateCategory creates a new category.
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	serviceReq := service.CreateCategoryRequest{Name: req.Name}

	category, err := h.category.CreateCategory(r.Context(), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusCreated, category)
}

// UpdateCategoryRequest is the request body for updating a category.
type UpdateCategoryRequest struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// UpdateCategory updates an existing category.
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "categoryID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	serviceReq := service.UpdateCategoryRequest{Name: req.Name, Active: req.Active}

	category, err := h.category.UpdateCategory(r.Context(), uint16(id), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, category)
}

// UpdateCategoryStatusRequest is the request body for updating category status.
type UpdateCategoryStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateCategoryStatus updates the active status of a category.
func (h *CategoryHandler) UpdateCategoryStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "categoryID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	var req UpdateCategoryStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.category.UpdateCategoryStatus(r.Context(), uint16(id), req.Active); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ReorderCategoriesRequest is the request body for reordering categories.
type ReorderCategoriesRequest struct {
	Orders []struct {
		ID           uint16 `json:"id"`
		DisplayOrder int    `json:"display_order"`
	} `json:"orders"`
}

// ReorderCategories updates the display order of multiple categories.
func (h *CategoryHandler) ReorderCategories(w http.ResponseWriter, r *http.Request) {
	var req ReorderCategoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
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

	if err := h.category.ReorderCategories(r.Context(), serviceReq); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}
