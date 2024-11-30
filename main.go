package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/chichigami/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(db)

	const port = "8080"
	apiCfg := apiConfig{
		db: dbQueries,
	}
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricInc(http.FileServer(http.FileSystem(http.Dir("."))))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetric)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricReset)
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving on port: %s\n", port)
	server.ListenAndServe()
}

func handlerHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}
