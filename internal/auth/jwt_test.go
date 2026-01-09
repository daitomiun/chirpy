package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	id := uuid.New()
	secret := "REMMY_WAS_HERE"
	signedString, err := MakeJWT(id, secret, 5*time.Hour)
	if err != nil {
		t.Errorf(`Could not create JWT err: %v`, err)
	}
	fmt.Printf("Signed string: %s", signedString)
	validId, err := ValidateJWT(signedString, secret)
	if err != nil {
		t.Errorf(`JWT is not valid err: %v`, err)
	}
	if validId != id {
		t.Errorf(`Ids do not match want: %s, have: %s`, id, validId)
	}
}
