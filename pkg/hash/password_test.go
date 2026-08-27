package hash

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "my-secret-password-123"

	hashed, err := Generate(password)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if hashed == "" {
		t.Fatal("expected non-empty hashed password")
	}

	if !Compare(hashed, password) {
		t.Fatal("expected password to match hash")
	}

	if Compare(hashed, "wrong-password") {
		t.Fatal("expected wrong password to fail comparison")
	}
}
