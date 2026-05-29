package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenSecret = "7b9b0b47-b0da-4ea1-bbf4-accb187ae08a"
	ExpiresIn   = 30 * time.Second
)

func TestMakeJWT_HappyPath(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if token == "" {
		t.Fatal("expected a non-empty token")
	}
}

func TestValidateJWT_ValidToken(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	resultUserId, err := auth.ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if userId != resultUserId {
		t.Fatalf("expected %s, got %s", userId.String(), resultUserId.String())
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = auth.ValidateJWT(token, tokenSecret)
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected expired token, got: %v", err)
	}
}

func TestValidateJWT_ValidTokenInvalidTokenSecret(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = auth.ValidateJWT(token, "")
	if !errors.Is(err, jwt.ErrSignatureInvalid) {
		t.Fatalf("expected invalid signature, got err: %v", err)
	}
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	_, err := auth.ValidateJWT("", tokenSecret)
	if !errors.Is(err, jwt.ErrTokenMalformed) {
		t.Fatalf("expected malformed token, got: %v", err)
	}
}

func TestValidateJWT_TamperedPayload(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	tamperedToken := token + "junk"
	_, err = auth.ValidateJWT(tamperedToken, tokenSecret)
	if !errors.Is(err, jwt.ErrSignatureInvalid) {
		t.Fatalf("expected invalid signature, got: %v", err)
	}
}

func TestValidateJWT_InvalidIssuer(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	claims := jwt.RegisteredClaims{
		Issuer:    "wrong-issuer",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ExpiresIn)),
		Subject:   userId.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = auth.ValidateJWT(tokenString, tokenSecret)
	if !errors.Is(err, auth.ErrInvalidIssuer) {
		t.Fatalf("expected invalid issuer, got: %v", err)
	}
}
