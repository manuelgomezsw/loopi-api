package http

import (
	"encoding/json"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
)

type ShiftHandler struct {
	shiftUseCase usecase.ShiftUseCase
}

func NewShiftHandler(shiftUseCase usecase.ShiftUseCase) *ShiftHandler {
	return &ShiftHandler{shiftUseCase: shiftUseCase}
}

func (h *ShiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	var shiftRequest domain.Shift
	if err := json.NewDecoder(r.Body).Decode(&shiftRequest); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	if err := h.shiftUseCase.CreateShift(shiftRequest); err != nil {
		BadRequest(w, err.Error())
		return
	}

	Created(w, map[string]string{"message": "Shift created"})
}
