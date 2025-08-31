package http

import (
	"encoding/json"
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
		BadRequest(w, "Invalid request body")
		return
	}
	if err := h.shiftUseCase.Create(req); err != nil {
		BadRequest(w, err.Error())
		return
	}
	Created(w, map[string]string{"message": "Shift created"})
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
		ServerError(w, err.Error())
		return
	}

	OK(w, shifts)
}

func (h *ShiftHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		BadRequest(w, "Missing id")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		BadRequest(w, "Invalid id")
		return
	}

	shift, err := h.shiftUseCase.GetByID(id)
	if err != nil || shift == nil {
		NotFound(w, "Shift not found")
		return
	}

	// Validar store_id contra contexto
	contextStoreID := middleware.GetStoreID(r.Context())
	if contextStoreID != shift.StoreID {
		Forbidden(w, "You do not have access to this shift")
		return
	}

	OK(w, shift)
}
