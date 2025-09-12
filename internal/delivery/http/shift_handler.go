package http

import (
	"encoding/json"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/middleware"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ShiftHandler struct {
	shiftUseCase usecase.ShiftUseCase
}

func NewShiftHandler(shiftUseCase usecase.ShiftUseCase) *ShiftHandler {
	return &ShiftHandler{shiftUseCase: shiftUseCase}
}

func (h *ShiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.Shift
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.shiftUseCase.Create(req); err != nil {
		rest.BadRequest(w, err.Error())
		return
	}

	rest.Created(w, map[string]string{"message": "Shift created"})
}

func (h *ShiftHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	storeID, _ := strconv.Atoi(r.URL.Query().Get("store"))
	activeOnly := r.URL.Query().Get("active") == "true"
	period := r.URL.Query().Get("period")

	var shifts []domain.Shift
	var err error

	// Handle different query scenarios
	if storeID > 0 && activeOnly {
		shifts, err = h.shiftUseCase.GetActiveShiftsByStore(storeID)
	} else if period != "" {
		shifts, err = h.shiftUseCase.GetShiftsByPeriod(period)
	} else if storeID > 0 {
		shifts, err = h.shiftUseCase.GetByStore(storeID)
	} else {
		shifts, err = h.shiftUseCase.GetAll()
	}

	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, shifts)
}

func (h *ShiftHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		rest.BadRequest(w, "Missing id")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		rest.BadRequest(w, "Invalid id")
		return
	}

	shift, err := h.shiftUseCase.GetByID(id)
	if err != nil || shift == nil {
		rest.NotFound(w, "Shift not found")
		return
	}

	// Validar store_id contra contexto
	contextStoreID := middleware.GetStoreID(r.Context())
	if contextStoreID != shift.StoreID {
		rest.Forbidden(w, "You do not have access to this shift")
		return
	}

	rest.OK(w, shift)
}

func (h *ShiftHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	storeIDStr := chi.URLParam(r, "store_id")
	storeID, err := strconv.Atoi(storeIDStr)
	if err != nil {
		rest.BadRequest(w, "invalid store_id")
		return
	}

	statistics, err := h.shiftUseCase.GetShiftStatistics(storeID)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, statistics)
}

func (h *ShiftHandler) GetByStore(w http.ResponseWriter, r *http.Request) {
	storeIDStr := chi.URLParam(r, "store_id")
	storeID, err := strconv.Atoi(storeIDStr)
	if err != nil {
		rest.BadRequest(w, "invalid store_id")
		return
	}

	// Check for active only parameter
	activeOnly := r.URL.Query().Get("active") == "true"

	var shifts []domain.Shift
	if activeOnly {
		shifts, err = h.shiftUseCase.GetActiveShiftsByStore(storeID)
	} else {
		shifts, err = h.shiftUseCase.GetByStore(storeID)
	}

	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, shifts)
}

func (h *ShiftHandler) GetByPeriod(w http.ResponseWriter, r *http.Request) {
	period := chi.URLParam(r, "period")
	if period == "" {
		rest.BadRequest(w, "Missing period parameter")
		return
	}

	shifts, err := h.shiftUseCase.GetShiftsByPeriod(period)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, shifts)
}
