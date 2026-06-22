package tenant

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"assignment-backend/internal/middleware"
)

// Handler, tenant modülünün HTTP katmanını taşır.
type Handler struct {
	svc *Service
}

// NewHandler, Handler örneği oluşturur.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes, tenant rotalarını mux'a kaydeder.
// Tüm rotalar AuthMiddleware arkasında korunur.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/tenants",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleCreate)))

	mux.Handle("GET /api/v1/tenants",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleList)))

	mux.Handle("POST /api/v1/tenants/{id}/members",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleAddMember)))
}

// ── İstek gövdesi tipleri ────────────────────────────────────────────────────

type createTenantRequest struct {
	Name       string `json:"name"`
	ThemeColor string `json:"theme_color"`
}

type addMemberRequest struct {
	SlugID string `json:"slug_id"`
}

// ── Handler'lar ───────────────────────────────────────────────────────────────

// handleCreate → POST /api/v1/tenants
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := mustUserID(w, r)
	if !ok {
		return
	}

	var req createTenantRequest
	if !decodeBody(w, r, &req) {
		return
	}

	tenant, err := h.svc.CreateTenant(r.Context(), userID, req.Name, req.ThemeColor)
	if err != nil {
		handleServiceError(w, r, "tenant.handleCreate", err)
		return
	}

	writeJSON(w, http.StatusCreated, tenant)
}

// handleList → GET /api/v1/tenants
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := mustUserID(w, r)
	if !ok {
		return
	}

	tenants, err := h.svc.ListTenants(r.Context(), userID)
	if err != nil {
		log.Printf("tenant.handleList: %v", err)
		writeJSON(w, http.StatusInternalServerError, errBody("Ev listesi alınamadı."))
		return
	}

	writeJSON(w, http.StatusOK, tenants)
}

// handleAddMember → POST /api/v1/tenants/{id}/members
func (h *Handler) handleAddMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := mustUserID(w, r)
	if !ok {
		return
	}

	// Go 1.22 path value — {id} segmentini al.
	tenantID := strings.TrimSpace(r.PathValue("id"))
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, errBody("Ev ID'si eksik."))
		return
	}

	var req addMemberRequest
	if !decodeBody(w, r, &req) {
		return
	}

	result, err := h.svc.AddMember(r.Context(), tenantID, userID, req.SlugID)
	if err != nil {
		handleServiceError(w, r, "tenant.handleAddMember", err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ── Yardımcılar ──────────────────────────────────────────────────────────────

// mustUserID, context'ten userID okur; başarısızsa 401 yazar ve false döner.
func mustUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("yetkisiz"))
	}
	return id, ok
}

// decodeBody, istek gövdesini (max 1 MB) v'ye JSON olarak çözer.
// Hata durumunda 400 yazar ve false döner.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		log.Printf("tenant.decodeBody: okuma hatası: %v", err)
		writeJSON(w, http.StatusBadRequest, errBody("İstek gövdesi okunamadı."))
		return false
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		log.Printf("tenant.decodeBody: json hatası: %v", err)
		writeJSON(w, http.StatusBadRequest, errBody("Geçersiz JSON formatı."))
		return false
	}
	return true
}

// handleServiceError, servis katmanı hatalarını HTTP kodlarına dönüştürür.
// ValidationError → 400, ErrForbidden → 403, ErrNotFound → 404, diğerleri → 500.
func handleServiceError(w http.ResponseWriter, r *http.Request, caller string, err error) {
	var ve ValidationError
	switch {
	case errors.As(err, &ve):
		writeJSON(w, http.StatusBadRequest, errBody(ve.Error()))

	case errors.Is(err, ErrForbidden):
		writeJSON(w, http.StatusForbidden, errBody("Bu işlem için yetkiniz yok."))

	case errors.Is(err, ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, errBody("Bu kısa ID'ye sahip kullanıcı bulunamadı."))

	case errors.Is(err, ErrAlreadyMember):
		writeJSON(w, http.StatusConflict, errBody("Kullanıcı bu eve zaten üye."))

	default:
		log.Printf("%s: %v", caller, err)
		writeJSON(w, http.StatusInternalServerError, errBody("İşlem sırasında bir hata oluştu."))
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("tenant.writeJSON: encode hatası: %v", err)
	}
}

func errBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
