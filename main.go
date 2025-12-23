package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	fileSystemDir := http.Dir("./")
	fileServerHandler := http.FileServer(fileSystemDir)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServerHandler)))
	mux.HandleFunc("GET /api/healthz", handlerCheckHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	port := "8080"

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Listening to port: %s", port)

	log.Fatal(server.ListenAndServe())

}
