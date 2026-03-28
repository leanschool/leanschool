package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Joel-Haeberli/file-service/internal/handler"
)

func main() {
	dataDir := getenv("DATA_DIR", "/data")
	h := handler.New(dataDir)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /files", h.Upload)
	mux.HandleFunc("GET /files/{id}", h.Download)
	mux.HandleFunc("DELETE /files/{id}", h.Delete)

	addr := ":8083"
	log.Printf("[file-service] listening on %s (data=%s)", addr, dataDir)
	auth := handler.NewAuthMiddleware("file_service_read", "file_service_write")
	log.Fatal(http.ListenAndServe(addr, corsMiddleware(loggingMiddleware(auth(mux)))))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
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
		log.Printf("[file-service] %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
