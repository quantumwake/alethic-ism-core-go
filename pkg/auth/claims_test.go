package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := []byte("test-secret-key-for-unit-tests")
	userID := "user-123"

	token, err := GenerateToken(secret, userID, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user_id=%s, got %s", userID, claims.UserID)
	}
}

func TestGenerateServiceToken(t *testing.T) {
	secret := []byte("test-secret-key")

	token, err := GenerateServiceToken(secret, "file-source-service")
	if err != nil {
		t.Fatalf("failed to generate service token: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("failed to parse service token: %v", err)
	}

	if claims.UserID != "file-source-service" {
		t.Errorf("expected user_id=file-source-service, got %s", claims.UserID)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken([]byte("secret-a"), "user-1", time.Hour)

	_, err := ParseToken([]byte("secret-b"), token)
	if err == nil {
		t.Fatal("expected error when parsing with wrong secret")
	}
}
