package handler

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ── Context key for JWT claims ─────────────────────────────────────────────────

type contextKey string

const claimsKey contextKey = "claims"

// ClaimsFromContext retrieves the validated JWT claims stored in ctx by the auth
// middleware. Returns nil if no claims are present.
func ClaimsFromContext(ctx context.Context) *jwtClaims {
	c, _ := ctx.Value(claimsKey).(*jwtClaims)
	return c
}

// hasRole returns true if claims contain the given realm role.
func hasRole(claims *jwtClaims, role string) bool {
	if claims == nil {
		return false
	}
	for _, r := range claims.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// ── JWKS types ────────────────────────────────────────────────────────────────

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwkResponse struct {
	Keys []jwkKey `json:"keys"`
}

// ── JWT types ─────────────────────────────────────────────────────────────────

type jwtHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

type realmAccess struct {
	Roles []string `json:"roles"`
}

type jwtClaims struct {
	Exp                int64       `json:"exp"`
	Iss                string      `json:"iss"`
	Sub                string      `json:"sub"`
	Name               string      `json:"name"`
	PreferredUsername   string      `json:"preferred_username"`
	RealmAccess        realmAccess `json:"realm_access"`
}

// ── key cache ─────────────────────────────────────────────────────────────────

type keyCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	jwksURL string
}

func newKeyCache(jwksURL string) *keyCache {
	return &keyCache{
		keys:    make(map[string]*rsa.PublicKey),
		jwksURL: jwksURL,
	}
}

func (c *keyCache) get(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	c.mu.RUnlock()
	if ok {
		return key, nil
	}
	if err := c.refresh(); err != nil {
		return nil, err
	}
	c.mu.RLock()
	key, ok = c.keys[kid]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}
	return key, nil
}

func (c *keyCache) refresh() error {
	resp, err := http.Get(c.jwksURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("fetching JWKS: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var set jwkResponse
	if err := json.Unmarshal(body, &set); err != nil {
		return fmt.Errorf("parsing JWKS: %w", err)
	}

	m := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := rsaFromJWK(k)
		if err != nil {
			log.Printf("[auth] skipping JWKS key %s: %v", k.Kid, err)
			continue
		}
		m[k.Kid] = pub
	}

	c.mu.Lock()
	c.keys = m
	c.mu.Unlock()
	return nil
}

func rsaFromJWK(k jwkKey) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}
	e := 0
	for _, b := range eb {
		e = e<<8 | int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}, nil
}

// ── JWT verification ──────────────────────────────────────────────────────────

func verifyJWT(tokenStr string, cache *keyCache, issuer string) (*jwtClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed token")
	}

	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	var header jwtHeader
	if err := json.Unmarshal(hb, &header); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported alg %q", header.Alg)
	}

	pub, err := cache.get(header.Kid)
	if err != nil {
		return nil, err
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}
	h := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	pb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	var claims jwtClaims
	if err := json.Unmarshal(pb, &claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}
	if claims.Iss != issuer {
		return nil, fmt.Errorf("issuer mismatch: got %q, want %q", claims.Iss, issuer)
	}

	return &claims, nil
}

// ── Middleware ────────────────────────────────────────────────────────────────

// NewAuthMiddleware returns an HTTP middleware that validates Bearer JWT tokens
// issued by Keycloak and checks realm roles.
//
// GET/HEAD requests require readRole; all other methods require writeRole.
// OPTIONS (CORS preflight) always passes through.
// Routes handled by per-handler role checks bypass the global role check —
// the handler itself is responsible for calling hasRole and returning 403.
// Invalid/missing tokens return 401 Unauthorized.
//
// Env vars: KEYCLOAK_URL, KEYCLOAK_REALM, KEYCLOAK_CLIENT_ID,
// KEYCLOAK_CLIENT_SECRET, KEYCLOAK_ISSUER (optional, derived from URL+realm by default).
func NewAuthMiddleware(readRole, writeRole string) func(http.Handler) http.Handler {
	keycloakURL := authEnv("KEYCLOAK_URL", "http://localhost:8180")
	realm := authEnv("KEYCLOAK_REALM", "leanschool")
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, realm)
	issuer := authEnv("KEYCLOAK_ISSUER", fmt.Sprintf("%s/realms/%s", keycloakURL, realm))

	cache := newKeyCache(jwksURL)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// CORS preflight passes through unchanged.
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			raw := r.Header.Get("Authorization")
			if !strings.HasPrefix(raw, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := verifyJWT(strings.TrimPrefix(raw, "Bearer "), cache, issuer)
			if err != nil {
				log.Printf("[auth] token invalid: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store claims in context for all authenticated routes.
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			r = r.WithContext(ctx)

			// All routes in this service perform their own per-resource role checks.
			if isPerHandlerRoleRoute(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			required := writeRole
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				required = readRole
			}

			for _, role := range claims.RealmAccess.Roles {
				if role == required {
					next.ServeHTTP(w, r)
					return
				}
			}

			log.Printf("[auth] role %q not present for sub=%s on %s %s", required, claims.Sub, r.Method, r.URL.Path)
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

// isPerHandlerRoleRoute returns true for paths whose handlers manage their
// own role checks, bypassing the global timetable_read/write gate.
func isPerHandlerRoleRoute(path string) bool {
	prefixes := []string{
		"/plans",
	}
	for _, prefix := range prefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func authEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
