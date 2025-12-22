package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	fileSystemDir := http.Dir("./")
	fileServerHandler := http.FileServer(fileSystemDir)

	mux.Handle("/app/", http.StripPrefix("/app/", fileServerHandler))
	mux.HandleFunc("/healthz", func(res http.ResponseWriter, req *http.Request) {
		req.Header.Add("charset", "utf-8")
		req.Header.Add("Content-Type", "text/plain")
		res.WriteHeader(200)
		res.Write([]byte("OK"))
	})

	port := "8080"

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Listening to port: %s", port)

	log.Fatal(server.ListenAndServe())

}
