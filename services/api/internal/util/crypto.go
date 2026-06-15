package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func RandomToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func HMACSHA256(secret, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func ConstantEqualHex(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}

func HashFingerprint(secret, value string) string {
	if value == "" {
		return ""
	}
	return HMACSHA256(secret, value)
}
