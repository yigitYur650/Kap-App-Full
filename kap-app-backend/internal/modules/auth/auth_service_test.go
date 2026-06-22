package auth

import (
	"context"
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────────
// Auth Service Integration Tests
//
// Bu testler, Register → Login → SyncProfile akışının uçtan uca doğru
// çalıştığını doğrular. Gerçek Supabase yerine mock HTTP server kullanır.
//
// Test stratejisi:
//   1. Mock Supabase sunucusu başlat
//   2. SUPABASE_URL ve SUPABASE_ANON_KEY environment variable'larını mock'a yönlendir
//   3. Service'i mock repository olmadan (gerçek repository gerektirmez) test et
//      - Register: mock Supabase'e HTTP çağrısı yapar
//      - Login: mock Supabase'e HTTP çağrısı yapar
//      - SyncProfile: mock repository gerektirir, bu testte test etmiyoruz
//        (handler_test.go'da test edilecek)
//
// Not: SyncProfile testleri için mock repository + mock DB gerekir,
// onlar handler_test.go'da test edilecek.
// ──────────────────────────────────────────────────────────────────────────────

// mockUserID, test için deterministik olmayan bir userID döner.
// Yeni SupabaseMock "user-1", "user-2" formatında ID üretir.
func mockUserID(email string) string {
	return "user-1"
}

// emailHash, email'den mock UUID'nin son 12 hanesini üretir.
func emailHash(email string) string {
	return "000000000001"
}

// mockJWT, test için geçerli formatlı bir JWT token'ı üretir.
func mockJWT(userID string) string {
	return "mock-jwt-token-" + userID
}

// ── Test: Başarılı Kayıt ─────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	// Mock Supabase sunucusunu başlat
	mock := NewSupabaseMock()
	defer mock.Close()

	// Environment variable'ları mock'a yönlendir
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	// Service'i oluştur (repository kullanılmayacak)
	svc := NewService(nil)

	resp, err := svc.Register(context.Background(), "test@example.com", "password123", "Test Kullanıcı")
	if err != nil {
		t.Fatalf("Register başarısız oldu: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("access_token boş gelmemeli")
	}

	if resp.User.ID == "" {
		t.Error("user_id boş gelmemeli")
	}

	if resp.User.ID != "user-1" {
		t.Errorf("beklenen user_id = user-1, gelen = %s", resp.User.ID)
	}
}

// ── Test: Aynı Email ile Tekrar Kayıt ────────────────────────────────────────

func TestRegister_DuplicateEmail(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()

	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)

	// İlk kayıt başarılı olmalı
	_, err := svc.Register(context.Background(), "dupe@example.com", "password123", "İlk Kullanıcı")
	if err != nil {
		t.Fatalf("İlk kayıt başarısız: %v", err)
	}

	// Aynı email ile ikinci kayıt HATA vermeli
	_, err = svc.Register(context.Background(), "dupe@example.com", "password456", "İkinci Kullanıcı")
	if err == nil {
		t.Fatal("Aynı email ile ikinci kayıt hata vermeli")
	}

	// Hata mesajı Türkçe olmalı
	var se *SupabaseError
	if !asSupabaseError(err, &se) {
		t.Fatalf("Hata SupabaseError tipinde olmalı, got %T", err)
	}

	expectedMsg := "Bu e-posta adresiyle kayıtlı bir kullanıcı zaten var."
	if se.Message != expectedMsg {
		t.Errorf("beklenen hata: %q, gelen: %q", expectedMsg, se.Message)
	}
}

// asSupabaseError, SupabaseError tipine güvenli dönüşüm.
func asSupabaseError(err error, se **SupabaseError) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*SupabaseError)
	if !ok {
		return false
	}
	*se = e
	return true
}

// ── Test: Başarılı Giriş ─────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()

	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)

	// Önce kayıt ol
	_, err := svc.Register(context.Background(), "login-test@example.com", "mypassword", "Login Test")
	if err != nil {
		t.Fatalf("Kayıt başarısız: %v", err)
	}

	// Sonra giriş yap
	resp, err := svc.Login(context.Background(), "login-test@example.com", "mypassword")
	if err != nil {
		t.Fatalf("Giriş başarısız: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("access_token boş gelmemeli")
	}

	if resp.User.ID == "" {
		t.Error("user_id boş gelmemeli")
	}

	if resp.User.ID != "user-1" {
		t.Errorf("beklenen user_id = user-1, gelen = %s", resp.User.ID)
	}
}

