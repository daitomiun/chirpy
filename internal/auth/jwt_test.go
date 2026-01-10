package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userId := uuid.New()
	secret := "REMMY_WAS_HERE"
	signedString, err := MakeJWT(userId, secret, 5*time.Hour)
	if err != nil {
		t.Errorf(`Could not create JWT err: %v`, err)
	}
	validId, err := ValidateJWT(signedString, secret)
	if err != nil {
		t.Errorf(`JWT is not valid err: %v`, err)
	}
	if validId != userId {
		t.Errorf(`Ids do not match want: %s, have: %s`, userId, validId)
	}
}

func TestGetBearerToken(t *testing.T) {
	newHeader := http.Header{}
	newHeader.Set("Authorization", "Bearer elias")

	tokenString, err := GetBearerToken(newHeader)
	if err != nil {
		t.Errorf(`Token does not exist err: %v`, err)
	}
	if tokenString != "elias" {
		t.Errorf(`Invalid token want: 'elias', have: '%s'`, tokenString)
	}
}
