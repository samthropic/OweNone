package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordVerifyRoundTrip(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !VerifyPassword(hash, "hunter2") {
		t.Error("VerifyPassword() = false, want true for correct password")
	}
}

func TestVerifyPasswordWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if VerifyPassword(hash, "wrong") {
		t.Error("VerifyPassword() = true, want false for wrong password")
	}
}

func TestHashPasswordRandomSalt(t *testing.T) {
	hash1, err := HashPassword("same")
	if err != nil {
		t.Fatalf("HashPassword() first call error = %v", err)
	}
	hash2, err := HashPassword("same")
	if err != nil {
		t.Fatalf("HashPassword() second call error = %v", err)
	}
	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes for same input, want different (random salt)")
	}
}

func TestVerifyPasswordMalformed(t *testing.T) {
	cases := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"wrong scheme", "bcrypt$100$abc$def"},
		{"non-numeric iterations", "pbkdf2_sha256$abc$abc$abc"},
		{"too few parts", "pbkdf2_sha256$100$abc"},
		{"bad salt base64", "pbkdf2_sha256$100$!!!$abc"},
		{"bad key base64", "pbkdf2_sha256$100$YWJj$!!!"},
		{"zero iterations", "pbkdf2_sha256$0$YWJj$YWJj"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if VerifyPassword(tc.hash, "anything") {
				t.Errorf("VerifyPassword(%q) = true, want false", tc.hash)
			}
		})
	}
}

func TestNewSessionTokenDistinct(t *testing.T) {
	tok1, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken() error = %v", err)
	}
	tok2, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken() error = %v", err)
	}
	if tok1 == tok2 {
		t.Error("NewSessionToken() returned identical tokens, want distinct")
	}
}

func TestHashTokenDeterministicAndLength(t *testing.T) {
	tok, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken() error = %v", err)
	}
	h1 := HashToken(tok)
	h2 := HashToken(tok)
	if h1 != h2 {
		t.Error("HashToken() is not deterministic")
	}
	if len(h1) != 64 {
		t.Errorf("HashToken() length = %d, want 64", len(h1))
	}
	if strings.ToLower(h1) != h1 {
		t.Errorf("HashToken() = %q, want lowercase hex", h1)
	}
}
