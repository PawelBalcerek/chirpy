package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	issuer = "chirpy-access"
)

var (
	ErrMissingTokenSecret = errors.New("missing token secret")
	ErrInvalidIssuer      = errors.New("invalid issuer")
)

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	if tokenSecret == "" {
		return "", ErrMissingTokenSecret
	}

	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userId.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	if tokenSecret == "" {
		return uuid.Nil, ErrMissingTokenSecret
	}

	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	iss, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to retrieve issuer: %w", err)
	}
	if iss != issuer {
		return uuid.Nil, ErrInvalidIssuer
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to retrieve subject: %w", err)
	}
	userId, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to retrieve userId: %w", err)
	}

	return userId, nil
}
