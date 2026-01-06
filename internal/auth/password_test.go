package auth

import (
	"testing"
)

func TestHashPassord(t *testing.T) {
	password := "REMNANT"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf(`Error trying to hash password, err: %v`, err)
	}
	match, err := CheckPasswordHash(password, hash)
	if !match || err != nil {
		t.Errorf(`Hash does not match -> %s with password -> %s`, hash, password)
	}
}
