package main

import (
	"net/http"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/fumbwejohnny-jfk/bootdev-chirpy/middleware"
	"github.com/fumbwejohnny-jfk/bootdev-chirpy/internal/database"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"os"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body     string    `json:"body"`
	UserID   uuid.UUID `json:"user_id"`
}


func main(){

	router := new(http.ServeMux)

	// get env variables
	godotenv.Load()

	// get database 
	dbURL := os.Getenv("DB_URL")

	// connection to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error connection: %v", err)
	}
	defer db.Close()

	// get Queries
	dbQueries := database.New(db)

	// get api config
	cfg := middleware.NewApiConfig()

	// store database into apiConfig
	cfg.DB = *dbQueries
	
	// Serve files from the current directory
	fileServer := http.FileServer(http.Dir("."))
	
	// Handle requests to the root path
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

	// Handle reset endpoint
	router.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		cfg.ResetMetrics()

		// get database 
		platform := os.Getenv("PLATFORM")
		if platform != "dev" {
			w.WriteHeader(403)
			return
		}

		// delete users
		err := dbQueries.DeleteUsers(r.Context())
		if err != nil {
			fmt.Printf("Error deleting users: %v", err)
		}
		// w.WriteHeader(http.StatusOK)
		//  counter := fmt.Sprintf("Hits: %v", cfg.GetMetrics())
		 counter := fmt.Sprintf("All users have been deleted and the current user has been cleared.")
		w.Write([]byte(counter))
		// fmt.Println("All users have been deleted and the current user has been cleared.")
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

	// users endpoint
	router.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
	
		// Create a new user in the database
		newUser := database.User {}

		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		createdUser, err := dbQueries.CreateUser(r.Context(), newUser.Email)
		if err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		user := User{
			ID:        createdUser.ID,
			CreatedAt: createdUser.CreatedAt,
			UpdatedAt: createdUser.UpdatedAt,
			Email:     createdUser.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// chirps endpoint
	router.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {

		type parameters struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
		}

		var params parameters

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&params)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Something went wrong",
			})
			return
		}

		if len([]rune(params.Body)) > 140 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Chirp is too long",
			})
			return
		}

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

		// Create database parameters using the cleaned body.
		newChirp := database.CreateChirpParams{
			Body:   cleaned,
			UserID: params.UserID,
		}
		
		// Create chirp in database.
		createdChirp, err := dbQueries.CreateChirp(r.Context(), newChirp)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "failed to create chirp", http.StatusInternalServerError)
			return
		}

		chirp := Chirp{
			ID        : createdChirp.ID,
			CreatedAt : createdChirp.CreatedAt,
			UpdatedAt : createdChirp.UpdatedAt,
			Body      : createdChirp.Body,
			UserID    : createdChirp.UserID,
		}

		// Return the created chirp as JSON.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(chirp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	
	// chirps endpoint
	router.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {

		chirps, err := dbQueries.GetChirps(r.Context())
		if err != nil {
			http.Error(w, "failed to get chirps", http.StatusInternalServerError)
			return
		}

		// Convert database chirps to API chirps
		response := make([]Chirp, 0, len(chirps))

		for _, dbChirp := range chirps {
			chirp := Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.UpdatedAt,
				Body:      dbChirp.Body,
				UserID:    dbChirp.UserID,
			}

			response = append(response, chirp)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// chirp endpoint: singleton
	router.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		// id is a string, so convert it to UUID
		chirpID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "invalid chirp ID", http.StatusBadRequest)
			return
		}

		chirp, err := dbQueries.GetChirp(r.Context(), chirpID)
		if err != nil {
			http.Error(w, "chirp not found", http.StatusNotFound)
			return
		}

		response := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
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