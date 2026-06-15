package util

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

var orderNoPattern = regexp.MustCompile(`^D\d{14}[A-F0-9]{6}$`)

func GenerateOrderNo(now time.Time) (string, error) {
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return "D" + now.UTC().Format("20060102150405") + strings.ToUpper(hex.EncodeToString(suffix[:])), nil
}

func IsOrderNo(value string) bool {
	return orderNoPattern.MatchString(value)
}
