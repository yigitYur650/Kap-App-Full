package tenant

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// ── Modeller ─────────────────────────────────────────────────────────────────

// Tenant, bir ev/aile grubunu temsil eder.
type Tenant struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ThemeColor string    `json:"theme_color"`
	OwnerID    string    `json:"owner_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// ProfileRef, davet akışında slug_id ile çözümlenen kullanıcı referansıdır.
type ProfileRef struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// AddMemberResult, üye ekleme işleminin sonucunu taşır.
type AddMemberResult struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	SlugID   string `json:"slug_id"`
	Name     string `json:"display_name"`
}

// ── Hata sabitleri ────────────────────────────────────────────────────────────

// ErrAlreadyMember, kullanıcının o eve zaten üye olduğunu belirtir.
var ErrAlreadyMember = errors.New("kullanıcı bu eve zaten üye")

// ── Service ───────────────────────────────────────────────────────────────────

// Service, tenant modülünün iş mantığı katmanıdır.
type Service struct {
	repo *Repository
}

// NewService, Service örneği oluşturur.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// reHexColor, geçerli CSS hex renk (#RGB veya #RRGGBB) kontrolü için.
var reHexColor = regexp.MustCompile(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`)

// CreateTenant, iş kurallarını doğrulayıp yeni bir ev oluşturur.
func (s *Service) CreateTenant(ctx context.Context, ownerID, name, themeColor string) (*Tenant, error) {
	// Doğrulama.
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("tenant.Service.CreateTenant: %w", ErrValidation("ev adı boş olamaz"))
	}
	if utf8.RuneCountInString(name) > 100 {
		return nil, fmt.Errorf("tenant.Service.CreateTenant: %w", ErrValidation("ev adı en fazla 100 karakter olabilir"))
	}

	if themeColor == "" {
		themeColor = "#4F46E5" // varsayılan: indigo
	}
	if !reHexColor.MatchString(themeColor) {
		return nil, fmt.Errorf("tenant.Service.CreateTenant: %w", ErrValidation("tema rengi geçersiz hex renk kodu (#RGB veya #RRGGBB)"))
	}

	t, err := s.repo.CreateTenant(ctx, ownerID, name, themeColor)
	if err != nil {
		return nil, fmt.Errorf("tenant.Service.CreateTenant: %w", err)
	}

	log.Printf("tenant.Service.CreateTenant: ev oluşturuldu tenant_id=%s owner=%s name=%q", t.ID, ownerID, name)
	return t, nil
}

// ListTenants, kullanıcının üye olduğu tüm evleri döner.
func (s *Service) ListTenants(ctx context.Context, userID string) ([]Tenant, error) {
	tenants, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("tenant.Service.ListTenants: %w", err)
	}

	// nil yerine boş dilim dön — frontend'de null yerine [] gelir.
	if tenants == nil {
		tenants = []Tenant{}
	}

	return tenants, nil
}

// AddMember, slugID ile kullanıcı arar ve belirtilen eve ekler.
// İstekte bulunanın o evin üyesi olması zorunludur (yetki kontrolü).
func (s *Service) AddMember(ctx context.Context, tenantID, requesterID, slugID string) (*AddMemberResult, error) {
	// 1. Davet edeni doğrula: o evin üyesi mi?
	isMember, err := s.repo.IsTenantMember(ctx, tenantID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("tenant.Service.AddMember: IsTenantMember: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("tenant.Service.AddMember: %w", ErrForbidden)
	}

	// 2. slug_id'ye sahip kullanıcıyı bul.
	slugID = strings.TrimSpace(slugID)
	if slugID == "" {
		return nil, fmt.Errorf("tenant.Service.AddMember: %w", ErrValidation("slug_id boş olamaz"))
	}

	profile, err := s.repo.FindUserBySlugID(ctx, slugID)
	if err != nil {
		return nil, fmt.Errorf("tenant.Service.AddMember: FindUserBySlugID: %w", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("tenant.Service.AddMember: %w", ErrUserNotFound)
	}

	// 3. Üye ekle.
	if err := s.repo.AddMember(ctx, tenantID, profile.UserID); err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			return nil, fmt.Errorf("tenant.Service.AddMember: %w", ErrAlreadyMember)
		}
		return nil, fmt.Errorf("tenant.Service.AddMember: AddMember: %w", err)
	}

	log.Printf("tenant.Service.AddMember: üye eklendi tenant_id=%s user_id=%s slug=%s", tenantID, profile.UserID, slugID)

	return &AddMemberResult{
		TenantID: tenantID,
		UserID:   profile.UserID,
		SlugID:   slugID,
		Name:     profile.DisplayName,
	}, nil
}

// ── Özel hata tipleri ────────────────────────────────────────────────────────

// ValidationError, iş kuralı ihlallerini taşır; mesaj doğrudan frontend'e gönderilebilir.
type ValidationError string

func (e ValidationError) Error() string { return string(e) }

// ErrValidation, ValidationError üretir.
func ErrValidation(msg string) ValidationError { return ValidationError(msg) }

// ErrForbidden, kullanıcının bu işlem için yetkisi olmadığını belirtir.
var ErrForbidden = errors.New("bu işlem için yetkiniz yok")

// ErrUserNotFound, slug_id'ye sahip kullanıcı bulunamadığında kullanılır.
var ErrUserNotFound = errors.New("bu kısa ID'ye sahip kullanıcı bulunamadı")
