package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateBlindIndex returns a stable HMAC-SHA-256 index for equality lookup.
func GenerateBlindIndex(value, key string) string {
	if value == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}