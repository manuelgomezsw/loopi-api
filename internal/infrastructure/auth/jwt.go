package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
)

// Claims represents the JWT claims.
type Claims struct {
	EmployeeID uint16      `json:"employee_id"`
	Username   string      `json:"username"`
	Role       entity.Role `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token generation and validation.
type JWTManager struct {
	secret          []byte
	expirationHours int
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		secret:          []byte(cfg.Secret),
		expirationHours: cfg.ExpirationHours,
	}
}

// GenerateToken generates a new JWT token for an employee.
func (m *JWTManager) GenerateToken(employee *entity.Employee) (string, error) {
	now := time.Now()
	claims := Claims{
		EmployeeID: employee.ID,
		Username:   employee.Username,
		Role:       employee.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "loopi-api",
			Subject:   fmt.Sprintf("%d", employee.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims.
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
