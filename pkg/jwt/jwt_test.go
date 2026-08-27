package jwt

import (
	"testing"
	"time"
)

func TestJWTService(t *testing.T) {
	svc, err := NewService("test-secret-key-12345", 1*time.Hour)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	userID := int64(42)
	email := "test@example.com"
	role := "customer"

	tokenStr, err := svc.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := svc.ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("expected Role %s, got %s", role, claims.Role)
	}
}

func TestJWTInvalidSecret(t *testing.T) {
	_, err := NewService("", 1*time.Hour)
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}
