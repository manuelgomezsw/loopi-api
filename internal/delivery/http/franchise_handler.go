package http

import (
  "loopi-api/internal/usecase"
  "net/http"
)

type FranchiseHandler struct {
  franchiseUseCase usecase.FranchiseUseCase
}

func NewFranchiseHandler(franchiseUseCase usecase.FranchiseUseCase) *FranchiseHandler {
  return &FranchiseHandler{franchiseUseCase: franchiseUseCase}
}

func (h *FranchiseHandler) Create(w http.ResponseWriter, r *http.Request) {
  response, err := h.franchiseUseCase.Create()
  if err != nil {
    ServerError(w, err.Error())
  }

  OK(w, response)
}
