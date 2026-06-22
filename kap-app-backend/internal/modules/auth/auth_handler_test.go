package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"assignment-backend/internal/middleware"
)

func testRequest(handler http.HandlerFunc, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestHandleRegister_Success(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	w := testRequest(handler.handleRegister, "POST", "/api/v1/auth/register", registerRequest{
		Email: "handler-test@example.com", Password: "password123", Name: "Handler Test",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("beklenen status: 200, gelen: %d, body: %s", w.Code, w.Body.String())
	}
	var resp authResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("yanit JSON cozumlenemedi: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("access_token bos gelmemeli")
	}
	if resp.UserID == "" {
		t.Error("user_id bos gelmemeli")
	}
}

func TestHandleRegister_EmptyFields(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	tests := []struct {
		name string
		body registerRequest
	}{
		{"bos email", registerRequest{Email: "", Password: "pass", Name: "Test"}},
		{"bos sifre", registerRequest{Email: "a@b.com", Password: "", Name: "Test"}},
		{"bos isim", registerRequest{Email: "a@b.com", Password: "pass", Name: ""}},
		{"hepsi bos", registerRequest{Email: "", Password: "", Name: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testRequest(handler.handleRegister, "POST", "/api/v1/auth/register", tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("beklenen: 400, gelen: %d", w.Code)
			}
			var errResp map[string]string
			json.Unmarshal(w.Body.Bytes(), &errResp)
			if errResp["error"] == "" {
				t.Error("hata mesaji bos olmamali")
			}
		})
	}
}

func TestHandleRegister_InvalidJSON(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString("{{gecersiz json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.handleRegister(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("beklenen: 400, gelen: %d", w.Code)
	}
}

func TestHandleRegister_DuplicateEmail(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	w1 := testRequest(handler.handleRegister, "POST", "/api/v1/auth/register", registerRequest{
		Email: "dupe-handler@example.com", Password: "pass123", Name: "Ilk",
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("Ilk kayit basarisiz: %d", w1.Code)
	}
	w2 := testRequest(handler.handleRegister, "POST", "/api/v1/auth/register", registerRequest{
		Email: "dupe-handler@example.com", Password: "pass456", Name: "Ikinci",
	})
	if w2.Code != http.StatusBadRequest {
		t.Errorf("beklenen: 400, gelen: %d", w2.Code)
	}
	var errResp map[string]string
	json.Unmarshal(w2.Body.Bytes(), &errResp)
	if errResp["error"] != "" {
		t.Logf("hata mesaji: %s", errResp["error"])
	}
}

func TestHandleLogin_Success(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	wReg := testRequest(handler.handleRegister, "POST", "/api/v1/auth/register", registerRequest{
		Email: "login-handler@example.com", Password: "mypassword", Name: "Login Test",
	})
	if wReg.Code != http.StatusOK {
		t.Fatalf("Kayit basarisiz: %d", wReg.Code)
	}

	w := testRequest(handler.handleLogin, "POST", "/api/v1/auth/login", loginRequest{
		Email: "login-handler@example.com", Password: "mypassword",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("beklenen: 200, gelen: %d, body: %s", w.Code, w.Body.String())
	}
	var resp authResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("yanit cozulmelenemedi: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("access_token bos gelmemeli")
	}
	if resp.UserID == "" {
		t.Error("user_id bos gelmemeli")
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	testRequest(handler.handleRegister, "POST", "/api/v1/auth/register", registerRequest{
		Email: "wrong-pw-handler@example.com", Password: "dogru-sifre", Name: "PW Test",
	})

	w := testRequest(handler.handleLogin, "POST", "/api/v1/auth/login", loginRequest{
		Email: "wrong-pw-handler@example.com", Password: "yanlis-sifre",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("beklenen: 400, gelen: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestHandleLogin_EmptyFields(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	tests := []struct {
		name string
		body loginRequest
	}{
		{"bos email", loginRequest{Email: "", Password: "pass"}},
		{"bos sifre", loginRequest{Email: "a@b.com", Password: ""}},
		{"ikisi de bos", loginRequest{Email: "", Password: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testRequest(handler.handleLogin, "POST", "/api/v1/auth/login", tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("beklenen: 400, gelen: %d", w.Code)
			}
		})
	}
}

func testSyncProfileRequest(handler http.HandlerFunc, method, path string, body any, userID string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestHandleSyncProfile_EmptyDisplayName(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	w := testSyncProfileRequest(handler.handleSyncProfile, "POST", "/api/v1/auth/sync-profile",
		syncProfileRequest{DisplayName: ""}, "test-user-id")
	if w.Code != http.StatusBadRequest {
		t.Errorf("beklenen: 400, gelen: %d", w.Code)
	}
}

func TestHandleSyncProfile_TooLongDisplayName(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	longName := strings.Repeat("a", 51)
	w := testSyncProfileRequest(handler.handleSyncProfile, "POST", "/api/v1/auth/sync-profile",
		syncProfileRequest{DisplayName: longName}, "test-user-id")
	if w.Code != http.StatusBadRequest {
		t.Errorf("beklenen: 400, gelen: %d", w.Code)
	}
}

func TestHandleSyncProfile_Unauthenticated(t *testing.T) {
	mock := NewSupabaseMock()
	defer mock.Close()
	t.Setenv("SUPABASE_URL", mock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")

	svc := NewService(nil)
	handler := NewHandler(svc)

	w := testRequest(handler.handleSyncProfile, "POST", "/api/v1/auth/sync-profile",
		syncProfileRequest{DisplayName: "Test User"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("beklenen: 401, gelen: %d", w.Code)
	}
}

func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"bos isim", "", true, "display_name boş olamaz."},
		{"50 karakter (gecerli)", strings.Repeat("a", 50), false, ""},
		{"51 karakter (gecersiz)", strings.Repeat("a", 51), true, "display_name en fazla 50 karakter olabilir."},
		{"normal isim", "Yigit Yur", false, ""},
		{"tek karakter", "A", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDisplayName(tt.input)
			if tt.wantErr && err == nil {
				t.Error("hata bekleniyordu, nil dondu")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("hata beklenmiyordu, hata: %v", err)
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("beklenen: %q, gelen: %q", tt.errMsg, err.Error())
			}
		})
	}
}
