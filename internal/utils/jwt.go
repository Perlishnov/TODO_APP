package utils

import (
	"errors"
	"time"

	"github.com/Perlishnov/TODO_APP/internal/config"
	"github.com/Perlishnov/TODO_APP/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil struct {
	secret []byte
	exp    time.Duration
}

// JWTService defines the methods needed for JWT operations
type JWTService interface {
	GenerateToken(user models.User) (string, error)
	ValidateToken(token string) (*Claims, error)
}

func NewJWTUtil(cfg *config.Config) *JWTUtil {
	return &JWTUtil{
		secret: []byte(cfg.JWTSecret),
		exp:    time.Duration(cfg.JWTExpirationHours) * time.Hour,
	}
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (j *JWTUtil) GenerateToken(user models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTUtil) ValidateToken(RecievedToken string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(RecievedToken, claims, func(t *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("Invalid token")
	}
	return claims, nil
}
