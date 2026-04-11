package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const sessionName = "mrp-session"

var (
	oauthConfig   *oauth2.Config
	sessionStore  *sessions.CookieStore
	allowedEmails map[string]bool
)

type contextKey string

const (
	ctxUserEmail contextKey = "user_email"
	ctxUserID    contextKey = "user_id"
)

func Initialize() {
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	if oauthConfig.RedirectURL == "" {
		oauthConfig.RedirectURL = "http://localhost:8080/auth/google/callback"
	}

	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Fatal("SESSION_SECRET is required")
	}
	sessionStore = newCookieStore(secret)

	allowedEmails = make(map[string]bool)
	emails := os.Getenv("ALLOWED_EMAILS")
	for _, email := range strings.Split(emails, ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			allowedEmails[email] = true
		}
	}
}

func newCookieStore(secret string) *sessions.CookieStore {
	hash := sha256.Sum256([]byte(secret))
	store := sessions.NewCookieStore(hash[:])
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") != "development",
		SameSite: http.SameSiteLaxMode,
	}
	return store
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	session, _ := sessionStore.Get(r, sessionName)
	session.Values["oauth_state"] = state
	session.Save(r, w)

	url := oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type UserUpsertFunc func(ctx context.Context, email, name string, avatarURL *string) (userID string, err error)

var userUpsertFn UserUpsertFunc

func SetUserUpsertFunc(fn UserUpsertFunc) {
	userUpsertFn = fn
}

func HandleCallbackWithUpsert(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, sessionName)
	expectedState, _ := session.Values["oauth_state"].(string)
	if r.URL.Query().Get("state") != expectedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(session.Values, "oauth_state")

	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("OAuth exchange error: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	if !allowedEmails[userInfo.Email] {
		http.Error(w, "Access denied — this tool is for Subwave team members only", http.StatusForbidden)
		return
	}

	var avatarURL *string
	if userInfo.Picture != "" {
		avatarURL = &userInfo.Picture
	}

	userID := ""
	if userUpsertFn != nil {
		userID, err = userUpsertFn(r.Context(), userInfo.Email, userInfo.Name, avatarURL)
		if err != nil {
			log.Printf("Failed to upsert user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	session.Values["user_email"] = userInfo.Email
	session.Values["user_id"] = userID
	session.Save(r, w)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	http.Redirect(w, r, frontendURL+"/dashboard", http.StatusTemporaryRedirect)
}

// HandleDevLogin bypasses OAuth for local development. Only registered when ENV=development.
func HandleDevLogin(w http.ResponseWriter, r *http.Request) {
	email := "dev@subwave.music"
	name := "Dev User"

	userID := ""
	if userUpsertFn != nil {
		var err error
		userID, err = userUpsertFn(r.Context(), email, name, nil)
		if err != nil {
			log.Printf("Failed to upsert dev user: %v", err)
			http.Error(w, "Failed to create dev user", http.StatusInternalServerError)
			return
		}
	}

	session, _ := sessionStore.Get(r, sessionName)
	session.Values["user_email"] = email
	session.Values["user_id"] = userID
	session.Save(r, w)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	http.Redirect(w, r, frontendURL+"/dashboard", http.StatusTemporaryRedirect)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, sessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, sessionName)
		email, ok := session.Values["user_email"].(string)
		if !ok || email == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
			return
		}

		userID, _ := session.Values["user_id"].(string)
		if userID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserEmail, email)
		ctx = context.WithValue(ctx, ctxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserEmail(r *http.Request) string {
	email, _ := r.Context().Value(ctxUserEmail).(string)
	return email
}

func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(ctxUserID).(string)
	return id
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, sessionName)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"email":   session.Values["user_email"],
		"user_id": session.Values["user_id"],
	})
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
