package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ──────────────────────────────────────────────────────────────────────────────
// JWT Middleware Testleri
//
// Bu testler, AuthMiddleware'in HTTP isteklerini doğru şekilde doğruladığını
// ve geçersiz/eksik token'ları düzgün reddettiğini doğrular.
// Ayrıca extractBearerToken, parseAndValidateToken, UserIDFromContext
// ve CORS middleware fonksiyonlarını ayrı ayrı test eder.
// ──────────────────────────────────────────────────────────────────────────────

// testSecret, testlerde kullanılan sabit JWT secret.
const testSecret = "test-super-secret-key-that-is-long-enough-for-hs256"

// dummyHandler, middleware'den geçen isteklerde context'teki userID'yi JSON olarak dönen
// basit bir handler. Middleware'in context'e userID'yi doğru yerleştirdiğini doğrular.
var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "context'te userID yok"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
})

// ── Yardımcı Fonksiyonlar ─────────────────────────────────────────────────────

// generateHS256JWT, testler için HMAC-SHA256 ile imzalanmış geçerli bir JWT üretir.
func generateHS256JWT(userID string, exp time.Duration) string {
	claims := supabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		panic("JWT imzalama hatası: " + err.Error())
	}
	return signed
}

// generateExpiredJWT, süresi dolmuş bir HS256 JWT üretir.
func generateExpiredJWT(userID string) string {
	return generateHS256JWT(userID, -1*time.Hour)
}

// generateES256JWT, header'da alg=ES256 olan bir token üretir.
// Gerçek ES256 imzası olmadan claim validasyonunu test eder.
func generateES256JWT(userID string) string {
	claims := supabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["alg"] = "ES256"
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

// generateTokenWithWrongAlg, None algoritması ile token üretir (güvenlik testi).
func generateTokenWithWrongAlg(userID string) string {
	claims := supabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	return signed
}

// ── 1. extractBearerToken Testleri ────────────────────────────────────────────

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantErr   bool
	}{
		{name: "geçerli Bearer token", header: "Bearer xxx.yyy.zzz", wantToken: "xxx.yyy.zzz", wantErr: false},
		{name: "Authorization başlığı yok", header: "", wantToken: "", wantErr: true},
		{name: "Bearer prefix yok", header: "Token xxx.yyy.zzz", wantToken: "", wantErr: true},
		{name: "boş Bearer token", header: "Bearer ", wantToken: "", wantErr: true},
		{name: "Bearer küçük harf", header: "bearer xxx.yyy.zzz", wantToken: "xxx.yyy.zzz", wantErr: false},
		{name: "Bearer karışık harf", header: "BEARER xxx.yyy.zzz", wantToken: "xxx.yyy.zzz", wantErr: false},
		{name: "sadece Bearer", header: "Bearer", wantToken: "", wantErr: true},
		{name: "çok boşluklu", header: "Bearer   xxx.yyy.zzz", wantToken: "xxx.yyy.zzz", wantErr: false},
		{name: "token başında/sonunda boşluk", header: "Bearer  xxx.yyy.zzz  ", wantToken: "xxx.yyy.zzz", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token, err := extractBearerToken(req)
			if tt.wantErr && err == nil {
				t.Error("hata bekleniyordu, nil döndü")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("hata beklenmiyordu, hata: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("beklenen token: %q, gelen: %q", tt.wantToken, token)
			}
		})
	}
}

// ── 2. parseAndValidateToken Testleri ─────────────────────────────────────────

func TestParseAndValidateToken_ValidHS256(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	userID := "550e8400-e29b-41d4-a716-446655440000"
	token := generateHS256JWT(userID, 1*time.Hour)

	result, err := parseAndValidateToken(token)
	if err != nil {
		t.Fatalf("geçerli HS256 token doğrulanamadı: %v", err)
	}
	if result != userID {
		t.Errorf("beklenen userID: %s, gelen: %s", userID, result)
	}
}

func TestParseAndValidateToken_ExpiredHS256(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	token := generateExpiredJWT("550e8400-e29b-41d4-a716-446655440000")

	_, err := parseAndValidateToken(token)
	if err == nil {
		t.Fatal("süresi dolmuş token hata vermeli")
	}
	if !strings.Contains(err.Error(), "token is expired") {
		t.Errorf("beklenen: 'token is expired', gelen: %s", err.Error())
	}
}

func TestParseAndValidateToken_WrongSecret(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	claims := supabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test-user",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	wrongToken, _ := token.SignedString([]byte("wrong-secret"))

	_, err := parseAndValidateToken(wrongToken)
	if err == nil {
		t.Fatal("yanlış secret ile imzalanmış token hata vermeli")
	}
	if !strings.Contains(err.Error(), "signature is invalid") {
		t.Errorf("beklenen: 'signature is invalid', gelen: %s", err.Error())
	}
}

func TestParseAndValidateToken_ES256(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	userID := "550e8400-e29b-41d4-a716-446655440001"
	token := generateES256JWT(userID)

	result, err := parseAndValidateToken(token)
	if err != nil {
		t.Fatalf("ES256 token doğrulanamadı: %v", err)
	}
	if result != userID {
		t.Errorf("beklenen userID: %s, gelen: %s", userID, result)
	}
}

func TestParseAndValidateToken_ExpiredES256(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	claims := supabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test-user",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["alg"] = "ES256"
	signed, _ := token.SignedString([]byte(testSecret))

	_, err := parseAndValidateToken(signed)
	if err == nil {
		t.Fatal("süresi dolmuş ES256 token hata vermeli")
	}
	if !strings.Contains(err.Error(), "token is expired") {
		t.Errorf("beklenen: 'token is expired', gelen: %s", err.Error())
	}
}

