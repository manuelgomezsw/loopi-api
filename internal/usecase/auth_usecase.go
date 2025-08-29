package usecase

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"loopi-api/config"
	"loopi-api/internal/repository"
)

type AuthUseCase interface {
	Login(email, password string) (string, error)
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
		log.Fatalf("error: %v", err)
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Buscar el rol asociado a la franquicia seleccionada
	/*
	   var roleName string
	   var permissions []string
	   matchFound := false
	   for _, ur := range user.UserRoles {
	     if int(ur.FranchiseID) == franchiseID {
	       roleName = ur.Role.Name
	       for _, perm := range ur.Role.RolePermissions {
	         permissions = append(permissions, perm.Permission.Name)
	       }
	       matchFound = true
	       break
	     }
	   }
	   if !matchFound {
	     return "", errors.New("user does not belong to this franchise")
	   }
	*/
	token, err := config.GenerateJWT(
		int(user.ID),
		user.Email,
		"roleName",
		1,
	)
	if err != nil {
		return "", errors.New("could not generate token")
	}

	return token, nil
}
