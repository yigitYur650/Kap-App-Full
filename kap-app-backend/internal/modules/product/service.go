package product

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

// ── Model ─────────────────────────────────────────────────────────────────────

// Product, ürün kaydını temsil eder.
// Nullable alanlar pointer tiplidir; JSON'da omitempty ile gizlenir.
type Product struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	Name           string     `json:"name"`
	Price          *float64   `json:"price,omitempty"`
	MarketName     *string    `json:"market_name,omitempty"`
	Category       *string    `json:"category,omitempty"`
	Unit           *string    `json:"unit,omitempty"`
	Quantity       int        `json:"quantity"`
	Status         string     `json:"status"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
	AddedBy        string     `json:"added_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ── Hata sabitleri ────────────────────────────────────────────────────────────

var (
	ErrProductNotFound = errors.New("ürün bulunamadı")
	ErrForbidden       = errors.New("bu işlem için yetkiniz yok")
)

// validStatuses, kabul edilen durum değerleri kümesidir.
var validStatuses = map[string]struct{}{
	"var":    {},
	"azaldı": {},
	"yok":    {},
}

// ValidationError, iş kuralı ihlalini taşır; mesaj doğrudan frontend'e gönderilebilir.
type ValidationError string

func (e ValidationError) Error() string { return string(e) }

func errValidation(msg string) ValidationError { return ValidationError(msg) }

// ── Service ───────────────────────────────────────────────────────────────────

// Service, product modülünün iş mantığı katmanıdır.
type Service struct {
	repo *Repository
}

// NewService, Service örneği oluşturur.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// AddProductInput, handler'dan servise taşınan ham istek verisidir.
type AddProductInput struct {
	TenantID       string
	Name           string
	Price          *float64
	MarketName     *string
	Category       *string
	Unit           *string
	Quantity       *int    // nil → 1
	ExpirationDate *string // YYYY-MM-DD
}

// AddProduct, iş kurallarını doğrulayarak yeni ürün oluşturur.
func (s *Service) AddProduct(ctx context.Context, userID string, in AddProductInput) (*Product, error) {
	// 1. Yetki kontrolü.
	if err := s.checkMembership(ctx, in.TenantID, userID, "AddProduct"); err != nil {
		return nil, err
	}

	// 2. Doğrulama.
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errValidation("ürün adı boş olamaz")
	}
	if utf8.RuneCountInString(name) > 200 {
		return nil, errValidation("ürün adı en fazla 200 karakter olabilir")
	}

	qty := 1
	if in.Quantity != nil {
		if *in.Quantity < 1 {
			return nil, errValidation("miktar en az 1 olmalıdır")
		}
		qty = *in.Quantity
	}

	if in.Price != nil && *in.Price < 0 {
		return nil, errValidation("fiyat negatif olamaz")
	}

	// 3. Tarih ayrıştırma.
	var expDate *time.Time
	if in.ExpirationDate != nil {
		t, err := time.Parse("2006-01-02", *in.ExpirationDate)
		if err != nil {
			return nil, errValidation("son kullanma tarihi YYYY-MM-DD formatında olmalıdır")
		}
		expDate = &t
	}

	// 4. Kayıt oluştur.
	p, err := s.repo.Create(ctx, CreateInput{
		TenantID:       in.TenantID,
		AddedBy:        userID,
		Name:           name,
		Price:          in.Price,
		MarketName:     in.MarketName,
		Category:       in.Category,
		Unit:           in.Unit,
		Quantity:       qty,
		ExpirationDate: expDate,
	})
	if err != nil {
		return nil, fmt.Errorf("product.Service.AddProduct: Create: %w", err)
	}

	log.Printf("product.Service.AddProduct: ürün eklendi product_id=%s tenant_id=%s user=%s", p.ID, in.TenantID, userID)
	return p, nil
}

// ListProducts, kullanıcının üye olduğu evin ürünlerini döner.
func (s *Service) ListProducts(ctx context.Context, userID, tenantID string) ([]Product, error) {
	if err := s.checkMembership(ctx, tenantID, userID, "ListProducts"); err != nil {
		return nil, err
	}

	list, err := s.repo.ListByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("product.Service.ListProducts: %w", err)
	}

	if list == nil {
		list = []Product{} // null yerine boş dizi
	}
	return list, nil
}

// UpdateStatus, ürünün durumunu değiştirir.
func (s *Service) UpdateStatus(ctx context.Context, userID, productID, status string) (*Product, error) {
	// 1. Statü doğrulama (tenant kontrolünden önce — ucuz işlem).
	if _, ok := validStatuses[status]; !ok {
		return nil, errValidation(`durum "var", "azaldı" veya "yok" olmalıdır`)
	}

	// 2. Ürünün var olduğunu ve hangi eve ait olduğunu öğren.
	existing, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product.Service.UpdateStatus: FindByID: %w", err)
	}
	if existing == nil {
		return nil, ErrProductNotFound
	}

	// 3. Yetki kontrolü.
	if err := s.checkMembership(ctx, existing.TenantID, userID, "UpdateStatus"); err != nil {
		return nil, err
	}

	// 4. Güncelle.
	updated, err := s.repo.UpdateStatus(ctx, productID, status)
	if err != nil {
		return nil, fmt.Errorf("product.Service.UpdateStatus: %w", err)
	}

	log.Printf("product.Service.UpdateStatus: product_id=%s status=%s user=%s", productID, status, userID)
	return updated, nil
}

// DeleteProduct, ürünü siler.
func (s *Service) DeleteProduct(ctx context.Context, userID, productID string) error {
	// 1. Ürünün hangi eve ait olduğunu öğren.
	existing, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("product.Service.DeleteProduct: FindByID: %w", err)
	}
	if existing == nil {
		return ErrProductNotFound
	}

	// 2. Yetki kontrolü.
	if err := s.checkMembership(ctx, existing.TenantID, userID, "DeleteProduct"); err != nil {
		return err
	}

	// 3. Sil.
	if err := s.repo.Delete(ctx, productID); err != nil {
		return fmt.Errorf("product.Service.DeleteProduct: %w", err)
	}

	log.Printf("product.Service.DeleteProduct: product_id=%s user=%s", productID, userID)
	return nil
}

// checkMembership, yetki kontrol yardımcısıdır.
// Üye değilse ErrForbidden'ı fmt.Errorf ile sarmalar.
func (s *Service) checkMembership(ctx context.Context, tenantID, userID, caller string) error {
	if tenantID == "" {
		return errValidation("tenant_id boş olamaz")
	}

	ok, err := s.repo.IsTenantMember(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("product.Service.%s: IsTenantMember: %w", caller, err)
	}
	if !ok {
		return fmt.Errorf("product.Service.%s: %w", caller, ErrForbidden)
	}
	return nil
}
