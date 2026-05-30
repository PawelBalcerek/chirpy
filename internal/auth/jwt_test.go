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
	tokenSecret      = "7b9b0b47-b0da-4ea1-bbf4-accb187ae08a"
	expiresInSeconds = 30 * time.Second
)

func TestMakeJWT_ReturnsToken(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, expiresInSeconds)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if token == "" {
		t.Fatal("expected a non-empty token")
	}
}

func TestMakeJWT_ReturnsMissingTokenSecretError(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = auth.MakeJWT(userId, "", expiresInSeconds)
	if !errors.Is(err, auth.ErrMissingTokenSecret) {
		t.Fatalf("expected missing token secret error, got: %v", err)
	}
}

func TestValidateJWT_ReturnsUserId(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, expiresInSeconds)
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

func TestValidateJWT_ReturnsMissingTokenSecretError(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = auth.ValidateJWT(token, "")
	if !errors.Is(err, auth.ErrMissingTokenSecret) {
		t.Fatalf("expected missing token secret error, got: %v", err)
	}
}

func TestValidateJWT_ReturnsExpiredTokenError(t *testing.T) {
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
		t.Fatalf("expected expired token error, got: %v", err)
	}
}

func TestValidateJWT_ReturnsInvalidSignatureErrorWhenTokenSecretDiffers(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, expiresInSeconds)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = auth.ValidateJWT(token, "differentTokenSecret")
	if !errors.Is(err, jwt.ErrSignatureInvalid) {
		t.Fatalf("expected invalid signature error, got err: %v", err)
	}
}

func TestValidateJWT_ReturnsMalformedTokenError(t *testing.T) {
	_, err := auth.ValidateJWT("", tokenSecret)
	if !errors.Is(err, jwt.ErrTokenMalformed) {
		t.Fatalf("expected malformed token error, got: %v", err)
	}
}

func TestValidateJWT_ReturnsInvalidSignatureErrorWhenTamperedPayload(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	token, err := auth.MakeJWT(userId, tokenSecret, expiresInSeconds)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	tamperedToken := token + "junk"
	_, err = auth.ValidateJWT(tamperedToken, tokenSecret)
	if !errors.Is(err, jwt.ErrSignatureInvalid) {
		t.Fatalf("expected invalid signature error, got: %v", err)
	}
}

func TestValidateJWT_ReturnsInvalidIssuerError(t *testing.T) {
	userId, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	claims := jwt.RegisteredClaims{
		Issuer:    "wrong-issuer",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresInSeconds)),
		Subject:   userId.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = auth.ValidateJWT(tokenString, tokenSecret)
	if !errors.Is(err, auth.ErrInvalidIssuer) {
		t.Fatalf("expected invalid issuer error, got: %v", err)
	}
}
