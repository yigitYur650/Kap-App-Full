package tenant

import (
	"context"
	"errors"
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────────
// Tenant Service & Handler Unit Tests
//
// Bu testler:
//   - Service katmanının iş mantığını mock repository ile test eder
//   - Handler katmanının HTTP istek/yanıt döngüsünü test eder
//
// Gerçek veritabanı bağlantısı gerektirmez.
// ──────────────────────────────────────────────────────────────────────────────

// ── Mock Repository ───────────────────────────────────────────────────────────

type mockRepo struct {
	createTenantFunc    func(ctx context.Context, ownerID, name, themeColor string) (*Tenant, error)
	listByUserIDFunc    func(ctx context.Context, userID string) ([]Tenant, error)
	findUserBySlugFunc  func(ctx context.Context, slugID string) (*ProfileRef, error)
	addMemberFunc       func(ctx context.Context, tenantID, userID string) error
	isTenantMemberFunc  func(ctx context.Context, tenantID, userID string) (bool, error)
}

func (m *mockRepo) CreateTenant(ctx context.Context, ownerID, name, themeColor string) (*Tenant, error) {
	return m.createTenantFunc(ctx, ownerID, name, themeColor)
}

func (m *mockRepo) ListByUserID(ctx context.Context, userID string) ([]Tenant, error) {
	return m.listByUserIDFunc(ctx, userID)
}

func (m *mockRepo) FindUserBySlugID(ctx context.Context, slugID string) (*ProfileRef, error) {
	return m.findUserBySlugFunc(ctx, slugID)
}

func (m *mockRepo) AddMember(ctx context.Context, tenantID, userID string) error {
	return m.addMemberFunc(ctx, tenantID, userID)
}

func (m *mockRepo) IsTenantMember(ctx context.Context, tenantID, userID string) (bool, error) {
	return m.isTenantMemberFunc(ctx, tenantID, userID)
}

// ── Dummy veriler ─────────────────────────────────────────────────────────────

const (
	testUserID   = "user-owner-001"
	testTenantID = "tenant-abc-123"
	testSlugID   = "abc123xy"
)

// ── 1. CreateTenant Service Testleri ─────────────────────────────────────────

func TestCreateTenant_Success(t *testing.T) {
	repo := &mockRepo{
		createTenantFunc: func(_ context.Context, ownerID, name, themeColor string) (*Tenant, error) {
			return &Tenant{
				ID: testTenantID, Name: name, ThemeColor: themeColor,
				OwnerID: ownerID,
			}, nil
		},
	}
	svc := NewService(repo)

	tenant, err := svc.CreateTenant(context.Background(), testUserID, "Aile Evi", "#FF5733")
	if err != nil {
		t.Fatalf("CreateTenant başarısız: %v", err)
	}

	if tenant.Name != "Aile Evi" {
		t.Errorf("beklenen: 'Aile Evi', gelen: %s", tenant.Name)
	}
	if tenant.ThemeColor != "#FF5733" {
		t.Errorf("beklenen: '#FF5733', gelen: %s", tenant.ThemeColor)
	}
	if tenant.OwnerID != testUserID {
		t.Errorf("beklenen: %s, gelen: %s", testUserID, tenant.OwnerID)
	}
}

func TestCreateTenant_DefaultThemeColor(t *testing.T) {
	repo := &mockRepo{
		createTenantFunc: func(_ context.Context, ownerID, name, themeColor string) (*Tenant, error) {
			if themeColor != "#4F46E5" {
				t.Errorf("beklenen varsayılan renk: #4F46E5, gelen: %s", themeColor)
			}
			return &Tenant{ID: testTenantID, Name: name, ThemeColor: themeColor, OwnerID: ownerID}, nil
		},
	}
	svc := NewService(repo)

	tenant, err := svc.CreateTenant(context.Background(), testUserID, "Aile Evi", "")
	if err != nil {
		t.Fatalf("CreateTenant başarısız: %v", err)
	}
	if tenant.ThemeColor != "#4F46E5" {
		t.Errorf("beklenen: #4F46E5, gelen: %s", tenant.ThemeColor)
	}
}

func TestCreateTenant_EmptyName(t *testing.T) {
	svc := NewService(nil)

	_, err := svc.CreateTenant(context.Background(), testUserID, "", "#FF5733")
	if err == nil {
		t.Fatal("boş isim hata vermeli")
	}
	if !errors.Is(err, ErrValidation("ev adı boş olamaz")) {
		t.Errorf("beklenen: ValidationError('ev adı boş olamaz'), gelen: %v", err)
	}
}

func TestCreateTenant_NameTooLong(t *testing.T) {
	svc := NewService(nil)

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	_, err := svc.CreateTenant(context.Background(), testUserID, longName, "#FF5733")
	if err == nil {
		t.Fatal("çok uzun isim hata vermeli")
	}
	if !errors.Is(err, ErrValidation("ev adı en fazla 100 karakter olabilir")) {
		t.Errorf("beklenen: ValidationError, gelen: %v", err)
	}
}

func TestCreateTenant_InvalidHexColor_Short(t *testing.T) {
	svc := NewService(nil)

	_, err := svc.CreateTenant(context.Background(), testUserID, "Evim", "#XYZ")
	if err == nil {
		t.Fatal("geçersiz hex rengi hata vermeli")
	}
}

func TestCreateTenant_InvalidHexColor_Long(t *testing.T) {
	svc := NewService(nil)

	_, err := svc.CreateTenant(context.Background(), testUserID, "Evim", "#GGGGGG")
	if err == nil {
		t.Fatal("geçersiz hex rengi hata vermeli")
	}
}

func TestCreateTenant_ValidHexVariants(t *testing.T) {
	validColors := []string{
		"#FFF", "#000", "#abc", "#ABC",
		"#FFFFFF", "#000000", "#aabbcc", "#AABBCC",
		"#4F46E5", "#FF5733",
	}

	for _, color := range validColors {
		t.Run(color, func(t *testing.T) {
			called := false
			repo := &mockRepo{
				createTenantFunc: func(_ context.Context, _, _, _ string) (*Tenant, error) {
					called = true
					return &Tenant{Name: "test"}, nil
				},
			}
			svc := NewService(repo)

			_, err := svc.CreateTenant(context.Background(), testUserID, "Ev", color)
			if err != nil {
				t.Errorf("geçerli renk %s hata vermemeli: %v", color, err)
			}
			if !called {
				t.Error("repo.CreateTenant çağrılmadı")
			}
		})
	}
}

// ── 2. ListTenants Service Testleri ───────────────────────────────────────────

func TestListTenants_Success(t *testing.T) {
	repo := &mockRepo{
		listByUserIDFunc: func(_ context.Context, userID string) ([]Tenant, error) {
			return []Tenant{
				{ID: "t1", Name: "Ev 1", OwnerID: userID},
				{ID: "t2", Name: "Ev 2", OwnerID: userID},
			}, nil
		},
	}
	svc := NewService(repo)

	tenants, err := svc.ListTenants(context.Background(), testUserID)
	if err != nil {
		t.Fatalf("ListTenants başarısız: %v", err)
	}
	if len(tenants) != 2 {
		t.Errorf("beklenen: 2 ev, gelen: %d", len(tenants))
	}
}

func TestListTenants_EmptyList(t *testing.T) {
	repo := &mockRepo{
		listByUserIDFunc: func(_ context.Context, _ string) ([]Tenant, error) {
			return nil, nil
		},
	}
	svc := NewService(repo)

	tenants, err := svc.ListTenants(context.Background(), testUserID)
	if err != nil {
		t.Fatalf("ListTenants başarısız: %v", err)
	}
	if tenants == nil {
		t.Fatal("nil yerine boş dilim dönmeli")
	}
	if len(tenants) != 0 {
		t.Errorf("beklenen: 0, gelen: %d", len(tenants))
	}
}

func TestListTenants_NoMembership(t *testing.T) {
	repo := &mockRepo{
		listByUserIDFunc: func(_ context.Context, _ string) ([]Tenant, error) {
			return []Tenant{}, nil
		},
	}
	svc := NewService(repo)

	tenants, err := svc.ListTenants(context.Background(), "non-member-user")
	if err != nil {
		t.Fatalf("ListTenants başarısız: %v", err)
	}
	if len(tenants) != 0 {
		t.Errorf("üye olmayan kullanıcı için 0 ev beklenir, gelen: %d", len(tenants))
	}
}

// ── 3. AddMember Service Testleri ─────────────────────────────────────────────

func TestAddMember_Success(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, tenantID, userID string) (bool, error) {
			return true, nil // requester üyedir
		},
		findUserBySlugFunc: func(_ context.Context, slugID string) (*ProfileRef, error) {
			return &ProfileRef{UserID: "new-user", DisplayName: "Ali"}, nil
		},
		addMemberFunc: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	svc := NewService(repo)

	result, err := svc.AddMember(context.Background(), testTenantID, testUserID, testSlugID)
	if err != nil {
		t.Fatalf("AddMember başarısız: %v", err)
	}
	if result.UserID != "new-user" {
		t.Errorf("beklenen: 'new-user', gelen: %s", result.UserID)
	}
	if result.Name != "Ali" {
		t.Errorf("beklenen: 'Ali', gelen: %s", result.Name)
	}
	if result.TenantID != testTenantID {
		t.Errorf("beklenen: %s, gelen: %s", testTenantID, result.TenantID)
	}
}

