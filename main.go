package main

import (
	// "fmt"
	"net/http"
)


func main(){

	router := new(http.ServeMux)

	// Serve files from the current directory
	fileServer := http.FileServer(http.Dir("."))
	
	// Handle requests to the root path
	router.Handle("/app/", http.StripPrefix("/app", fileServer))
	

	// Readiness endpoint
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Handle requests to the assets path
	// router.Handle("/app/assets", fileServer)

	server := http.Server{
		Handler: router,
		Addr:    ":8000",
	}

	server.ListenAndServe()

}