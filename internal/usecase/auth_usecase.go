package usecase

import (
  "errors"
  "log"
  "loopi-api/config"
  "loopi-api/internal/repository"
)

type AuthUseCase interface {
  Login(email, password string) (string, error)
  SelectContext(userID int, franchiseID, storeID uint) (string, error)
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

  /*
  	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
  		return "", errors.New("invalid credentials")
  	}
  */
  token, err := config.GenerateJWT(
    int(user.ID),
    user.Email,
    "none",
    0,
    1,
  )
  if err != nil {
    return "", errors.New("could not generate token")
  }

  return token, nil
}

func (a *authUseCase) SelectContext(userID int, franchiseID, storeID uint) (string, error) {
  user, err := a.userRepo.FindByID(userID)
  if err != nil {
    return "", errors.New("user not found")
  }

  // Validar que pertenece a la franquicia
  var roleName string
  found := false

  for _, roles := range user.UserRoles {
    if roles.FranchiseID == franchiseID {
      roleName = roles.Role.Name
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
    roleName,
    franchiseID,
    storeID,
  )
  if err != nil {
    return "", errors.New("could not generate token")
  }

  return token, nil
}
