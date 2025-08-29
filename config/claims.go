package config

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type CustomClaims struct {
	UserID      int    `json:"user_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	FranchiseID int    `json:"franchise_id"`
	jwt.RegisteredClaims
}

// NewClaims Helper para generar claims estándar
func NewClaims(userID int, email, role string, franchiseID int) CustomClaims {
	now := time.Now()
	return CustomClaims{
		UserID:      userID,
		Email:       email,
		Role:        role,
		FranchiseID: franchiseID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		},
	}
}
