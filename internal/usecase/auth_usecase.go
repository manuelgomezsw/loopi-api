package usecase

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"loopi-api/config"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type AuthUseCase interface {
	Login(email, password string) (string, error)
	SelectContext(userID int, franchiseID, storeID int) (string, error)
}

type authUseCase struct {
	userRepo repository.UserRepository
}

func NewAuthUseCase(userRepo repository.UserRepository) AuthUseCase {
	return &authUseCase{userRepo: userRepo}
}

func (a *authUseCase) Login(email, password string) (string, error) {
	user, err := a.userRepo.FindByEmail(email)

	if err != nil {
		log.Printf("error: %v\n", err)
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err := config.GenerateJWT(
		user.UserID,
		user.Email,
		getRoles(user),
		0,
		1,
	)
	if err != nil {
		return "", errors.New("could not generate token")
	}

	return token, nil
}

func (a *authUseCase) SelectContext(userID int, franchiseID, storeID int) (string, error) {
	user, err := a.userRepo.FindByID(userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Validar que pertenece a la franquicia
	found := false
	for _, roles := range user.UserRoles {
		if roles.FranchiseID == franchiseID {
			found = true
			break
		}
	}

	if !found {
		return "", errors.New("user does not belong to this franchise")
	}

	token, err := config.GenerateJWT(
		userID,
		user.Email,
		getRoles(user),
		franchiseID,
		storeID,
	)
	if err != nil {
		return "", errors.New("could not generate token")
	}

	return token, nil
}

func getRoles(user *domain.User) []string {
	roles := make([]string, 0)
	for _, ur := range user.UserRoles {
		roles = append(roles, ur.Role.Name)
	}
	return roles
}
