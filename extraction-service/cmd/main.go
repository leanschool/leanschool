package main

import (
	"log"
	"net/http"

	"github.com/Joel-Haeberli/extraction-service/internal/handler"
	"github.com/Joel-Haeberli/extraction-service/internal/storage"
)

func main() {
	// Initialize storage
	store := storage.NewMemoryStorage()
	h := handler.New(store)

	// Set up router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /extract", h.Extract)
	mux.HandleFunc("POST /templates", h.CreateTemplate)
	mux.HandleFunc("PUT /templates", h.UpdateTemplate)
	mux.HandleFunc("DELETE /templates/{id}", h.DeleteTemplate)
	mux.HandleFunc("GET /templates", h.GetAllTemplates)
	mux.HandleFunc("GET /templates/{id}", h.GetTemplate)

	outerMux := http.NewServeMux()
	outerMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	auth := handler.NewAuthMiddleware("extraction_read", "extraction_write")
	outerMux.Handle("/", auth(mux))

	addr := ":8084"
	log.Printf("[extraction-service] listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, corsMiddleware(loggingMiddleware(outerMux))))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[extraction-service] %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