func TestAddMember_RequesterNotMember(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil // requester üye DEĞİL
		},
	}
	svc := NewService(repo)

	_, err := svc.AddMember(context.Background(), testTenantID, testUserID, testSlugID)
	if err == nil {
		t.Fatal("üye olmayan kullanıcı üye ekleyememeli")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("beklenen: ErrForbidden, gelen: %v", err)
	}
}

func TestAddMember_EmptySlugID(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.AddMember(context.Background(), testTenantID, testUserID, "")
	if err == nil {
		t.Fatal("boş slug_id hata vermeli")
	}
}

func TestAddMember_SlugIDNotFound(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
		findUserBySlugFunc: func(_ context.Context, _ string) (*ProfileRef, error) {
			return nil, nil // bulunamadı
		},
	}
	svc := NewService(repo)

	_, err := svc.AddMember(context.Background(), testTenantID, testUserID, "nonexistent")
	if err == nil {
		t.Fatal("bulunamayan slug_id hata vermeli")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("beklenen: ErrUserNotFound, gelen: %v", err)
	}
}

func TestAddMember_AlreadyMember(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
		findUserBySlugFunc: func(_ context.Context, _ string) (*ProfileRef, error) {
			return &ProfileRef{UserID: "existing-user", DisplayName: "Zaten Var"}, nil
		},
		addMemberFunc: func(_ context.Context, _, _ string) error {
			return ErrAlreadyMember
		},
	}
	svc := NewService(repo)

	_, err := svc.AddMember(context.Background(), testTenantID, testUserID, testSlugID)
	if err == nil {
		t.Fatal("zaten üye olan kullanıcı hata vermeli")
	}
	if !errors.Is(err, ErrAlreadyMember) {
		t.Errorf("beklenen: ErrAlreadyMember, gelen: %v", err)
	}
}

// ── 4. ValidationError Testleri ──────────────────────────────────────────────

func TestValidationError_Message(t *testing.T) {
	err := ErrValidation("test mesajı")
	if err.Error() != "test mesajı" {
		t.Errorf("beklenen: 'test mesajı', gelen: %s", err.Error())
	}
}

func TestValidationError_IsComparable(t *testing.T) {
	err := ErrValidation("bir hata")
	if !errors.Is(err, ErrValidation("bir hata")) {
		t.Error("ValidationError errors.Is ile karşılaştırılabilir olmalı")
	}
}
