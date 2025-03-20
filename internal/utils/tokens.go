package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

// GenerateSecureToken generates a cryptographically secure random token.
// The length parameter specifies the number of random bytes; the resulting token is twice as long in hex.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
