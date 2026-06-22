package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// preRegisteredUser, mock sunucuda önceden kayıtlı kullanıcıyı temsil eder.
type preRegisteredUser struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AccessToken string `json:"access_token"`
	SlugID      string `json:"slug_id,omitempty"`
}

// SupabaseMock, Supabase Auth API'sini mock'lar.
type SupabaseMock struct {
	*httptest.Server
	registeredUsers map[string]*preRegisteredUser // email → user
	jwtSecret       string
}

// NewSupabaseMock, yeni bir mock Supabase sunucusu oluşturur.
func NewSupabaseMock() *SupabaseMock {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		secret = "test-jwt-secret-key-for-hs256-at-least-32-characters!"
	}
	m := &SupabaseMock{
		registeredUsers: make(map[string]*preRegisteredUser),
		jwtSecret:       secret,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/v1/signup", m.handleSignup)
	mux.HandleFunc("/auth/v1/token", m.handleLogin)
	mux.HandleFunc("/auth/v1/user", m.handleUser)
	m.Server = httptest.NewServer(mux)
	return m
}

// Start, HTTP sunucusunu yeniden başlatır (opsiyonel).
func (m *SupabaseMock) Start() {
	if m.Server != nil {
		m.Server.Close()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/v1/signup", m.handleSignup)
	mux.HandleFunc("/auth/v1/token", m.handleLogin)
	mux.HandleFunc("/auth/v1/user", m.handleUser)
	m.Server = httptest.NewServer(mux)
}

// generateJWT, userID için HS256 ile imzalanmış geçerli bir JWT üretir.
func (m *SupabaseMock) generateJWT(userID string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"iat":   now.Unix(),
		"exp":   now.Add(1 * time.Hour).Unix(),
		"aud":   "authenticated",
		"role":  "authenticated",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.jwtSecret))
	if err != nil {
		return fmt.Sprintf("invalid-token-%s", userID)
	}
	return signed
}

// Close, sunucuyu durdurur.
func (m *SupabaseMock) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

// PreRegisterUser, bir kullanıcıyı önceden kaydeder (duplicate email testleri için).
func (m *SupabaseMock) PreRegisterUser(email, password, userID, displayName, slugID string) {
	m.registeredUsers[email] = &preRegisteredUser{
		Email: email, Password: password, UserID: userID,
		DisplayName: displayName, SlugID: slugID,
	}
}

// handleSignup, Supabase signup endpoint'ini mock'lar.
func (m *SupabaseMock) handleSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if _, exists := m.registeredUsers[body.Email]; exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "User already registered",
			"error_description": "User already registered",
		})
		return
	}

	userID := fmt.Sprintf("user-%d", len(m.registeredUsers)+1)
	slugID := fmt.Sprintf("slug%d", len(m.registeredUsers)+1)
	accessToken := m.generateJWT(userID)

	m.registeredUsers[body.Email] = &preRegisteredUser{
		Email: body.Email, Password: body.Password, UserID: userID,
		DisplayName: "", SlugID: slugID, AccessToken: accessToken,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": accessToken,
		"user": map[string]any{
			"id":    userID,
			"email": body.Email,
		},
	})
}

// handleLogin, Supabase token endpoint'ini mock'lar.
func (m *SupabaseMock) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	email := body.Email
	password := body.Password

	user, exists := m.registeredUsers[email]
	if !exists || user.Password != password {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "Invalid login credentials",
			"error_description": "Invalid login credentials",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": user.AccessToken,
		"user": map[string]any{
			"id":    user.UserID,
			"email": user.Email,
		},
	})
}

// handleUser, Supabase /auth/v1/user endpoint'ini mock'lar.
func (m *SupabaseMock) handleUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	for _, user := range m.registeredUsers {
		if user.AccessToken == token {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"id":    user.UserID,
				"email": user.Email,
			})
			return
		}
	}

	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}

// URL, mock sunucunun temel URL'sini döner.
func (m *SupabaseMock) URL() string {
	return strings.TrimRight(m.Server.URL, "/")
}