func TestParseAndValidateToken_NoneAlgorithm(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	token := generateTokenWithWrongAlg("test-user")

	_, err := parseAndValidateToken(token)
	if err == nil {
		t.Fatal("None algoritması ile token hata vermeli")
	}
}

func TestParseAndValidateToken_InvalidFormat(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	_, err := parseAndValidateToken("this-is-not-a-valid-jwt")
	if err == nil {
		t.Fatal("geçersiz JWT formatı hata vermeli")
	}
}

func TestParseAndValidateToken_EmptyToken(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	_, err := parseAndValidateToken("")
	if err == nil {
		t.Fatal("boş token hata vermeli")
	}
}

func TestParseAndValidateToken_NoSubject(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	claims := supabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))

	_, err := parseAndValidateToken(signed)
	if err == nil {
		t.Fatal("sub claim'i olmayan token hata vermeli")
	}
}

// ── 3. AuthMiddleware Entegrasyon Testleri ────────────────────────────────────

func TestAuthMiddleware_WithValidToken(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	userID := "550e8400-e29b-41d4-a716-446655440000"
	token := generateHS256JWT(userID, 1*time.Hour)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("beklenen: 200, gelen: %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["user_id"] != userID {
		t.Errorf("beklenen user_id: %s, gelen: %s", userID, resp["user_id"])
	}
}

func TestAuthMiddleware_WithExpiredToken(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	token := generateExpiredJWT("550e8400-e29b-41d4-a716-446655440000")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("beklenen: 401, gelen: %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "Geçersiz veya süresi dolmuş oturum. Lütfen tekrar giriş yapın." {
		t.Errorf("beklenen hata mesajı yanlış: %s", resp["error"])
	}
}

func TestAuthMiddleware_WithoutToken(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	req := httptest.NewRequest("GET", "/protected", nil)

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("beklenen: 401, gelen: %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "Yetkilendirme başlığı eksik veya hatalı." {
		t.Errorf("beklenen hata: 'Yetkilendirme başlığı eksik veya hatalı.', gelen: %s", resp["error"])
	}
}

func TestAuthMiddleware_WithInvalidBearerFormat(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	token := generateHS256JWT("test-user", 1*time.Hour)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Token "+token)

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("beklenen: 401, gelen: %d", w.Code)
	}
}

func TestAuthMiddleware_WithMalformedToken(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer this-is-not-a-valid-jwt")

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("beklenen: 401, gelen: %d", w.Code)
	}
}

func TestAuthMiddleware_WithWrongSecretToken(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	claims := supabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test-user",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	wrongToken, _ := token.SignedString([]byte("different-secret"))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+wrongToken)

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("beklenen: 401, gelen: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_ResponseIsJSON(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("beklenen Content-Type: 'application/json; charset=utf-8', gelen: %q", ct)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("yanıt geçerli JSON değil: %v", err)
	}
}

func TestAuthMiddleware_WithES256Token(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testSecret)
	InitJWTSecret()

	userID := "550e8400-e29b-41d4-a716-446655440002"
	token := generateES256JWT(userID)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	AuthMiddleware(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ES256 token ile beklenen: 200, gelen: %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["user_id"] != userID {
		t.Errorf("beklenen user_id: %s, gelen: %s", userID, resp["user_id"])
	}
}

// ── 4. UserIDFromContext Testleri ─────────────────────────────────────────────

func TestUserIDFromContext_Success(t *testing.T) {
	ctx := context.Background()
	ctxWithVal := context.WithValue(ctx, ContextKeyUserID, "550e8400-e29b-41d4-a716-446655440000")

	got, ok := UserIDFromContext(ctxWithVal)
	if !ok {
		t.Error("userID context'te olmasına rağmen ok=false döndü")
	}
	if got != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("beklenen: %s, gelen: %s", "550e8400-e29b-41d4-a716-446655440000", got)
	}
}

func TestUserIDFromContext_EmptyString(t *testing.T) {
	ctx := context.Background()
	ctxWithVal := context.WithValue(ctx, ContextKeyUserID, "")

	_, ok := UserIDFromContext(ctxWithVal)
	if ok {
		t.Error("boş userID için ok=false dönmeli")
	}
}

func TestUserIDFromContext_NoValue(t *testing.T) {
	_, ok := UserIDFromContext(context.Background())
	if ok {
		t.Error("context'te değer yokken ok=false dönmeli")
	}
}

func TestUserIDFromContext_WrongType(t *testing.T) {
	ctx := context.Background()
	ctxWithVal := context.WithValue(ctx, ContextKeyUserID, 12345)

	_, ok := UserIDFromContext(ctxWithVal)
	if ok {
		t.Error("yanlış tip için ok=false dönmeli")
	}
}

// ── 5. CORS Middleware Testleri ───────────────────────────────────────────────

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Access-Control-Allow-Origin: * olmalı")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Access-Control-Allow-Methods boş olmamalı")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Access-Control-Allow-Headers boş olmamalı")
	}
	if w.Header().Get("Access-Control-Max-Age") != "86400" {
		t.Error("Access-Control-Max-Age: 86400 olmalı")
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("OPTIONS isteği handler'a ulaşmamalı")
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("beklenen: 204, gelen: %d", w.Code)
	}
}
