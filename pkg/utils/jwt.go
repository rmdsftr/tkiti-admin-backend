package utils

import (
	"admin-panel/internal/config"
	"admin-panel/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Nim     string          `json:"nim"`
	NoAslab string          `json:"no_aslab"`
	Role    models.RoleEnum `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func GenerateTokenPair(cfg *config.Config, nim string, noAslab string, role models.RoleEnum) (*TokenPair, error) {
	accessToken, err := generateToken(
		cfg.JWT.Secret,
		cfg.JWT.ExpiresIn,
		nim,
		noAslab,
		role,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(
		cfg.JWT.RefreshSecret,
		cfg.JWT.RefreshExpiresIn,
		nim,
		noAslab,
		role,
	)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func generateToken(secret string, expiresIn time.Duration, nim string, noAslab string, role models.RoleEnum) (string, error) {
	claims := JWTClaims{
		Nim:     nim,
		NoAslab: noAslab,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
