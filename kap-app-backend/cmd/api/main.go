package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"assignment-backend/internal/config"
	"assignment-backend/internal/middleware"
	"assignment-backend/internal/modules/auth"
	"assignment-backend/internal/modules/product"
	"assignment-backend/internal/modules/tenant"
)

// loadEnvFile, harici kütüphane kullanmadan .env dosyasını satır satır okur
// ve çevre değişkenlerini sisteme enjekte eder.
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("main.go: loadEnvFile: .env dosyası bulunamadı, sistem çevre değişkenleri kullanılacak.")
			return
		}
		log.Fatalf("main.go: loadEnvFile: .env dosyası açılamadı - %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("main.go: loadEnvFile: geçersiz satır formatı (satır %d) - '=' bulunamadı", lineNum)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Çift veya tek tırnakları temizle
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}

		if err := os.Setenv(key, val); err != nil {
			log.Fatalf("main.go: loadEnvFile: çevre değişkeni set edilemedi (%s) - %v", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("main.go: loadEnvFile: dosya okuma hatası - %v", err)
	}
	log.Println("main.go: loadEnvFile: .env dosyası başarıyla yüklendi.")
}

func main() {
	// 0. Çevre değişkenlerini .env dosyasından yükle.
	loadEnvFile(".env")

	// 1. Veritabanı bağlantı havuzunu başlat.
	db := config.NewDatabase()
	defer db.Close()

	// 2. JWT secret'ı çevre değişkeninden yükle.
	middleware.InitJWTSecret()

	// 3. Uygulama context'i — SIGINT/SIGTERM gelince iptal edilir.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. Router'ı kur ve rotaları bağla; GC goroutine'i başlat.
	mux := http.NewServeMux()
	productRepo := registerRoutes(mux, db)
	product.StartGarbageCollector(ctx, productRepo)

	// 5. Sunucu portunu env'den oku; varsayılan 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Kap-App backend %s adresinde dinleniyor...", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: middleware.CORSMiddleware(mux),
	}

	// 6. Graceful shutdown: context iptal edilince sunucuyu durdur.
	go func() {
		<-ctx.Done()
		log.Println("main: kapatma sinyali alındı, sunucu durduruluyor...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("main: graceful shutdown hatası: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("main.ListenAndServe: sunucu başlatılamadı: %v", err)
	}
	log.Println("main: sunucu başarıyla durduruldu.")
}


// registerRoutes, tüm HTTP rotalarını mux'a kaydeder.
// Public ve korunan rotalar burada ayrıştırılır.
// db parametresi, modül repository'leri için bağımlılık enjeksiyonu sağlar.
// productRepo döner — GC goroutine tarafından kullanılır.
func registerRoutes(mux *http.ServeMux, db *config.Database) *product.Repository {
	// ── Public rotalar ────────────────────────────────────────────
	mux.HandleFunc("GET /health", handleHealth)

	// ── Korunan rotalar (AuthMiddleware zinciri) ──────────────────
	mux.Handle("GET /api/v1/me", middleware.AuthMiddleware(http.HandlerFunc(handleMe)))

	// ── Auth modülü ───────────────────────────────────────────────
	authRepo := auth.NewRepository(db.Pool)
	authSvc := auth.NewService(authRepo)
	auth.NewHandler(authSvc).RegisterRoutes(mux)

	// ── Tenant modülü ─────────────────────────────────────────────
	tenantRepo := tenant.NewRepository(db.Pool)
	tenantSvc := tenant.NewService(tenantRepo)
	tenant.NewHandler(tenantSvc).RegisterRoutes(mux)

	// ── Product modülü ────────────────────────────────────────────
	// POST   /api/v1/products               → ürün ekle
	// GET    /api/v1/products?tenant_id=... → ürün listele
	// PATCH  /api/v1/products/{id}/status   → durum güncelle
	// DELETE /api/v1/products/{id}          → ürün sil
	productRepo := product.NewRepository(db.Pool)
	productSvc := product.NewService(productRepo)
	product.NewHandler(productSvc).RegisterRoutes(mux)

	return productRepo
}

// handleHealth, yük dengeleyiciler ve monitoring için basit sağlık kontrolü döner.
// Auth gerektirmez.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleMe, AuthMiddleware'in context'e gömdüğü userID'yi okuyarak döner.
// Bu rota yalnızca geçerli JWT ile erişilebilir; test ve geliştirme amaçlıdır.
func handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		// AuthMiddleware bypass edilmeden buraya gelinemez — defensive check.
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkisiz"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"user_id": userID})
}

// writeJSON, w'ye Content-Type: application/json başlığını ve JSON gövdesini yazar.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("main.writeJSON: encode hatası: %v", err)
	}
}
