package main

import (
	"net/http"
	"fmt"
	"github.com/fumbwejohnny-jfk/bootdev-chirpy/middleware"
	"encoding/json"
	"regexp"
	"strings"
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
	router.HandleFunc("GET /api/metrics", func(w http.ResponseWriter, r *http.Request){
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
	router.HandleFunc("POST /api/reset", func(w http.ResponseWriter, r *http.Request){
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

	// validate_chirp endpoint
	router.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}

		err := decoder.Decode(&params)

		w.Header().Set("Content-Type", "application/json")

		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Something went wrong",
			})
			return
		}

		if len([]rune(params.Body)) > 140 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Chirp is too long",
			})
			return
		}
		// re := regexp.MustCompile(`(?i)(^)(kerfuffle|sharbert|fornax)($)`)
		// params.Body = censoringWords(params.Body, []string{"kerfuffle", "sharbert", "fornax"})
		// Assuming length validation has already passed.
		var profaneWords = map[string]struct{}{
			"kerfuffle": {},
			"sharbert":  {},
			"fornax":    {},
		}
		words := strings.Fields(params.Body)

		for i, word := range words {
			if _, ok := profaneWords[strings.ToLower(word)]; ok {
				words[i] = "****"
			}
		}

		cleaned := strings.Join(words, " ")
		
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{
			"cleaned_body": cleaned,
		})
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

func censorWords(s string, words []string) string {
	for _, word := range words {
		re := regexp.MustCompile(`(?i)(^|[\s])` + regexp.QuoteMeta(word) + `($|[\s])`)
		s = re.ReplaceAllStringFunc(s, func(match string) string {
			return strings.Repeat("*", 4) + match[len(word):]
		})
	}

	return s
}

func censoringWords(s string, words []string) string{
	for _, word := range words {
		
		if strings.Contains(strings.ToLower(s), strings.ToLower(word)){
			
			 	if strings.Contains(s, strings.ToUpper(word[:1]) + word[1:]){
					s = strings.ReplaceAll(s, strings.ToUpper(word[:1]) + word[1:], "****")
					fmt.Println(s, strings.ToUpper(word[:1]) + word[1:])
				}
				if strings.Contains(s, strings.ToUpper(word)){
					s = strings.ReplaceAll(s, strings.ToUpper(word), "****")
					fmt.Println(s, strings.ToUpper(word) )
				}
				if strings.Contains(s, strings.ToLower(word)){
					s = strings.ReplaceAll(s, strings.ToLower(word), "****")
					fmt.Println(s, strings.ToLower(word) )
				}
			 
		}
	}
	fmt.Println(s)
	return s
}

// DECODE Json
// func handler(w http.ResponseWriter, r *http.Request){
//     type parameters struct {
//         Name string `json:"name"`
//         Age int `json:"age"`
//     }

//     decoder := json.NewDecoder(r.Body)
//     params := parameters{}
//     err := decoder.Decode(&params)
//     if err != nil {
// 		log.Printf("Error decoding parameters: %s", err)
// 		w.WriteHeader(500)
// 		return
//     }
    // params is a struct with data populated successfully
    // ...
// }

// ENCODE Json
// func handler(w http.ResponseWriter, r *http.Request){
//     // ...

//     type returnVals struct {
//         CreatedAt time.Time `json:"created_at"`
//         ID int `json:"id"`
//     }
//     respBody := returnVals{
//         CreatedAt: time.Now(),
//         ID: 123,
//     }
//     dat, err := json.Marshal(respBody)
// 	if err != nil {
// 			log.Printf("Error marshalling JSON: %s", err)
// 			w.WriteHeader(500)
// 			return
// 	}
//     w.Header().Set("Content-Type", "application/json")
//     w.WriteHeader(200)
//     w.Write(dat)
// }