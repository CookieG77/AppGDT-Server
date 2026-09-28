package httpjson

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
)

type input struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func decode(body string) error {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	var dst input
	return DecodeJSON(httptest.NewRecorder(), req, &dst)
}

func TestDecodeJSON(t *testing.T) {
	if err := decode(`{"name":"Devoirs","age":3}`); err != nil {
		t.Fatalf("a valid body must be accepted: %v", err)
	}

	tests := map[string]string{
		"":                          "vide",
		`{"name":`:                  "mal formé",
		`{"name":"a"} {"name":"b"}`: "un seul objet",
		`{"name":"a","admin":true}`: `"admin"`,
		`{"age":"trois"}`:           `"age"`,
		`["a"]`:                     "objet JSON",
		`{"name":"` + strings.Repeat("a", maxBodySize) + `"}`: "octets",
	}
	for body, want := range tests {
		err := decode(body)
		if err == nil || !strings.Contains(err.Error(), want) {
			short := body
			if len(short) > 30 {
				short = short[:30] + "…"
			}
			t.Errorf("body %q: error %v, want a message containing %q", short, err, want)
		}
	}
}

func TestWriteErrorFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	vErr := &domain.ValidationError{}
	vErr.Add("name", "Le nom est obligatoire.")
	WriteValidationError(rec, vErr)

	if rec.Code != http.StatusBadRequest || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status %d, content type %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	var body struct {
		Code    string              `json:"code"`
		Message string              `json:"message"`
		Details []domain.FieldError `json:"details"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != "VALIDATION_ERROR" || len(body.Details) != 1 || body.Details[0].Field != "name" {
		t.Errorf("unexpected body %+v", body)
	}

	// Without details, the field is omitted
	rec = httptest.NewRecorder()
	WriteError(rec, http.StatusNotFound, "NOT_FOUND", "Ressource introuvable.")
	if strings.Contains(rec.Body.String(), "details") {
		t.Errorf("details must be omitted, got %s", rec.Body.String())
	}
}
