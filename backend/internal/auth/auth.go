package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const Iterations = 210000

const keyLen = 32
const saltLen = 16

// HashPassword returns an encoded hash of the form
// "pbkdf2_sha256$<iterations>$<base64RawStdSalt>$<base64RawStdKey>".
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, Iterations, keyLen)
	if err != nil {
		return "", fmt.Errorf("derive key: %w", err)
	}
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s",
		Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether password matches encodedHash using a constant-time
// comparison. It returns false (never panics) for empty or malformed hashes.
func VerifyPassword(encodedHash, password string) bool {
	if encodedHash == "" {
		return false
	}
	parts := strings.SplitN(encodedHash, "$", 4)
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	storedKey, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(storedKey))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(key, storedKey) == 1
}

// NewSessionToken returns a URL-safe opaque token from 32 crypto/rand bytes,
// base64.RawURLEncoding encoded.
func NewSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns the lowercase hex-encoded SHA-256 of a session token,
// which is what gets persisted (never the raw token).
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
