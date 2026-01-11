package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	expiresAt := jwt.NewNumericDate(time.Now().UTC().Add(expiresIn))
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: expiresAt,
			Subject:   userID.String(),
		})
	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	type ChirpyClaims struct {
		jwt.RegisteredClaims
	}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&ChirpyClaims{},
		func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Name {
				return []byte(""), errors.New("Token is invalid")
			}
			return []byte(tokenSecret), nil
		})
	if err != nil {
		return uuid.Nil, err
	}
	subjectString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	subjectId, err := uuid.Parse(subjectString)
	if err != nil {
		return uuid.Nil, err
	}
	return subjectId, nil

}

func GetBearerToken(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")
	if len(bearer) == 0 {
		return "", errors.New("Invalid bearer token")
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer"))

	return tokenString, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)

	hexString := hex.EncodeToString(key)

	return hexString, nil
}

func GetApiKey(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")
	if len(bearer) == 0 {
		return "", errors.New("Invalid Api key")
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(bearer, "ApiKey"))

	return tokenString, nil
}
