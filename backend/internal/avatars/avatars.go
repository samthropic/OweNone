package avatars

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxBytes = 2 << 20 // 2 MiB
	PublicPathPrefix = "/api/v1/avatars/"
)

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// Storage persists profile photos on the local filesystem.
type Storage struct {
	Dir string
}

// NewStorage persists profile photos on the local filesystem.
// On read-only hosts (e.g. Vercel) it falls back to a temp directory.
func NewStorage(dir string) (*Storage, error) {
	candidates := make([]string, 0, 3)
	if dir != "" {
		candidates = append(candidates, dir)
	}
	candidates = append(candidates, "data/avatars", filepath.Join(os.TempDir(), "owenone-avatars"))

	var lastErr error
	for _, candidate := range candidates {
		if err := os.MkdirAll(candidate, 0o755); err != nil {
			lastErr = err
			continue
		}
		return &Storage{Dir: candidate}, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no avatar directory candidates")
	}
	return nil, fmt.Errorf("create avatar dir: %w", lastErr)
}

// PublicURL returns the stable public path for a user's avatar.
func PublicURL(userID string) string {
	return PublicPathPrefix + userID
}

// Save validates and writes an avatar for userID, removing any previous file.
func (storage *Storage) Save(userID string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty image")
	}
	if len(data) > MaxBytes {
		return "", fmt.Errorf("image must be %d bytes or smaller", MaxBytes)
	}
	contentType := http.DetectContentType(data)
	ext, ok := allowedTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type %q (use jpeg, png, webp, or gif)", contentType)
	}
	if err := storage.Remove(userID); err != nil {
		return "", err
	}
	path := filepath.Join(storage.Dir, userID+ext)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write avatar: %w", err)
	}
	return PublicURL(userID), nil
}

// Remove deletes any on-disk avatar for userID.
func (storage *Storage) Remove(userID string) error {
	matches, err := filepath.Glob(filepath.Join(storage.Dir, userID+".*"))
	if err != nil {
		return fmt.Errorf("list avatars: %w", err)
	}
	for _, match := range matches {
		if err := os.Remove(match); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove avatar: %w", err)
		}
	}
	return nil
}

// Open returns a readable file and content type for userID, or os.ErrNotExist.
func (storage *Storage) Open(userID string) (*os.File, string, error) {
	matches, err := filepath.Glob(filepath.Join(storage.Dir, userID+".*"))
	if err != nil {
		return nil, "", fmt.Errorf("list avatars: %w", err)
	}
	if len(matches) == 0 {
		return nil, "", os.ErrNotExist
	}
	file, err := os.Open(matches[0])
	if err != nil {
		return nil, "", err
	}
	ext := strings.ToLower(filepath.Ext(matches[0]))
	contentType := "application/octet-stream"
	for mime, allowedExt := range allowedTypes {
		if allowedExt == ext {
			contentType = mime
			break
		}
	}
	return file, contentType, nil
}

// ReadLimited reads at most MaxBytes+1 from r so callers can reject oversized bodies.
func ReadLimited(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, MaxBytes+1))
}
