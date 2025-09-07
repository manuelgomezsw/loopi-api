package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/middleware"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
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
	var shifts []domain.Shift
	var err error

	if storeID > 0 {
		shifts, err = h.shiftUseCase.GetByStore(storeID)
	} else {
		shifts, err = h.shiftUseCase.GetAll()
	}

	if err != nil {
		rest.ServerError(w, err.Error())
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

func (h *ShiftHandler) GetByStore(w http.ResponseWriter, r *http.Request) {
	storeIDStr := chi.URLParam(r, "store_id")
	storeID, err := strconv.Atoi(storeIDStr)
	if err != nil {
		rest.BadRequest(w, "invalid store_id")
		return
	}

	shifts, err := h.shiftUseCase.GetByStore(storeID)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, shifts)
}
