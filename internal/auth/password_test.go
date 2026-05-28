package auth

import (
	"strings"
	"testing"
)

func TestHashPassword_HappyPath(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if hash == "" {
		t.Fatal("expected a non-empty hash")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("expected no error for empty password, got: %v", err)
	}
	if hash == "" {
		t.Fatal("expected a non-empty hash even for an empty password")
	}
}

func TestHashPassword_LongPassword(t *testing.T) {
	long := strings.Repeat("a", 1000)
	hash, err := HashPassword(long)
	if err != nil {
		t.Fatalf("expected no error for a long password, got: %v", err)
	}
	if hash == "" {
		t.Fatal("expected a non-empty hash for a long password")
	}
}

func TestHashPassword_IsUnique(t *testing.T) {
	password := "same-password"
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected two hashes of the same password to differ (random salt)")
	}
}

func TestCheckPasswordHash_CorrectPassword(t *testing.T) {
	password := "super-secret"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hashing failed: %v", err)
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !match {
		t.Error("expected password to match its hash")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hashing failed: %v", err)
	}

	match, err := CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("expected no error for a wrong password, got: %v", err)
	}
	if match {
		t.Error("expected wrong password NOT to match the hash")
	}
}

func TestCheckPasswordHash_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("non-empty-password")
	if err != nil {
		t.Fatalf("hashing failed: %v", err)
	}

	match, err := CheckPasswordHash("", hash)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match {
		t.Error("expected empty string NOT to match a non-empty password hash")
	}
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	match, err := CheckPasswordHash("any-password", "not-a-valid-hash")
	if err == nil {
		t.Error("expected an error for an invalid hash, got nil")
	}
	if match {
		t.Error("expected match to be false when hash is invalid")
	}
}

func TestCheckPasswordHash_EmptyHash(t *testing.T) {
	match, err := CheckPasswordHash("any-password", "")
	if err == nil {
		t.Error("expected an error for an empty hash, got nil")
	}
	if match {
		t.Error("expected match to be false for an empty hash")
	}
}
