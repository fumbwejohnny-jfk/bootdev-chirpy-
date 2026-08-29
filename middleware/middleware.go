package middleware

import (
	"log"
	"net/http"
	"github.com/fumbwejohnny-jfk/bootdev-chirpy/internal/database"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB database.Queries
}


func NewApiConfig() *apiConfig{
	return &apiConfig{}
}


func MiiddlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) GetMetrics() int32 {
	return cfg.fileserverHits.Load()
}

func (cfg *apiConfig) ResetMetrics(){
	cfg.fileserverHits.Store(0)
}