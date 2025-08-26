package usecase

type FranchiseUseCase interface {
	Create() (string, error)
}

type franchiseUseCase struct {
	// Inyectar los repository
}

func NewFranchiseUseCase() FranchiseUseCase {
	return &franchiseUseCase{}
}

func (f *franchiseUseCase) Create() (string, error) {
	return "Franchise created", nil
}