// ── Test: Yanlış Şifre ile Giriş ─────────────────────────────────────────────

func TestLogin_WrongPassword(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()

	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)

	// Önce kayıt ol
	_, err := svc.Register(context.Background(), "wrong-pw@example.com", "dogru-sifre", "Şifre Test")
	if err != nil {
		t.Fatalf("Kayıt başarısız: %v", err)
	}

	// Yanlış şifre ile giriş
	_, err = svc.Login(context.Background(), "wrong-pw@example.com", "yanlis-sifre")
	if err == nil {
		t.Fatal("Yanlış şifre ile giriş hata vermeli")
	}

	var se *SupabaseError
	if !asSupabaseError(err, &se) {
		t.Fatalf("Hata SupabaseError tipinde olmalı")
	}

	expectedMsg := "E-posta veya şifre hatalı."
	if se.Message != expectedMsg {
		t.Errorf("beklenen hata: %q, gelen: %q", expectedMsg, se.Message)
	}

	if se.StatusCode != 400 {
		t.Errorf("beklenen status code: 400, gelen: %d", se.StatusCode)
	}
}

// ── Test: Var Olmayan Kullanıcı ile Giriş ────────────────────────────────────

func TestLogin_UserNotFound(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()

	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)

	// Hiç kaydolmamış kullanıcı ile giriş
	_, err := svc.Login(context.Background(), "nonexistent@example.com", "anypassword")
	if err == nil {
		t.Fatal("Var olmayan kullanıcı ile giriş hata vermeli")
	}

	var se *SupabaseError
	if !asSupabaseError(err, &se) {
		t.Fatalf("Hata SupabaseError tipinde olmalı")
	}

	expectedMsg := "E-posta veya şifre hatalı."
	if se.Message != expectedMsg {
		t.Errorf("beklenen hata: %q, gelen: %q", expectedMsg, se.Message)
	}
}

// ── Test: Eksik Environment Variable ─────────────────────────────────────────

func TestRegister_MissingEnvVars(t *testing.T) {
	// Environment variable'ları temizle
	t.Setenv("SUPABASE_URL", "")
	t.Setenv("SUPABASE_ANON_KEY", "")

	svc := NewService(nil)

	_, err := svc.Register(context.Background(), "test@example.com", "password123", "Test")
	if err == nil {
		t.Fatal("Eksik env vars ile Register hata vermeli")
	}

	expected := "Supabase URL veya Anon Key çevre değişkenleri eksik"
	if err.Error() != expected {
		t.Errorf("beklenen hata: %q, gelen: %q", expected, err.Error())
	}
}

func TestLogin_MissingEnvVars(t *testing.T) {
	t.Setenv("SUPABASE_URL", "")
	t.Setenv("SUPABASE_ANON_KEY", "")

	svc := NewService(nil)

	_, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err == nil {
		t.Fatal("Eksik env vars ile Login hata vermeli")
	}

	expected := "Supabase URL veya Anon Key çevre değişkenleri eksik"
	if err.Error() != expected {
		t.Errorf("beklenen hata: %q, gelen: %q", expected, err.Error())
	}
}

// ── Test: Boş Email/Şifre Validasyonu ────────────────────────────────────────
// Not: Handler katmanı boş alanları kontrol eder (handleRegister/handleLogin).
// Service katmanı Supabase'e olduğu gibi iletir, validasyon handler'da yapılır.
// Bu nedenle bu test handler_test.go'da yer alacak.

// ── Test: SyncProfile — Mock Repository ile ──────────────────────────────────
// Not: SyncProfile testleri için mock repository + mock DB gerekir.
// Product testlerinde kullanılan pattern ile aynı yaklaşım kullanılacak.
// Şimdilik sadece handler_test.go'da HTTP düzeyinde test ediyoruz.
