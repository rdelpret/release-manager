package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth_NoSession(t *testing.T) {
	initTestStore()

	handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/campaigns", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_WithValidSession(t *testing.T) {
	initTestStore()

	handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := GetUserEmail(r)
		if email != "test@subwave.music" {
			t.Errorf("expected email 'test@subwave.music', got '%s'", email)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/campaigns", nil)
	rr := httptest.NewRecorder()

	session, _ := sessionStore.Get(req, sessionName)
	session.Values["user_email"] = "test@subwave.music"
	session.Values["user_id"] = "test-uuid"
	session.Save(req, rr)

	req2 := httptest.NewRequest("GET", "/api/campaigns", nil)
	for _, cookie := range rr.Result().Cookies() {
		req2.AddCookie(cookie)
	}
	rr2 := httptest.NewRecorder()

	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}
}

func initTestStore() {
	sessionStore = newCookieStore("test-secret-key-for-testing-only")
}
