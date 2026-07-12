package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type customClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

type resetClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"target_user_id"`
}
