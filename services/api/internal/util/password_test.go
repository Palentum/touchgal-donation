package util

import (
	"strings"
	"testing"
)

func TestPasswordHashVerify(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("unexpected hash prefix: %s", hash)
	}
	if !VerifyPassword("correct horse battery staple", hash) {
		t.Fatalf("password did not verify")
	}
	if VerifyPassword("wrong", hash) {
		t.Fatalf("wrong password verified")
	}
}
