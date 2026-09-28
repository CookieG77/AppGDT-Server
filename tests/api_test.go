// Package tests holds the integration tests of the API: they start the whole
// server (routes, middlewares, services, repositories) on top of a real
// PostgreSQL database, and check the behavior seen by a client.
//
// They only run when GDT_TEST_DATABASE_URL is set, for example:
//
//	GDT_TEST_DATABASE_URL=postgres://gdt:change-me@localhost:5432/gdt?sslmode=disable go test ./tests/...
//
// Each test creates its own accounts (unique emails) and deletes them at the
// end, so the database can be the development one.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/auth"
	"github.com/CookieG77/AppGDT-Server/internal/database"
	"github.com/CookieG77/AppGDT-Server/internal/handler"
	"github.com/CookieG77/AppGDT-Server/internal/middleware"
	"github.com/CookieG77/AppGDT-Server/internal/ratelimit"
	"github.com/CookieG77/AppGDT-Server/internal/repository"
	"github.com/CookieG77/AppGDT-Server/internal/server"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

// Low limit of failed logins per email, to test the blocking quickly.
const maxFailuresPerEmail = 3

var (
	apiURL   string                     // address of the API started by TestMain
	userRepo *repository.UserRepository // used to clean up the test accounts
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("GDT_TEST_DATABASE_URL")
	if dbURL == "" {
		fmt.Println("GDT_TEST_DATABASE_URL not set: integration tests skipped")
		os.Exit(0)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	if err := database.Migrate(dbURL); err != nil {
		fmt.Println("migrations failed:", err)
		os.Exit(1)
	}
	pool, err := database.Connect(context.Background(), dbURL)
	if err != nil {
		fmt.Println("database connection failed:", err)
		os.Exit(1)
	}

	// Same wiring as cmd/server, with a fast bcrypt cost and low limits
	users := repository.NewUserRepository(pool)
	userRepo = users
	spaces := repository.NewSpaceRepository(pool)
	notes := repository.NewNoteRepository(pool)
	hasher := auth.NewPasswordHasher(10)
	tokens := auth.NewTokenManager("integration-tests-secret-32-characters", time.Hour)
	limiters := service.LoginLimiters{
		ByEmail: ratelimit.New(maxFailuresPerEmail, time.Minute),
		ByIP:    ratelimit.New(1000, time.Minute),
	}
	authService, err := service.NewAuthService(users, hasher, tokens, limiters)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	srv := server.New("", server.Handlers{
		Health: handler.NewHealthHandler(pool),
		Auth:   handler.NewAuthHandler(authService),
		User:   handler.NewUserHandler(authService, service.NewAccountService(users, spaces, notes, hasher, limiters)),
		Space:  handler.NewSpaceHandler(service.NewSpaceService(spaces)),
		Note:   handler.NewNoteHandler(service.NewNoteService(notes, spaces)),
	}, middleware.Authenticate(tokens, users))

	ts := httptest.NewServer(srv.Handler)
	apiURL = ts.URL
	code := m.Run()
	ts.Close()
	pool.Close()
	os.Exit(code)
}

// --- Helpers ---------------------------------------------------------------------

type response struct {
	status int
	header http.Header
	body   map[string]any
	list   []map[string]any
	raw    string
}

// call sends a JSON request to the API. body may be nil, a string (sent as
// is) or any value encoded as JSON.
func call(t *testing.T, method, path, token string, body any) response {
	t.Helper()
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		reader = strings.NewReader(b)
	default:
		payload, _ := json.Marshal(b)
		reader = bytes.NewReader(payload)
	}
	req, _ := http.NewRequest(method, apiURL+path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	r := response{status: resp.StatusCode, header: resp.Header, raw: string(raw)}
	if len(raw) > 0 && raw[0] == '[' {
		_ = json.Unmarshal(raw, &r.list)
	} else {
		_ = json.Unmarshal(raw, &r.body)
	}
	return r
}

func expect(t *testing.T, r response, status int, code string) {
	t.Helper()
	if r.status != status {
		t.Fatalf("status = %d, want %d (body %s)", r.status, status, r.raw)
	}
	if code != "" && r.body["code"] != code {
		t.Fatalf("code = %v, want %s (body %s)", r.body["code"], code, r.raw)
	}
}

var emailCounter atomic.Int64

// user is an account created for one test.
type user struct {
	email, password, token string
}

// newUser registers and logs in a new account, deleted at the end of the test.
func newUser(t *testing.T) *user {
	t.Helper()
	u := &user{
		email:    fmt.Sprintf("it-%d-%d@example.com", time.Now().UnixNano(), emailCounter.Add(1)),
		password: "motdepasse-test",
	}
	expect(t, call(t, "POST", "/auth/register", "", map[string]string{
		"email": u.email, "username": "Test", "password": u.password,
	}), http.StatusCreated, "")
	u.token = login(t, u.email, u.password)
	t.Cleanup(func() {
		// Directly in the database: the API could refuse it (blocked login).
		// Nothing to do if the test already deleted the account.
		if existing, err := userRepo.GetByEmail(context.Background(), u.email); err == nil {
			_ = userRepo.Delete(context.Background(), existing.ID)
		}
	})
	return u
}

func login(t *testing.T, email, password string) string {
	t.Helper()
	r := call(t, "POST", "/auth/login", "", map[string]string{"email": email, "password": password})
	expect(t, r, http.StatusOK, "")
	token, _ := r.body["token"].(string)
	return token
}

// id returns the "id" field of a JSON object as a path segment.
func id(r response) string {
	return fmt.Sprintf("%.0f", r.body["id"])
}

// --- Tests -----------------------------------------------------------------------------

func TestHealthAndUnknownRoutes(t *testing.T) {
	expect(t, call(t, "GET", "/health", "", nil), http.StatusOK, "")
	expect(t, call(t, "GET", "/nope", "", nil), http.StatusNotFound, "ROUTE_NOT_FOUND")

	r := call(t, "PATCH", "/spaces", "", nil)
	expect(t, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
	if allow := r.header.Get("Allow"); !strings.Contains(allow, "GET") || !strings.Contains(allow, "POST") {
		t.Errorf("Allow = %q", allow)
	}
}

func TestRegistrationAndLogin(t *testing.T) {
	u := newUser(t)

	// The profile never contains the password
	r := call(t, "GET", "/users/me", u.token, nil)
	expect(t, r, http.StatusOK, "")
	if strings.Contains(strings.ToLower(r.raw), "password") {
		t.Errorf("the profile must not contain the password: %s", r.raw)
	}

	// Same email with another case: refused
	expect(t, call(t, "POST", "/auth/register", "", map[string]string{
		"email": strings.ToUpper(u.email), "username": "Autre", "password": "motdepasse",
	}), http.StatusConflict, "EMAIL_ALREADY_USED")

	// Every invalid field is reported at once
	r = call(t, "POST", "/auth/register", "", map[string]string{"email": "pas-un-email", "username": "", "password": "court"})
	expect(t, r, http.StatusBadRequest, "VALIDATION_ERROR")
	if details, _ := r.body["details"].([]any); len(details) != 3 {
		t.Errorf("expected 3 field errors, got %v", r.body["details"])
	}

	// Unknown field refused (mass assignment)
	expect(t, call(t, "POST", "/auth/register", "", `{"email":"a@b.fr","username":"a","password":"motdepasse","isAdmin":true}`),
		http.StatusBadRequest, "INVALID_JSON")

	// Wrong password and unknown email give the same answer
	expect(t, call(t, "POST", "/auth/login", "", map[string]string{"email": u.email, "password": "mauvais"}),
		http.StatusUnauthorized, "INVALID_CREDENTIALS")
	expect(t, call(t, "POST", "/auth/login", "", map[string]string{"email": "inconnu@example.com", "password": "mauvais"}),
		http.StatusUnauthorized, "INVALID_CREDENTIALS")

	// Protected routes need a valid token
	expect(t, call(t, "GET", "/spaces", "", nil), http.StatusUnauthorized, "UNAUTHORIZED")
	expect(t, call(t, "GET", "/spaces", "falsifie", nil), http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestSpacesAndNotes(t *testing.T) {
	u := newUser(t)

	space := call(t, "POST", "/spaces", u.token, map[string]string{"name": "  Devoirs  ", "description": "École"})
	expect(t, space, http.StatusCreated, "")
	if space.body["name"] != "Devoirs" {
		t.Errorf("the name must be trimmed, got %q", space.body["name"])
	}
	spacePath := "/spaces/" + id(space)

	expect(t, call(t, "POST", "/spaces", u.token, map[string]string{"name": ""}), http.StatusBadRequest, "VALIDATION_ERROR")
	expect(t, call(t, "GET", "/spaces/abc", u.token, nil), http.StatusBadRequest, "INVALID_ID")

	r := call(t, "PUT", spacePath, u.token, map[string]string{"name": "Cours"})
	expect(t, r, http.StatusOK, "")
	if r.body["name"] != "Cours" || r.body["description"] != "" {
		t.Errorf("update must replace the name and reset the omitted description: %s", r.raw)
	}

	// A note without status is "todo"; the space comes from the path
	note := call(t, "POST", spacePath+"/notes", u.token, map[string]string{"title": "Exercices", "content": "p.52"})
	expect(t, note, http.StatusCreated, "")
	if note.body["status"] != "todo" || id(note) == "" {
		t.Errorf("unexpected note %s", note.raw)
	}
	notePath := "/notes/" + id(note)

	expect(t, call(t, "POST", spacePath+"/notes", u.token, map[string]string{"title": "x", "status": "urgent"}),
		http.StatusBadRequest, "VALIDATION_ERROR")
	expect(t, call(t, "POST", spacePath+"/notes", u.token, map[string]any{"title": "x", "spaceId": 1}),
		http.StatusBadRequest, "INVALID_JSON")

	r = call(t, "PUT", notePath, u.token, map[string]string{"title": "Exercices", "content": "- [x] fait", "status": "done"})
	expect(t, r, http.StatusOK, "")
	if r.body["status"] != "done" {
		t.Errorf("status not updated: %s", r.raw)
	}

	r = call(t, "GET", spacePath+"/notes", u.token, nil)
	expect(t, r, http.StatusOK, "")
	if len(r.list) != 1 {
		t.Fatalf("expected 1 note, got %s", r.raw)
	}

	// Deleting the space deletes its notes
	expect(t, call(t, "DELETE", spacePath, u.token, nil), http.StatusNoContent, "")
	expect(t, call(t, "GET", notePath, u.token, nil), http.StatusNotFound, "NOT_FOUND")
	expect(t, call(t, "GET", spacePath+"/notes", u.token, nil), http.StatusNotFound, "NOT_FOUND")
}

func TestUsersCannotReachEachOthersData(t *testing.T) {
	owner, other := newUser(t), newUser(t)

	space := call(t, "POST", "/spaces", owner.token, map[string]string{"name": "Privé"})
	spacePath := "/spaces/" + id(space)
	note := call(t, "POST", spacePath+"/notes", owner.token, map[string]string{"title": "Secret"})
	notePath := "/notes/" + id(note)

	// Every access gives the same 404 as a missing resource
	spaceBody := map[string]string{"name": "x"}
	noteBody := map[string]string{"title": "x"}
	for _, c := range []struct {
		method, path string
		body         any
	}{
		{"GET", spacePath, nil}, {"PUT", spacePath, spaceBody}, {"DELETE", spacePath, nil},
		{"GET", spacePath + "/notes", nil}, {"POST", spacePath + "/notes", noteBody},
		{"GET", notePath, nil}, {"PUT", notePath, noteBody}, {"DELETE", notePath, nil},
	} {
		expect(t, call(t, c.method, c.path, other.token, c.body), http.StatusNotFound, "NOT_FOUND")
	}

	r := call(t, "GET", "/spaces", other.token, nil)
	if len(r.list) != 0 {
		t.Errorf("the other user must not list the space: %s", r.raw)
	}
	expect(t, call(t, "GET", notePath, owner.token, nil), http.StatusOK, "")
}

func TestLoginIsBlockedAfterTooManyFailures(t *testing.T) {
	u := newUser(t)
	for i := 0; i < maxFailuresPerEmail; i++ {
		call(t, "POST", "/auth/login", "", map[string]string{"email": u.email, "password": "mauvais"})
	}

	// Blocked even with the right password, with the waiting time
	r := call(t, "POST", "/auth/login", "", map[string]string{"email": u.email, "password": u.password})
	expect(t, r, http.StatusTooManyRequests, "TOO_MANY_ATTEMPTS")
	if r.header.Get("Retry-After") == "" {
		t.Error("Retry-After header missing")
	}
}

func TestAccountExportAndDeletion(t *testing.T) {
	u := newUser(t)
	space := call(t, "POST", "/spaces", u.token, map[string]string{"name": "Jobs"})
	call(t, "POST", "/spaces/"+id(space)+"/notes", u.token, map[string]string{"title": "CV"})

	r := call(t, "GET", "/users/me/export", u.token, nil)
	expect(t, r, http.StatusOK, "")
	if !strings.HasPrefix(r.header.Get("Content-Disposition"), "attachment") {
		t.Errorf("Content-Disposition = %q", r.header.Get("Content-Disposition"))
	}
	if strings.Contains(strings.ToLower(r.raw), "password") || !strings.Contains(r.raw, `"CV"`) {
		t.Errorf("unexpected export: %s", r.raw)
	}

	expect(t, call(t, "DELETE", "/users/me", u.token, map[string]string{"password": "mauvais"}), http.StatusForbidden, "INVALID_PASSWORD")
	expect(t, call(t, "DELETE", "/users/me", u.token, map[string]string{"password": u.password}), http.StatusNoContent, "")

	// The token of a deleted account is refused, and the login fails
	expect(t, call(t, "GET", "/users/me", u.token, nil), http.StatusUnauthorized, "UNAUTHORIZED")
	expect(t, call(t, "POST", "/auth/login", "", map[string]string{"email": u.email, "password": u.password}),
		http.StatusUnauthorized, "INVALID_CREDENTIALS")
}
