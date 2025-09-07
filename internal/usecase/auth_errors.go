package usecase

import appErr "loopi-api/internal/common/errors"

var (
	ErrUserInvalidCredentials    = appErr.NewDomainError(401, "invalid credentials")
	ErrUserNotFound              = appErr.NewDomainError(404, "user not found")
	ErrUserCouldNotGenerateToken = appErr.NewDomainError(500, "could not generate token")
)
