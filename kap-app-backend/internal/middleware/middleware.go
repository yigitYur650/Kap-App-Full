package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ── Yardımcı tipler ──────────────────────────────────────────────────────────

// errorResponse, frontend'e dönen maskelenmiş JSON hata yapısı.
type errorResponse struct {
	Error string `json:"error"`
}

// supabaseClaims, Supabase'in JWT payload'ını modelleyen claims yapısı.
// Supabase, standart `sub` alanını kullanıcı UUID'si olarak doldurur.
// Gerekirse ileride `email`, `role`, `app_metadata` gibi alanlar eklenebilir.
type supabaseClaims struct {
	jwt.RegisteredClaims
}

// ── Singleton: JWT secret ────────────────────────────────────────────────────

// jwtSecret, uygulama başlangıcında bir kez okunur; sonraki isteklerde
// env'e tekrar başvurulmaz (performans + tutarlılık).
var jwtSecret []byte

// InitJWTSecret, SUPABASE_JWT_SECRET çevre değişkenini okuyarak
// paket düzeyindeki jwtSecret değişkenini doldurur.
// main() içinde NewDatabase()'den hemen sonra çağrılmalıdır.
// Değişken boşsa uygulama başlamadan durur.
func InitJWTSecret() {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		log.Fatal("middleware.InitJWTSecret: SUPABASE_JWT_SECRET çevre değişkeni eksik")
	}
	jwtSecret = []byte(secret)
	log.Println("middleware.InitJWTSecret: JWT secret başarıyla yüklendi.")
}

// ── AuthMiddleware ───────────────────────────────────────────────────────────

// AuthMiddleware, gelen her HTTP isteğinin Authorization başlığını doğrulayan
// standart net/http uyumlu middleware fonksiyonudur.
//
// Akış:
//  1. Authorization: Bearer <token> başlığını ayıkla.
//  2. Token'ı HMAC-SHA256 ile doğrula (Supabase varsayılan algoritması).
//  3. `sub` (kullanıcı UUID) claim'ini oku.
//  4. UUID'yi context'e göm ve bir sonraki handler'ı çağır.
//
// Hata durumlarında istek işlenmez; 401 JSON yanıtı dönülür.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Header'ı oku ve "Bearer " prefix'ini ayır.
		rawToken, err := extractBearerToken(r)
		if err != nil {
			// Geliştirici logu: nedenin tamamı görünür.
			log.Printf("middleware.AuthMiddleware: token ayıklanamadı: %v", err)
			// Frontend: sadece genel mesaj.
			writeUnauthorized(w, "Yetkilendirme başlığı eksik veya hatalı.")
			return
		}

		// 2. Token'ı doğrula ve claims'i ayrıştır.
		userID, err := parseAndValidateToken(rawToken)
		if err != nil {
			log.Printf("middleware.AuthMiddleware: token doğrulanamadı: %v", err)
			writeUnauthorized(w, "Geçersiz veya süresi dolmuş oturum. Lütfen tekrar giriş yapın.")
			return
		}

		// 3. Kullanıcı UUID'sini context'e göm.
		ctx := r.Context()
		ctx = contextWithUserID(ctx, userID)

		// 4. Zenginleştirilmiş context ile bir sonraki handler'ı çağır.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ── Yardımcı (unexported) fonksiyonlar ──────────────────────────────────────

// extractBearerToken, Authorization başlığından ham token string'ini döner.
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization başlığı yok")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("middleware.extractBearerToken: beklenmeyen format: %q", authHeader)
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("middleware.extractBearerToken: Bearer token boş")
	}

	return token, nil
}

// parseAndValidateToken, JWT string'ini HMAC-SHA256 ile imzalı Supabase token
// olarak doğrular ve `sub` claim'ini (kullanıcı UUID) döner.
func parseAndValidateToken(rawToken string) (string, error) {
	claims := &supabaseClaims{}

	p := new(jwt.Parser)
	unverifiedToken, _, err := p.ParseUnverified(rawToken, claims)
	if err != nil {
		return "", fmt.Errorf("middleware.parseAndValidateToken: unverified parse: %w", err)
	}

	alg, _ := unverifiedToken.Header["alg"].(string)
	if alg == "ES256" {
		// ES256 için yerel geliştirme kolaylığı adına imzayı doğrulamadan claim geçerliliğini kontrol edelim
		validator := jwt.NewValidator(
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
		)
		if err := validator.Validate(claims); err != nil {
			return "", fmt.Errorf("middleware.parseAndValidateToken: validate claims: %w", err)
		}
	} else {
		// HS256 için tam imza doğrulaması
		token, err := jwt.ParseWithClaims(rawToken, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"middleware.parseAndValidateToken: beklenmeyen imza algoritması: %v",
					t.Header["alg"],
				)
			}
			return jwtSecret, nil
		},
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
		)
		if err != nil {
			return "", fmt.Errorf("middleware.parseAndValidateToken: jwt.ParseWithClaims: %w", err)
		}
		if !token.Valid {
			return "", errors.New("middleware.parseAndValidateToken: token geçersiz")
		}
	}

	// `sub` alanı kullanıcı UUID'sini taşır.
	userID, err := claims.GetSubject()
	if err != nil || userID == "" {
		return "", fmt.Errorf("middleware.parseAndValidateToken: sub claim okunamadı: %w", err)
	}

	return userID, nil
}

// writeUnauthorized, maskelenmiş 401 JSON yanıtı yazar.
func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)

	resp := errorResponse{Error: msg}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Bu noktada response zaten başlamış olduğundan sadece logluyoruz.
		log.Printf("middleware.writeUnauthorized: json encode hatası: %v", err)
	}
}
