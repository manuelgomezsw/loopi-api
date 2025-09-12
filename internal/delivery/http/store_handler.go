package http

import (
	"encoding/json"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type StoreHandler struct {
	storeUseCase usecase.StoreUseCase
}

func NewStoreHandler(storeUseCase usecase.StoreUseCase) *StoreHandler {
	return &StoreHandler{storeUseCase: storeUseCase}
}

func (h *StoreHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Check for query parameters to determine which operation to use
	franchiseID, _ := strconv.Atoi(r.URL.Query().Get("franchise"))
	activeOnly := r.URL.Query().Get("active") == "true"
	withEmployeeCount := r.URL.Query().Get("with_employee_count") == "true"

	if franchiseID > 0 && withEmployeeCount {
		// Get stores with employee count for franchise
		stores, err := h.storeUseCase.GetStoresWithEmployeeCount(franchiseID)
		if err != nil {
			rest.HandleError(w, err)
			return
		}
		rest.OK(w, stores)
		return
	}

	if franchiseID > 0 && activeOnly {
		// Get only active stores for franchise
		stores, err := h.storeUseCase.GetActiveStoresByFranchise(franchiseID)
		if err != nil {
			rest.HandleError(w, err)
			return
		}
		rest.OK(w, stores)
		return
	}

	if franchiseID > 0 {
		// Get all stores for franchise
		stores, err := h.storeUseCase.GetByFranchiseID(franchiseID)
		if err != nil {
			rest.HandleError(w, err)
			return
		}
		rest.OK(w, stores)
		return
	}

	// Get all stores
	stores, err := h.storeUseCase.GetAll()
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, stores)
}

func (h *StoreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		rest.BadRequest(w, "Invalid store ID")
		return
	}

	store, err := h.storeUseCase.GetByID(id)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, store)
}

func (h *StoreHandler) GetByFranchiseID(w http.ResponseWriter, r *http.Request) {
	franchiseIDStr := chi.URLParam(r, "franchiseID")
	franchiseID, err := strconv.Atoi(franchiseIDStr)
	if err != nil {
		rest.BadRequest(w, "Invalid franchise ID")
		return
	}

	// Check for active only parameter
	activeOnly := r.URL.Query().Get("active") == "true"

	var stores []domain.Store
	if activeOnly {
		stores, err = h.storeUseCase.GetActiveStoresByFranchise(franchiseID)
	} else {
		stores, err = h.storeUseCase.GetByFranchiseID(franchiseID)
	}

	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, stores)
}

func (h *StoreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var store domain.Store
	if err := json.NewDecoder(r.Body).Decode(&store); err != nil {
		rest.BadRequest(w, "Invalid JSON format")
		return
	}

	if err := h.storeUseCase.Create(&store); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.Created(w, store)
}

func (h *StoreHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		rest.BadRequest(w, "Invalid store ID")
		return
	}

	var store domain.Store
	if err := json.NewDecoder(r.Body).Decode(&store); err != nil {
		rest.BadRequest(w, "Invalid JSON format")
		return
	}

	store.ID = uint(id)
	if err := h.storeUseCase.Update(&store); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, store)
}

func (h *StoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		rest.BadRequest(w, "Invalid store ID")
		return
	}

	if err := h.storeUseCase.Delete(id); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.NoContent(w)
}

// GetStatistics retrieves comprehensive statistics for a store
func (h *StoreHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		rest.BadRequest(w, "Invalid store ID")
		return
	}

	statistics, err := h.storeUseCase.GetStoreStatistics(id)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, statistics)
}

// GetStoresWithEmployeeCount retrieves stores with their employee counts for a franchise
func (h *StoreHandler) GetStoresWithEmployeeCount(w http.ResponseWriter, r *http.Request) {
	franchiseIDStr := chi.URLParam(r, "franchiseID")
	franchiseID, err := strconv.Atoi(franchiseIDStr)
	if err != nil {
		rest.BadRequest(w, "Invalid franchise ID")
		return
	}

	stores, err := h.storeUseCase.GetStoresWithEmployeeCount(franchiseID)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, stores)
}

// GetActiveStoresByFranchise retrieves only active stores for a franchise
func (h *StoreHandler) GetActiveStoresByFranchise(w http.ResponseWriter, r *http.Request) {
	franchiseIDStr := chi.URLParam(r, "franchiseID")
	franchiseID, err := strconv.Atoi(franchiseIDStr)
	if err != nil {
		rest.BadRequest(w, "Invalid franchise ID")
		return
	}

	stores, err := h.storeUseCase.GetActiveStoresByFranchise(franchiseID)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, stores)
}
