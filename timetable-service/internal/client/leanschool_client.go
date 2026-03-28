package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ── API response data types ──────────────────────────────────────────────────

// TeacherData represents a teacher as returned by the leanschool API.
type TeacherData struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Prename string `json:"prename"`
	Sub     string `json:"sub,omitempty"`
	Version int    `json:"version"`
}

// SubjectData represents a subject as returned by the leanschool API.
type SubjectData struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Teachers []TeacherData `json:"teachers,omitempty"`
	Version  int           `json:"version"`
}

// SchoolClassData represents a school class as returned by the leanschool API.
type SchoolClassData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Shortcut string `json:"shortcut,omitempty"`
	Version  int    `json:"version"`
}

// RoomData represents a room as returned by the leanschool API.
type RoomData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RoomType string `json:"roomType,omitempty"`
	Version  int    `json:"version"`
}

// EntityRef is a reference to an entity by ID, used in write requests.
type EntityRef struct {
	ID string `json:"id"`
}

// LessonData represents a lesson for the leanschool API.
// Used when creating lessons during finalization.
type LessonData struct {
	ID          string     `json:"id,omitempty"`
	Teacher     *EntityRef `json:"teacher,omitempty"`
	SchoolClass *EntityRef `json:"schoolClass,omitempty"`
	Subject     *EntityRef `json:"subject,omitempty"`
	Room        *EntityRef `json:"room,omitempty"`
	DayOfWeek   *int       `json:"dayOfWeek,omitempty"`
	Period      *int       `json:"period,omitempty"`
	StartTime   string     `json:"startTime,omitempty"`
	EndTime     string     `json:"endTime,omitempty"`
	Version     int        `json:"version"`
}

// ── Token response ───────────────────────────────────────────────────────────

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// ── Client ───────────────────────────────────────────────────────────────────

// LeanschoolClient communicates with the leanschool backend API to fetch
// teachers, subjects, school classes, and rooms for snapshot creation.
// It authenticates via Keycloak client credentials grant (OAuth2).
type LeanschoolClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	keycloakURL  string
	realm        string
	httpClient   *http.Client

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time
}

// NewLeanschoolClient creates a new client for the leanschool API.
func NewLeanschoolClient(baseURL, keycloakURL, realm, clientID, clientSecret string) *LeanschoolClient {
	return &LeanschoolClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
		keycloakURL:  strings.TrimRight(keycloakURL, "/"),
		realm:        realm,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// GetAccessToken obtains an OAuth2 access token from Keycloak using the client
// credentials grant. The token is cached and reused until it expires (with a
// 10-second safety margin).
func (c *LeanschoolClient) GetAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return cached token if still valid (with 10s safety margin).
	if c.cachedToken != "" && time.Now().Before(c.tokenExpiry.Add(-10*time.Second)) {
		return c.cachedToken, nil
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.keycloakURL, c.realm)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)

	resp, err := c.httpClient.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("requesting access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	c.cachedToken = tok.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)

	return c.cachedToken, nil
}

// doGet performs an authenticated GET request to the leanschool API and returns
// the response body.
func (c *LeanschoolClient) doGet(path string) ([]byte, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s returned status %d: %s", path, resp.StatusCode, string(body))
	}

	return body, nil
}

// FetchTeachers retrieves all teachers from the leanschool API.
func (c *LeanschoolClient) FetchTeachers() ([]TeacherData, error) {
	body, err := c.doGet("/teachers")
	if err != nil {
		return nil, err
	}
	var teachers []TeacherData
	if err := json.Unmarshal(body, &teachers); err != nil {
		return nil, fmt.Errorf("parsing teachers response: %w", err)
	}
	return teachers, nil
}

// FetchSubjects retrieves all subjects from the leanschool API.
func (c *LeanschoolClient) FetchSubjects() ([]SubjectData, error) {
	body, err := c.doGet("/subjects")
	if err != nil {
		return nil, err
	}
	var subjects []SubjectData
	if err := json.Unmarshal(body, &subjects); err != nil {
		return nil, fmt.Errorf("parsing subjects response: %w", err)
	}
	return subjects, nil
}

// FetchSchoolClasses retrieves all school classes from the leanschool API.
func (c *LeanschoolClient) FetchSchoolClasses() ([]SchoolClassData, error) {
	body, err := c.doGet("/school-classes")
	if err != nil {
		return nil, err
	}
	var classes []SchoolClassData
	if err := json.Unmarshal(body, &classes); err != nil {
		return nil, fmt.Errorf("parsing school classes response: %w", err)
	}
	return classes, nil
}

// FetchRooms retrieves all rooms from the leanschool API.
func (c *LeanschoolClient) FetchRooms() ([]RoomData, error) {
	body, err := c.doGet("/rooms")
	if err != nil {
		return nil, err
	}
	var rooms []RoomData
	if err := json.Unmarshal(body, &rooms); err != nil {
		return nil, fmt.Errorf("parsing rooms response: %w", err)
	}
	return rooms, nil
}

// CreateLesson creates a new lesson in the leanschool API (used during finalization).
func (c *LeanschoolClient) CreateLesson(lesson LessonData) (*LessonData, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(lesson)
	if err != nil {
		return nil, fmt.Errorf("marshalling lesson: %w", err)
	}

	reqURL := c.baseURL + "/lessons"
	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("building request for POST /lessons: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST /lessons: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from POST /lessons: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("POST /lessons returned status %d: %s", resp.StatusCode, string(body))
	}

	var created LessonData
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("parsing lesson response: %w", err)
	}

	return &created, nil
}
