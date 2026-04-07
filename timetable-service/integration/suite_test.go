//go:build integration

package integration_test

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Joel-Haeberli/timetable-service/internal/client"
	"github.com/Joel-Haeberli/timetable-service/internal/handler"
	"github.com/Joel-Haeberli/timetable-service/internal/planner"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
	testcontainers "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ── Shared test state ─────────────────────────────────────────────────────────

var (
	srv        *httptest.Server
	rsaKey     *rsa.PrivateKey
	issuer     string
	readToken  string
	writeToken string
	resolveToken string

	lsMu   sync.Mutex
	lsMock lsMockData
)

// lsMockData holds configurable data served by the leanschool stub server.
type lsMockData struct {
	Teachers       []client.TeacherData
	Subjects       []client.SubjectData
	Classes        []client.SchoolClassData
	Rooms          []client.RoomData
	LessonError    bool
	LessonsCreated []client.LessonData
}

func setMock(m lsMockData) {
	lsMu.Lock()
	m.LessonsCreated = nil
	lsMock = m
	lsMu.Unlock()
}

func getLessonsCreated() []client.LessonData {
	lsMu.Lock()
	defer lsMu.Unlock()
	out := make([]client.LessonData, len(lsMock.LessonsCreated))
	copy(out, lsMock.LessonsCreated)
	return out
}

// TestMain bootstraps the entire integration test suite:
//  1. Starts a Postgres testcontainer.
//  2. Generates an RSA key pair and starts an in-process JWKS+token server.
//  3. Starts an in-process leanschool API stub server.
//  4. Assembles the timetable service HTTP handler stack.
//  5. Pre-generates test JWT tokens.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// ── 1. Postgres testcontainer ─────────────────────────────────────────
	pgCtr, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("timetable"),
		tcpostgres.WithUsername("admin"),
		tcpostgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer pgCtr.Terminate(ctx) //nolint:errcheck

	dsn, err := pgCtr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres connection string: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to database: %v\n", err)
		os.Exit(1)
	}

	// ── 2. RSA key + JWKS/token server ───────────────────────────────────
	rsaKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate rsa key: %v\n", err)
		os.Exit(1)
	}

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/certs"):
			// JWKS endpoint
			pub := &rsaKey.PublicKey
			n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
			e := encodeE(pub.E)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"keys": []map[string]any{{
					"kid": "test-key-1",
					"kty": "RSA",
					"alg": "RS256",
					"use": "sig",
					"n":   n,
					"e":   e,
				}},
			})

		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/token"):
			// OAuth2 token endpoint (for LeanschoolClient)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "dummy-service-token",
				"expires_in":   3600,
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer authSrv.Close()

	issuer = authSrv.URL + "/realms/leanschool"

	// ── 3. Leanschool stub server ─────────────────────────────────────────
	lsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lsMu.Lock()
		mock := lsMock
		lsMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/teachers":
			json.NewEncoder(w).Encode(mock.Teachers)
		case r.Method == http.MethodGet && r.URL.Path == "/subjects":
			json.NewEncoder(w).Encode(mock.Subjects)
		case r.Method == http.MethodGet && r.URL.Path == "/school-classes":
			json.NewEncoder(w).Encode(mock.Classes)
		case r.Method == http.MethodGet && r.URL.Path == "/rooms":
			json.NewEncoder(w).Encode(mock.Rooms)
		case r.Method == http.MethodPost && r.URL.Path == "/lessons":
			if mock.LessonError {
				http.Error(w, "leanschool error", http.StatusInternalServerError)
				return
			}
			var lesson client.LessonData
			json.NewDecoder(r.Body).Decode(&lesson) //nolint:errcheck
			lsMu.Lock()
			lsMock.LessonsCreated = append(lsMock.LessonsCreated, lesson)
			lsMu.Unlock()
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(lesson)
		default:
			http.NotFound(w, r)
		}
	}))
	defer lsSrv.Close()

	// ── 4. Set env vars before constructing auth middleware ───────────────
	os.Setenv("KEYCLOAK_URL", authSrv.URL)
	os.Setenv("KEYCLOAK_REALM", "leanschool")
	os.Setenv("KEYCLOAK_ISSUER", issuer)

	// ── 5. Assemble service ───────────────────────────────────────────────
	lsClient := client.NewLeanschoolClient(
		lsSrv.URL,
		authSrv.URL,
		"leanschool",
		"test-client",
		"test-secret",
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
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	})
	auth := handler.NewAuthMiddleware("timetable_read", "timetable_write")
	outerMux.Handle("/", auth(mux))

	srv = httptest.NewServer(handler.CORSMiddleware(handler.LoggingMiddleware(outerMux)))
	defer srv.Close()

	// ── 6. Pre-generate tokens ─────────────────────────────────────────────
	readToken = tokenFor("timetable_read")
	writeToken = tokenFor("timetable_read", "timetable_write")
	resolveToken = tokenFor("timetable_read", "timetable_write", "timetable_resolve")

	os.Exit(m.Run())
}

// ── JWT helpers ───────────────────────────────────────────────────────────────

// tokenFor produces a hand-crafted RS256 JWT signed with rsaKey.
func tokenFor(roles ...string) string {
	header := base64url(mustJSON(map[string]any{
		"alg": "RS256",
		"kid": "test-key-1",
		"typ": "JWT",
	}))
	payload := base64url(mustJSON(map[string]any{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iss": issuer,
		"sub": "test-user-sub",
		"realm_access": map[string]any{
			"roles": roles,
		},
	}))

	msg := header + "." + payload
	h := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, h[:])
	if err != nil {
		panic(fmt.Sprintf("sign jwt: %v", err))
	}
	return msg + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func base64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal json: %v", err))
	}
	return b
}

// encodeE encodes the RSA public exponent as a base64url big-endian byte slice.
func encodeE(e int) string {
	var buf [4]byte
	n := big.NewInt(int64(e))
	b := n.Bytes()
	copy(buf[4-len(b):], b)
	// Trim leading zeros
	start := 0
	for start < 3 && buf[start] == 0 {
		start++
	}
	return base64.RawURLEncoding.EncodeToString(buf[start:])
}
