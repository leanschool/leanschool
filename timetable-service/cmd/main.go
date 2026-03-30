package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Joel-Haeberli/timetable-service/internal/client"
	"github.com/Joel-Haeberli/timetable-service/internal/handler"
	"github.com/Joel-Haeberli/timetable-service/internal/planner"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "admin"),
		getenv("DB_PASSWORD", "secret"),
		getenv("DB_NAME", "timetable"),
	)

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	lsClient := client.NewLeanschoolClient(
		getenv("LEANSCHOOL_URL", "http://localhost:8080"),
		getenv("KEYCLOAK_URL", "http://localhost:8180"),
		getenv("KEYCLOAK_REALM", "leanschool"),
		getenv("KEYCLOAK_CLIENT_ID", "timetable-service-client"),
		getenv("KEYCLOAK_CLIENT_SECRET", ""),
	)

	mux := http.NewServeMux()

	handler.NewPlanHandler(store).RegisterRoutes(mux)
	handler.NewTimeSlotHandler(store).RegisterRoutes(mux)
	handler.NewRequirementHandler(store).RegisterRoutes(mux)
	handler.NewConstraintHandler(store).RegisterRoutes(mux)
	handler.NewEntryHandler(store).RegisterRoutes(mux)
	handler.NewConflictHandler(store).RegisterRoutes(mux)
	p := planner.New(store)
	handler.NewWorkflowHandler(store, lsClient, p).RegisterRoutes(mux)
	handler.NewSnapshotHandler(store).RegisterRoutes(mux)

	outerMux := http.NewServeMux()
	outerMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	auth := handler.NewAuthMiddleware("timetable_read", "timetable_write")
	outerMux.Handle("/", auth(mux))

	addr := ":8085"
	log.Printf("[timetable-service] listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler.CORSMiddleware(handler.LoggingMiddleware(outerMux))))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
