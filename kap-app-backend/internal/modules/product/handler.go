package product

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"assignment-backend/internal/middleware"
)

// Handler, product modülünün HTTP katmanını taşır.
type Handler struct {
	svc *Service
}

// NewHandler, Handler örneği oluşturur.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes, product rotalarını mux'a kaydeder.
// Tüm rotalar AuthMiddleware arkasında korunur.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/products",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleAdd)))

	mux.Handle("GET /api/v1/products",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleList)))

	mux.Handle("PATCH /api/v1/products/{id}/status",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleUpdateStatus)))

	mux.Handle("DELETE /api/v1/products/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleDelete)))
}

// ── İstek gövdesi tipleri ─────────────────────────────────────────────────────

type addProductRequest struct {
	TenantID       string   `json:"tenant_id"`
	Name           string   `json:"name"`
	Price          *float64 `json:"price"`
	MarketName     *string  `json:"market_name"`
	Category       *string  `json:"category"`
	Quantity       *int     `json:"quantity"`
	ExpirationDate *string  `json:"expiration_date"` // YYYY-MM-DD
	Unit           *string  `json:"unit"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// ── Handler'lar ───────────────────────────────────────────────────────────────

// handleAdd → POST /api/v1/products
func (h *Handler) handleAdd(w http.ResponseWriter, r *http.Request) {
	userID, ok := mustUserID(w, r)
	if !ok {
		return
	}

	var req addProductRequest
	if !decodeBody(w, r, &req) {
		return
	}

	product, err := h.svc.AddProduct(r.Context(), userID, AddProductInput{
		TenantID:       strings.TrimSpace(req.TenantID),
		Name:           req.Name,
		Price:          req.Price,
		MarketName:     req.MarketName,
		Category:       req.Category,
		Unit:           req.Unit,
		Quantity:       req.Quantity,
		ExpirationDate: req.ExpirationDate,
	})
	if err != nil {
		handleServiceError(w, r, "product.handleAdd", err)
		return
	}

	writeJSON(w, http.StatusCreated, product)
}

// handleList → GET /api/v1/products?tenant_id=...
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := mustUserID(w, r)
	if !ok {
		return
	}

	tenantID := strings.TrimSpace(r.URL.Query().Get("tenant_id"))
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, errBody("tenant_id query parametresi zorunludur."))
		return
	}

	list, err := h.svc.ListProducts(r.Context(), userID, tenantID)
	if err != nil {
		handleServiceError(w, r, "product.handleList", err)
		return
	}

	writeJSON(w, http.StatusOK, list)
}

// handleUpdateStatus → PATCH /api/v1/products/{id}/status
func (h *Handler) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := mustUserID(w, r)
	if !ok {
		return
	}

	productID := strings.TrimSpace(r.PathValue("id"))
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, errBody("Ürün ID'si eksik."))
		return
	}

	var req updateStatusRequest
	if !decodeBody(w, r, &req) {
		return
	}

	updated, err := h.svc.UpdateStatus(r.Context(), userID, productID, strings.TrimSpace(req.Status))
	if err != nil {
		handleServiceError(w, r, "product.handleUpdateStatus", err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// handleDelete → DELETE /api/v1/products/{id}
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := mustUserID(w, r)
	if !ok {
		return
	}

	productID := strings.TrimSpace(r.PathValue("id"))
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, errBody("Ürün ID'si eksik."))
		return
	}

	if err := h.svc.DeleteProduct(r.Context(), userID, productID); err != nil {
		handleServiceError(w, r, "product.handleDelete", err)
		return
	}

	// 204 No Content — gövde yok.
	w.WriteHeader(http.StatusNoContent)
}

// ── Yardımcılar ──────────────────────────────────────────────────────────────

func mustUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("yetkisiz"))
	}
	return id, ok
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		log.Printf("product.decodeBody: okuma hatası: %v", err)
		writeJSON(w, http.StatusBadRequest, errBody("İstek gövdesi okunamadı."))
		return false
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		log.Printf("product.decodeBody: json hatası: %v", err)
		writeJSON(w, http.StatusBadRequest, errBody("Geçersiz JSON formatı."))
		return false
	}
	return true
}

// handleServiceError, servis hatalarını HTTP kodlarına dönüştürür.
func handleServiceError(w http.ResponseWriter, r *http.Request, caller string, err error) {
	var ve ValidationError
	switch {
	case errors.As(err, &ve):
		writeJSON(w, http.StatusBadRequest, errBody(ve.Error()))
	case errors.Is(err, ErrForbidden):
		writeJSON(w, http.StatusForbidden, errBody("Bu işlem için yetkiniz yok."))
	case errors.Is(err, ErrProductNotFound):
		writeJSON(w, http.StatusNotFound, errBody("Ürün bulunamadı."))
	default:
		log.Printf("%s: %v", caller, err)
		writeJSON(w, http.StatusInternalServerError, errBody("İşlem sırasında bir hata oluştu."))
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("product.writeJSON: encode hatası: %v", err)
	}
}

func errBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
