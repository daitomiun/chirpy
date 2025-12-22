package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("charset", "utf-8")
	r.Header.Add("Content-Type", "text/plain")
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

func checkHealth(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("charset", "utf-8")
	r.Header.Add("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	fileSystemDir := http.Dir("./")
	fileServerHandler := http.FileServer(fileSystemDir)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServerHandler)))
	mux.HandleFunc("GET /api/healthz", checkHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)

	port := "8080"

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Listening to port: %s", port)

	log.Fatal(server.ListenAndServe())

}
