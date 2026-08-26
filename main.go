package main

import (
	"net/http"
	"fmt"
	"github.com/fumbwejohnny-jfk/bootdev-chirpy/middleware"
)


func main(){

	router := new(http.ServeMux)

	// get api config
	cfg := middleware.NewApiConfig()
	

	// Serve files from the current directory
	fileServer := http.FileServer(http.Dir("."))
	
	// Handle requests to the root path
	// router.Handle("/app/",middleware.MiiddlewareLog(http.StripPrefix("/app", fileServer)))
	router.Handle("/app/", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	
	// Handle the number of requests
	router.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		text := fmt.Sprintf(`
			<html>	
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>`, cfg.GetMetrics())
		w.Write([]byte(text))
	})	

	// Handle reset
	router.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		cfg.ResetMetrics()
		counter := fmt.Sprintf("Hits: %v", cfg.GetMetrics())
		w.Write([]byte(counter))
	})	

	// Readiness endpoint
	router.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Handle requests to the assets path
	// router.Handle("/app/assets", fileServer)

	server := http.Server{
		Handler: router,
		Addr:    ":8080",
	}

	// server stats
	// mux.Handle("/app/", middleware.MiddlewareLog(handler))
	server.ListenAndServe()

}