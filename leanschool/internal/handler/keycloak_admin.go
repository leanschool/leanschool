package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)


// KeycloakAdminClient assigns roles to users via the Keycloak Admin REST API.
// It authenticates as a service account using client credentials.
type KeycloakAdminClient struct {
	keycloakURL  string
	realm        string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewKeycloakAdminClient creates a new client. clientSecret may be empty; in
// that case role assignment calls will be skipped with a warning.
func NewKeycloakAdminClient(keycloakURL, realm, clientID, clientSecret string) *KeycloakAdminClient {
	return &KeycloakAdminClient{
		keycloakURL:  strings.TrimRight(keycloakURL, "/"),
		realm:        realm,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
	}
}

// adminToken fetches a short-lived admin token via client_credentials.
func (k *KeycloakAdminClient) adminToken(ctx context.Context) (string, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.keycloakURL, k.realm)

	body := url.Values{}
	body.Set("grant_type", "client_credentials")
	body.Set("client_id", k.clientID)
	body.Set("client_secret", k.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching admin token: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("admin token response %d: %s", resp.StatusCode, raw)
	}

	var data struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}
	return data.AccessToken, nil
}

type kcRoleRepresentation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// getRoleRepresentation fetches a realm role by name and returns its ID+name.
func (k *KeycloakAdminClient) getRoleRepresentation(ctx context.Context, token, roleName string) (*kcRoleRepresentation, error) {
	roleURL := fmt.Sprintf("%s/admin/realms/%s/roles/%s", k.keycloakURL, k.realm, url.PathEscape(roleName))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, roleURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building role request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching role %q: %w", roleName, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get role %q response %d: %s", roleName, resp.StatusCode, raw)
	}

	var role kcRoleRepresentation
	if err := json.Unmarshal(raw, &role); err != nil {
		return nil, fmt.Errorf("parsing role response: %w", err)
	}
	return &role, nil
}

// AssignRealmRoles assigns the given business roles (and their service
// dependencies) to the Keycloak user identified by userSub (which equals the
// Keycloak user ID in this realm).
func (k *KeycloakAdminClient) AssignRealmRoles(ctx context.Context, userSub string, roles []string) error {
	if k.clientSecret == "" {
		log.Printf("[keycloak-admin] KEYCLOAK_CLIENT_SECRET not set — skipping role assignment for sub=%s", userSub)
		return nil
	}

	token, err := k.adminToken(ctx)
	if err != nil {
		return fmt.Errorf("obtaining admin token: %w", err)
	}

	// Resolve role names to Keycloak role representations.
	// Composite roles are expanded automatically by Keycloak on assignment.
	var roleReps []kcRoleRepresentation
	for _, roleName := range roles {
		rep, err := k.getRoleRepresentation(ctx, token, roleName)
		if err != nil {
			return fmt.Errorf("resolving role %q: %w", roleName, err)
		}
		roleReps = append(roleReps, *rep)
	}

	assignURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm",
		k.keycloakURL, k.realm, url.PathEscape(userSub))

	bodyBytes, err := json.Marshal(roleReps)
	if err != nil {
		return fmt.Errorf("marshalling roles: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, assignURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("building assign request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("assigning roles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("assign roles response %d: %s", resp.StatusCode, raw)
	}

	log.Printf("[keycloak-admin] assigned roles %v to user sub=%s", roles, userSub)
	return nil
}

// KCUser holds the minimal user representation returned by the Keycloak Admin API.
type KCUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// ListUsersWithRole returns all users who have the given realm role assigned.
func (k *KeycloakAdminClient) ListUsersWithRole(ctx context.Context, roleName string) ([]KCUser, error) {
	token, err := k.adminToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("obtaining admin token: %w", err)
	}

	usersURL := fmt.Sprintf("%s/admin/realms/%s/roles/%s/users", k.keycloakURL, k.realm, url.PathEscape(roleName))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building users request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching users with role: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list users with role response %d: %s", resp.StatusCode, raw)
	}

	var users []KCUser
	if err := json.Unmarshal(raw, &users); err != nil {
		return nil, fmt.Errorf("parsing users response: %w", err)
	}
	return users, nil
}

// RoleOption is a selectable real-world role exposed to the registration form.
type RoleOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListCompositeRoles returns all composite realm roles — these are the
// selectable real-world roles shown during registration.
func (k *KeycloakAdminClient) ListCompositeRoles(ctx context.Context) ([]RoleOption, error) {
	token, err := k.adminToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("obtaining admin token: %w", err)
	}

	rolesURL := fmt.Sprintf("%s/admin/realms/%s/roles?briefRepresentation=false", k.keycloakURL, k.realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rolesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building roles request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching roles: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list roles response %d: %s", resp.StatusCode, raw)
	}

	var all []struct {
		Name        string              `json:"name"`
		Description string              `json:"description"`
		Attributes  map[string][]string `json:"attributes"`
	}
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, fmt.Errorf("parsing roles response: %w", err)
	}

	var options []RoleOption
	for _, r := range all {
		vals := r.Attributes["registration_selectable"]
		if len(vals) > 0 && vals[0] == "true" {
			options = append(options, RoleOption{Name: r.Name, Description: r.Description})
		}
	}
	return options, nil
}
