package http

import (
	"encoding/json"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
)

type FranchiseHandler struct {
	franchiseUseCase usecase.FranchiseUseCase
}

func NewFranchiseHandler(franchiseUseCase usecase.FranchiseUseCase) *FranchiseHandler {
	return &FranchiseHandler{franchiseUseCase: franchiseUseCase}
}

func (h *FranchiseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.Franchise
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}
	if err := h.franchiseUseCase.Create(req); err != nil {
		BadRequest(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	OK(w, map[string]string{"message": "Franchise created"})
}

func (h *FranchiseHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	franchises, err := h.franchiseUseCase.GetAll()
	if err != nil {
		ServerError(w, err.Error())
		return
	}

	OK(w, franchises)
}

func (h *FranchiseHandler) GetById(w http.ResponseWriter, r *http.Request) {
	franchiseID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	if franchiseID == 0 {
		BadRequest(w, "Missing parameters")
		return
	}

	franchise, err := h.franchiseUseCase.GetById(franchiseID)
	if err != nil {
		ServerError(w, err.Error())
		return
	}

	OK(w, franchise)
}
