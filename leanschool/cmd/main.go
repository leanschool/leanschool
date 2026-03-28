package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Joel-Haeberli/leanschool/internal/api"
	"github.com/Joel-Haeberli/leanschool/internal/handler"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "admin"),
		getenv("DB_PASSWORD", "secret"),
		getenv("DB_NAME", "leanschool"),
	)

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	kc := handler.NewKeycloakAdminClient(
		getenv("KEYCLOAK_URL", "http://localhost:8180"),
		getenv("KEYCLOAK_REALM", "leanschool"),
		getenv("KEYCLOAK_CLIENT_ID", "leanschool-service"),
		getenv("KEYCLOAK_CLIENT_SECRET", ""),
	)

	mux := http.NewServeMux()

	// Non-spec handlers: account, user management (not covered by OpenAPI spec)
	handler.NewAccountHandler(store).RegisterRoutes(mux)
	handler.NewUserHandler(store, kc).RegisterRoutes(mux)
	handler.NewUserRegistryHandler(store).RegisterRoutes(mux)
	handler.NewRegistrationHandler(store).RegisterRoutes(mux)
	handler.NewRoleHandler(store).RegisterRoutes(mux)

	// Spec-driven handlers: all domain CRUD + receipts (generated routing from OpenAPI spec)
	handlers := handler.NewHandlers(store, "")
	specMux := api.HandlerFromMux(api.NewStrictHandler(handlers, nil), mux)

	addr := ":8080"
	log.Printf("leanschool listening on %s", addr)
	auth := handler.NewAuthMiddleware("leanschool_read", "leanschool_write")
	log.Fatal(http.ListenAndServe(addr, handler.CORSMiddleware(handler.LoggingMiddleware(auth(specMux)))))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
