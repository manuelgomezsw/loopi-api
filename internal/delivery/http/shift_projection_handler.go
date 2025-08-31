package http

import (
	"encoding/json"
	"loopi-api/internal/usecase"
	"loopi-api/internal/usecase/dto"
	"net/http"
)

type ShiftProjectionHandler struct {
	shiftProjectionUseCase usecase.ShiftProjectionUseCase
}

func NewShiftProjectionHandler(shiftProjectionUseCase usecase.ShiftProjectionUseCase) *ShiftProjectionHandler {
	return &ShiftProjectionHandler{shiftProjectionUseCase: shiftProjectionUseCase}
}

// Preview Proyecta la cantidad de días ordinarios, horas extras, dominicales y festivas
func (h *ShiftProjectionHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var req dto.ShiftProjectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid input")
		return
	}

	result, err := h.shiftProjectionUseCase.PreviewHours(req)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	OK(w, result)
}
