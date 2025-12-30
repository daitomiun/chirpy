package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/daitonium/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
	platform       string
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

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type successRes struct {
		CleanedBody string `json:"cleaned_body"`
	}

	type parameters struct {
		Data string `json:"body"`
	}
	chirp := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirp); err != nil {
		log.Printf("Error decoding chirp: %s  \n", err)
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return
	}
	if len(chirp.Data) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirpy is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, successRes{CleanedBody: replaceBadWords(chirp.Data)})
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
		Email string `json:"email"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding chirp: %s  \n", err)
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return
	}

	user, err := apiCfg.database.CreateUser(context.Background(), params.Email)
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
