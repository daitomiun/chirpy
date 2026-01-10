package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/daitonium/chirpy/internal/auth"
	"github.com/daitonium/chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
	platform       string
	secret         string
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("charset", "utf-8")
	r.Header.Add("Content-Type", "text/plain")

	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Cannot access resource", nil)
		return
	}

	err := cfg.database.DeleteUsers(context.Background())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something happened while trying to reset users", err)
		return
	}

	cfg.fileserverHits.Swap(0)
	fmt.Printf("Counter reset to %v \n", cfg.fileserverHits.Load())
	w.WriteHeader(200)
	w.Write([]byte("Ok"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("charset", "utf-8")
	r.Header.Add("Content-Type", "text/html")
	w.WriteHeader(200)
	panel := fmt.Sprintf("<html> <body> <h1>Welcome, Chirpy Admin</h1> <p>Chirpy has been visited %d times!</p> </body> </html>", cfg.fileserverHits.Load())
	var bytes []byte
	w.Write(fmt.Append(bytes, panel))
}

func handlerCheckHealth(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("charset", "utf-8")
	r.Header.Add("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

type chirpParams struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

func isValidChirp(chirp *chirpParams, w http.ResponseWriter, r *http.Request) bool {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirp); err != nil {
		log.Printf("Error decoding chirp: %s  \n", err)
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return false
	}
	if len(chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirpy is too long", nil)
		return false
	}
	return true
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorRes struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorRes{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
	return
}

func replaceBadWords(sentence string) string {
	words := strings.Split(sentence, " ")

	badWords := [3]string{"kerfuffle", "sharbert", "fornax"}

	for i, word := range words {

		for _, badWord := range badWords {
			lower := strings.ToLower(word)
			if lower == badWord {
				words[i] = "****"
			}
		}

	}
	newSentence := strings.Join(words, " ")
	return newSentence
}

func (apiCfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding chirp: %s  \n", err)
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user", nil)
	}

	user, err := apiCfg.database.CreateUser(context.Background(), database.CreateUserParams{Email: params.Email, Password: hash})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot insert user", nil)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (apiCfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	userId, err := auth.ValidateJWT(token, apiCfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	chirpParams := chirpParams{}
	if !isValidChirp(&chirpParams, w, r) {
		return
	}

	user, err := apiCfg.database.GetUserById(context.Background(), userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "User not found", err)
		return
	}

	chirp, err := apiCfg.database.CreateChirp(context.Background(), database.CreateChirpParams{Body: replaceBadWords(chirpParams.Body), UserID: user.ID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func (apiCfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := apiCfg.database.GetAllChirps(context.Background())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get chirps", err)
		return
	}
	respondWithJSON(w, http.StatusOK, chirpsResponse(chirps))
}

func chirpsResponse(chirps []database.Chirp) []Chirp {
	chirpsResponse := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		chirpsResponse[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}
	}
	return chirpsResponse
}

func (apiCfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not valid uuid for chirp", err)
		return
	}

	chirp, err := apiCfg.database.GetChirpById(context.Background(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "The chirp was not found", err)
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func (apiCfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding chirp: %s  \n", err)
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return
	}
	user, err := apiCfg.database.GetUserByEmail(context.Background(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username/password", nil)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username/password", nil)
		return
	}
	if !match {
		respondWithError(w, http.StatusUnauthorized, "Invalid username/password", nil)
		return
	}

	signedString, err := auth.MakeJWT(user.ID, apiCfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username/password", nil)
		return
	}
	newRefreshToken, _ := auth.MakeRefreshToken()

	refreshToken, err := apiCfg.database.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{Token: newRefreshToken, UserID: user.ID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not process the request", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        signedString,
		RefreshToken: refreshToken.Token,
	})

}

func (apiCfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	refreshToken, err := apiCfg.database.GetUserFromRefreshToken(context.Background(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	if !refreshToken.RevokedAt.Time.IsZero() {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	type response struct {
		Token string `json:"token"`
	}
	accessToken, err := auth.MakeJWT(refreshToken.UserID, apiCfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (apiCfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
	}
	refreshToken, err := apiCfg.database.GetUserFromRefreshToken(context.Background(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not process the request", err)
		return
	}
	if err := apiCfg.database.RevokeRefreshToken(context.Background(), refreshToken.Token); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not process the request", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (apiCfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	userId, err := auth.ValidateJWT(token, apiCfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding chirp: %s  \n", err)
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update password/username", err)
		return
	}

	user, err := apiCfg.database.UpdateUser(context.Background(), database.UpdateUserParams{Email: params.Email, Password: hashedPassword, ID: userId})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update password/username", err)
		return
	}
	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})

}

func (apiCfg *apiConfig) handlerDeleteChirps(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	userId, err := auth.ValidateJWT(token, apiCfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not valid uuid for chirp", err)
		return
	}
	chirp, err := apiCfg.database.GetChirpById(context.Background(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not find the chirp", err)
		return

	}
	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "You cannot delete the chirp!", err)
		return
	}

	if err := apiCfg.database.DeleteChirp(context.Background(), database.DeleteChirpParams{UserID: userId, ID: chirpId}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete the chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
