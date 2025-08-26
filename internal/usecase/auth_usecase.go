package usecase

import (
	"fmt"
	"loopi-api/config"
)

type AuthUseCase interface {
	Login(user, password string) (string, error)
}

type authUseCase struct {
	// Aquí podrías tener UserRepository inyectado
}

func NewAuthUseCase() AuthUseCase {
	return &authUseCase{}
}

func (a *authUseCase) Login(email, password string) (string, error) {
	// TODO: validar usuario contra la base
	if email != "admin@loopi.com" || password != "1234" {
		return "", fmt.Errorf("invalid credentials")
	}

	// Generar token
	token, err := config.GenerateJWT(1, email)
	if err != nil {
		return "", err
	}

	return token, nil
}
