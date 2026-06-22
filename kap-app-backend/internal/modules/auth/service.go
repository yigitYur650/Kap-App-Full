package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"assignment-backend/pkg/utils"
)

// UserProfile, auth modülünün profil modelidir.
// JSON tag'leri frontend ile sözleşmeyi oluşturur.
type UserProfile struct {
	UserID      string `json:"user_id"`
	SlugID      string `json:"slug_id"`
	DisplayName string `json:"display_name"`
	IsNew       bool   `json:"is_new"` // true → bu çağrıda yeni oluşturuldu
}

// Service, auth modülünün iş mantığı katmanıdır.
type Service struct {
	repo *Repository
}

// NewService, Service örneği oluşturur.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SupabaseError, Supabase'den dönen hataları sarmalar.
type SupabaseError struct {
	StatusCode int
	Message    string
}

func (e *SupabaseError) Error() string {
	return e.Message
}

// SupabaseAuthResponse, Supabase token ve user cevabını modeller.
type SupabaseAuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         SupabaseUser `json:"user"`
}

type SupabaseUser struct {
	ID string `json:"id"`
}

// Register, Supabase Auth'a yeni kullanıcı kaydeder (ANON_KEY ile).
func (s *Service) Register(ctx context.Context, email, password, name string) (*SupabaseAuthResponse, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if supabaseURL == "" || anonKey == "" {
		return nil, errors.New("Supabase URL veya Anon Key çevre değişkenleri eksik")
	}

	url := fmt.Sprintf("%s/auth/v1/signup", strings.TrimRight(supabaseURL, "/"))

	// Body hazırla
	reqBodyMap := map[string]any{
		"email":    email,
		"password": password,
		"options": map[string]any{
			"data": map[string]any{
				"full_name": name,
			},
		},
	}
	jsonBytes, err := json.Marshal(reqBodyMap)
	if err != nil {
		return nil, fmt.Errorf("json marshal hatası: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("http isteği oluşturulamadı: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Authorization", "Bearer "+anonKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Supabase signup isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Supabase yanıtı okunamadı: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseSupabaseError(resp.StatusCode, respBytes)
	}

	var authResp SupabaseAuthResponse
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		return nil, fmt.Errorf("Supabase yanıtı unmarshal edilemedi: %w", err)
	}

	return &authResp, nil
}

// Login, Supabase Auth ile kullanıcı girişi yapar (ANON_KEY ile).
func (s *Service) Login(ctx context.Context, email, password string) (*SupabaseAuthResponse, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if supabaseURL == "" || anonKey == "" {
		return nil, errors.New("Supabase URL veya Anon Key çevre değişkenleri eksik")
	}

	url := fmt.Sprintf("%s/auth/v1/token?grant_type=password", strings.TrimRight(supabaseURL, "/"))

	reqBodyMap := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonBytes, err := json.Marshal(reqBodyMap)
	if err != nil {
		return nil, fmt.Errorf("json marshal hatası: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("http isteği oluşturulamadı: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Authorization", "Bearer "+anonKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Supabase login isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Supabase yanıtı okunamadı: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseSupabaseError(resp.StatusCode, respBytes)
	}

	var authResp SupabaseAuthResponse
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		return nil, fmt.Errorf("Supabase yanıtı unmarshal edilemedi: %w", err)
	}

	return &authResp, nil
}

// parseSupabaseError, Supabase'den gelen hata JSON'unu okur ve kullanıcı dostu mesaj üretir.
func parseSupabaseError(statusCode int, body []byte) error {
	var errMap map[string]any
	if err := json.Unmarshal(body, &errMap); err == nil {
		if desc, ok := errMap["error_description"].(string); ok && desc != "" {
			return &SupabaseError{StatusCode: statusCode, Message: translateSupabaseError(desc)}
		}
		if msg, ok := errMap["msg"].(string); ok && msg != "" {
			return &SupabaseError{StatusCode: statusCode, Message: translateSupabaseError(msg)}
		}
		if message, ok := errMap["message"].(string); ok && message != "" {
			return &SupabaseError{StatusCode: statusCode, Message: translateSupabaseError(message)}
		}
		if errStr, ok := errMap["error"].(string); ok && errStr != "" {
			return &SupabaseError{StatusCode: statusCode, Message: translateSupabaseError(errStr)}
		}
	}
	return &SupabaseError{StatusCode: statusCode, Message: fmt.Sprintf("Supabase hatası (Status %d)", statusCode)}
}

// translateSupabaseError, yaygın Supabase hata mesajlarını Türkçe'ye çevirir.
func translateSupabaseError(msg string) string {
	lowerMsg := strings.ToLower(msg)
	switch {
	case strings.Contains(lowerMsg, "invalid login credentials") || strings.Contains(lowerMsg, "invalid credentials"):
		return "E-posta veya şifre hatalı."
	case strings.Contains(lowerMsg, "user already registered") || strings.Contains(lowerMsg, "email already exists"):
		return "Bu e-posta adresiyle kayıtlı bir kullanıcı zaten var."
	case strings.Contains(lowerMsg, "signup disabled"):
		return "Kayıt işlemleri şu an devre dışı."
	case strings.Contains(lowerMsg, "email address not confirmed") || strings.Contains(lowerMsg, "email not confirmed"):
		return "Lütfen e-posta adresinizi onaylayın."
	case strings.Contains(lowerMsg, "password should be at least"):
		return "Şifre en az 6 karakter olmalıdır."
	case strings.Contains(lowerMsg, "rate limit"):
		return "Çok fazla deneme yaptınız. Lütfen daha sonra tekrar deneyin."
	default:
		return msg
	}
}

// SyncProfile, Supabase Auth'tan gelen userID + display_name için:
//  1. profiles tablosunda kayıt var mı kontrol eder.
//  2. Varsa mevcut profili döner (display_name güncellenmez — idempotent).
//  3. Yoksa benzersiz slug_id üretip kaydeder (max maxRetries deneme).
func (s *Service) SyncProfile(ctx context.Context, userID, displayName string) (*UserProfile, error) {
	const maxRetries = 5

	// 1. Var olan profili getir.
	existing, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.SyncProfile: FindByUserID: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// 2. Yeni profil: benzersiz slug_id üret, çakışırsa yeniden dene.
	for attempt := 1; attempt <= maxRetries; attempt++ {
		slug, err := utils.GenerateSlugID()
		if err != nil {
			return nil, fmt.Errorf("auth.Service.SyncProfile: GenerateSlugID (deneme %d): %w", attempt, err)
		}

		profile, err := s.repo.Create(ctx, userID, slug, displayName)
		if err == nil {
			log.Printf(
				"auth.Service.SyncProfile: yeni profil oluşturuldu user_id=%s slug=%s display_name=%q",
				userID, slug, displayName,
			)
			return profile, nil
		}

		if isUniqueViolation(err) {
			log.Printf(
				"auth.Service.SyncProfile: slug çakışması, yeniden deneniyor (%d/%d): %v",
				attempt, maxRetries, err,
			)
			continue
		}

		// Farklı bir DB hatası → sarmala ve döndür.
		return nil, fmt.Errorf("auth.Service.SyncProfile: Create (deneme %d): %w", attempt, err)
	}

	return nil, fmt.Errorf("auth.Service.SyncProfile: %d denemede benzersiz slug üretilemedi", maxRetries)
}
