package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)

	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	var hits int32
	hits = cfg.fileserverHits.Load()
	res := fmt.Sprintf(`<html>
	  <body>
	    <h1>Welcome, Chirpy Admin</h1>
	    <p>Chirpy has been visited %d times!</p>
	  </body>
	</html>`, hits)
	w.Write([]byte(res))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func main() {
	var apiCfg apiConfig
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))

	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	s.ListenAndServe()
}
