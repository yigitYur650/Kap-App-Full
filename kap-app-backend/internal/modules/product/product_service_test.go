package product

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Product Service Unit Tests
//
// Bu testler, Service katmanının iş mantığını mock repository ile test eder.
// Gerçek veritabanı bağlantısı gerektirmez.
//
// Test stratejisi:
//   - mockRepository, tüm repository metotlarını taklit eder
//   - Her test için özel mock yapılandırması kullanılır
//   - İş kuralları (validation, yetki, statü geçişleri) mock ile test edilir
// ──────────────────────────────────────────────────────────────────────────────

// ── Mock Repository ───────────────────────────────────────────────────────────

// mockRepo, gerçek Repository'i taklit eden yapı.
type mockRepo struct {
	createFunc         func(ctx context.Context, in CreateInput) (*Product, error)
	findByIDFunc       func(ctx context.Context, productID string) (*Product, error)
	listByTenantIDFunc func(ctx context.Context, tenantID string) ([]Product, error)
	updateStatusFunc   func(ctx context.Context, productID, status string) (*Product, error)
	deleteFunc         func(ctx context.Context, productID string) error
	isTenantMemberFunc func(ctx context.Context, tenantID, userID string) (bool, error)
	deleteExpiredFunc  func(ctx context.Context) (int64, error)
}

func (m *mockRepo) Create(ctx context.Context, in CreateInput) (*Product, error) {
	return m.createFunc(ctx, in)
}

func (m *mockRepo) FindByID(ctx context.Context, productID string) (*Product, error) {
	return m.findByIDFunc(ctx, productID)
}

func (m *mockRepo) ListByTenantID(ctx context.Context, tenantID string) ([]Product, error) {
	return m.listByTenantIDFunc(ctx, tenantID)
}

func (m *mockRepo) UpdateStatus(ctx context.Context, productID, status string) (*Product, error) {
	return m.updateStatusFunc(ctx, productID, status)
}

func (m *mockRepo) Delete(ctx context.Context, productID string) error {
	return m.deleteFunc(ctx, productID)
}

func (m *mockRepo) IsTenantMember(ctx context.Context, tenantID, userID string) (bool, error) {
	return m.isTenantMemberFunc(ctx, tenantID, userID)
}

func (m *mockRepo) DeleteExpiredSoftDeleted(ctx context.Context) (int64, error) {
	return m.deleteExpiredFunc(ctx)
}

// ── Dummy veriler ─────────────────────────────────────────────────────────────

var (
	testTenantID    = "tenant-123"
	testUserID      = "user-456"
	testOtherUserID = "user-789"

	mockTime = time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)

	sampleProduct = &Product{
		ID:         "product-001",
		TenantID:   testTenantID,
		Name:       "Süt",
		Price:      ptrFloat(29.90),
		MarketName: ptrStr("Migros"),
		Category:   ptrStr("dairy"),
		Unit:       ptrStr("litre"),
		Quantity:   2,
		Status:     "yok",
		AddedBy:    testUserID,
		CreatedAt:  mockTime,
		UpdatedAt:  mockTime,
	}
)

func ptrFloat(f float64) *float64 { return &f }
func ptrStr(s string) *string     { return &s }
func ptrInt(i int) *int           { return &i }

// ── 1. AddProduct Testleri ───────────────────────────────────────────────────

func TestAddProduct_Success(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, tenantID, userID string) (bool, error) {
			return true, nil
		},
		createFunc: func(_ context.Context, in CreateInput) (*Product, error) {
			return &Product{
				ID:       "product-new",
				TenantID: in.TenantID,
				Name:     in.Name,
				Price:    in.Price,
				Quantity: in.Quantity,
				Status:   "yok",
				AddedBy:  in.AddedBy,
			}, nil
		},
	}
	svc := NewService(repo)

	p, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: testTenantID,
		Name:     "Süt",
		Price:    ptrFloat(29.90),
		Quantity: ptrInt(2),
	})
	if err != nil {
		t.Fatalf("AddProduct başarısız: %v", err)
	}
	if p.Name != "Süt" {
		t.Errorf("beklenen: 'Süt', gelen: %s", p.Name)
	}
	if p.Quantity != 2 {
		t.Errorf("beklenen: 2, gelen: %d", p.Quantity)
	}
	if p.Status != "yok" {
		t.Errorf("beklenen: 'yok', gelen: %s", p.Status)
	}
}

func TestAddProduct_EmptyName(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	}
	svc := NewService(repo)

	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: testTenantID,
		Name:     "",
	})
	if err == nil {
		t.Fatal("boş isim hata vermeli")
	}
	if err.Error() != "ürün adı boş olamaz" {
		t.Errorf("beklenen: 'ürün adı boş olamaz', gelen: %s", err.Error())
	}
}

func TestAddProduct_NameTooLong(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	}
	svc := NewService(repo)

	// 200 karakterden uzun (201 rune) isim
	longName := ""
	for idx := 0; idx < 201; idx++ {
		longName += "a"
	}

	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: testTenantID,
		Name:     longName,
	})
	if err == nil {
		t.Fatal("çok uzun isim hata vermeli")
	}
	if err.Error() != "ürün adı en fazla 200 karakter olabilir" {
		t.Errorf("beklenen: 'ürün adı en fazla 200 karakter olabilir', gelen: %s", err.Error())
	}
}

func TestAddProduct_NegativeQuantity(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	}
	svc := NewService(repo)

	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: testTenantID,
		Name:     "Elma",
		Quantity: ptrInt(0),
	})
	if err == nil {
		t.Fatal("sıfır miktar hata vermeli")
	}
	if err.Error() != "miktar en az 1 olmalıdır" {
		t.Errorf("beklenen: 'miktar en az 1 olmalıdır', gelen: %s", err.Error())
	}
}

func TestAddProduct_NegativePrice(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	}
	svc := NewService(repo)

	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: testTenantID,
		Name:     "Elma",
		Price:    ptrFloat(-5.0),
	})
	if err == nil {
		t.Fatal("negatif fiyat hata vermeli")
	}
	if err.Error() != "fiyat negatif olamaz" {
		t.Errorf("beklenen: 'fiyat negatif olamaz', gelen: %s", err.Error())
	}
}

func TestAddProduct_EmptyTenantID(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: "",
		Name:     "Elma",
	})
	if err == nil {
		t.Fatal("boş tenant_id hata vermeli")
	}
	if err.Error() != "tenant_id boş olamaz" {
		t.Errorf("beklenen: 'tenant_id boş olamaz', gelen: %s", err.Error())
	}
}

func TestAddProduct_NotMember(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: testTenantID,
		Name:     "Elma",
	})
	if err == nil {
		t.Fatal("üye olmayan kullanıcı hata vermeli")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("beklenen: ErrForbidden, gelen: %v", err)
	}
}

func TestAddProduct_DefaultQuantity(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
		createFunc: func(_ context.Context, in CreateInput) (*Product, error) {
			return &Product{Name: in.Name, Quantity: in.Quantity, Status: "yok"}, nil
		},
	}
	svc := NewService(repo)

	p, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID: testTenantID,
		Name:     "Ekmek",
		// Quantity nil → default 1
	})
	if err != nil {
		t.Fatalf("AddProduct başarısız: %v", err)
	}
	if p.Quantity != 1 {
		t.Errorf("beklenen default quantity: 1, gelen: %d", p.Quantity)
	}
}

func TestAddProduct_WithExpirationDate(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
		createFunc: func(_ context.Context, in CreateInput) (*Product, error) {
			if in.ExpirationDate == nil {
				t.Error("ExpirationDate nil olmamalı")
			}
			return &Product{Name: in.Name, Quantity: in.Quantity, Status: "yok"}, nil
		},
	}
	svc := NewService(repo)

	dateStr := "2026-12-31"
	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID:       testTenantID,
		Name:           "Yoğurt",
		ExpirationDate: &dateStr,
	})
	if err != nil {
		t.Fatalf("AddProduct başarısız: %v", err)
	}
}

func TestAddProduct_InvalidExpirationDateFormat(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	}
	svc := NewService(repo)

	invalidDate := "31-12-2026"
	_, err := svc.AddProduct(context.Background(), testUserID, AddProductInput{
		TenantID:       testTenantID,
		Name:           "Yoğurt",
		ExpirationDate: &invalidDate,
	})
	if err == nil {
		t.Fatal("geçersiz tarih formatı hata vermeli")
	}
	if err.Error() != "son kullanma tarihi YYYY-MM-DD formatında olmalıdır" {
		t.Errorf("beklenen hata mesajı yanlış: %s", err.Error())
	}
}

// ── 2. ListProducts Testleri ─────────────────────────────────────────────────

func TestListProducts_Success(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
		listByTenantIDFunc: func(_ context.Context, tenantID string) ([]Product, error) {
			return []Product{
				{ID: "p1", Name: "Süt", TenantID: tenantID},
				{ID: "p2", Name: "Ekmek", TenantID: tenantID},
			}, nil
		},
	}
	svc := NewService(repo)

	products, err := svc.ListProducts(context.Background(), testUserID, testTenantID)
	if err != nil {
		t.Fatalf("ListProducts başarısız: %v", err)
	}
	if len(products) != 2 {
		t.Errorf("beklenen: 2 ürün, gelen: %d", len(products))
	}
}

func TestListProducts_EmptyList(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
		listByTenantIDFunc: func(_ context.Context, _ string) ([]Product, error) {
			return nil, nil // null döndü
		},
	}
	svc := NewService(repo)

	products, err := svc.ListProducts(context.Background(), testUserID, testTenantID)
	if err != nil {
		t.Fatalf("ListProducts başarısız: %v", err)
	}
	if products == nil {
		t.Fatal("nil yerine boş dizi dönmeli")
	}
	if len(products) != 0 {
		t.Errorf("beklenen: 0 ürün, gelen: %d", len(products))
	}
}

func TestListProducts_NotMember(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.ListProducts(context.Background(), testOtherUserID, testTenantID)
	if err == nil {
		t.Fatal("üye olmayan kullanıcı hata vermeli")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("beklenen: ErrForbidden, gelen: %v", err)
	}
}

func TestListProducts_TenantIsolation(t *testing.T) {
	// Tenant A'nın ürünleri sadece Tenant A üyelerine görünür
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, tenantID, userID string) (bool, error) {
			// user-456 sadece "tenant-123" üyesi
			if userID == "user-456" && tenantID == "tenant-123" {
				return true, nil
			}
			if userID == "user-456" && tenantID == "tenant-999" {
				return false, nil
			}
			return false, nil
		},
		listByTenantIDFunc: func(_ context.Context, tenantID string) ([]Product, error) {
			return []Product{{ID: "p1", Name: "Süt", TenantID: tenantID}}, nil
		},
	}
	svc := NewService(repo)

	// Kendi evi → başarılı
	products, err := svc.ListProducts(context.Background(), "user-456", "tenant-123")
	if err != nil {
		t.Errorf("kendi evi başarısız: %v", err)
	}
	if len(products) != 1 {
		t.Errorf("beklenen: 1 ürün, gelen: %d", len(products))
	}

	// Başkasının evi → hata
	_, err = svc.ListProducts(context.Background(), "user-456", "tenant-999")
	if err == nil {
		t.Error("başkasının evine erişim hata vermeli")
	}
}

// ── 3. UpdateStatus Testleri ──────────────────────────────────────────────────

func TestUpdateStatus_Success(t *testing.T) {
	repo := &mockRepo{
		findByIDFunc: func(_ context.Context, _ string) (*Product, error) {
			return &Product{ID: "p1", TenantID: testTenantID}, nil
		},
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
		updateStatusFunc: func(_ context.Context, _, status string) (*Product, error) {
			return &Product{ID: "p1", Status: status}, nil
		},
	}
	svc := NewService(repo)

	p, err := svc.UpdateStatus(context.Background(), testUserID, "p1", "var")
	if err != nil {
		t.Fatalf("UpdateStatus başarısız: %v", err)
	}
	if p.Status != "var" {
		t.Errorf("beklenen: 'var', gelen: %s", p.Status)
	}
}

func TestUpdateStatus_AllValidTransitions(t *testing.T) {
	validStatuses := []string{"var", "azaldı", "yok"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			repo := &mockRepo{
				findByIDFunc: func(_ context.Context, _ string) (*Product, error) {
					return &Product{ID: "p1", TenantID: testTenantID}, nil
				},
				isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
				updateStatusFunc: func(_ context.Context, _, s string) (*Product, error) {
					return &Product{ID: "p1", Status: s}, nil
				},
			}
			svc := NewService(repo)

			p, err := svc.UpdateStatus(context.Background(), testUserID, "p1", status)
			if err != nil {
				t.Errorf("%s için UpdateStatus başarısız: %v", status, err)
			}
			if p.Status != status {
				t.Errorf("beklenen: %s, gelen: %s", status, p.Status)
			}
		})
	}
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	_, err := svc.UpdateStatus(context.Background(), testUserID, "p1", "geçersiz")
	if err == nil {
		t.Fatal("geçersiz status hata vermeli")
	}
	if err.Error() != `durum "var", "azaldı" veya "yok" olmalıdır` {
		t.Errorf("beklenen hata mesajı yanlış: %s", err.Error())
	}
}

func TestUpdateStatus_ProductNotFound(t *testing.T) {
	repo := &mockRepo{
		findByIDFunc: func(_ context.Context, _ string) (*Product, error) {
			return nil, nil // bulunamadı
		},
	}
	svc := NewService(repo)

	_, err := svc.UpdateStatus(context.Background(), testUserID, "nonexistent", "var")
	if err == nil {
		t.Fatal("var olmayan ürün hata vermeli")
	}
	if !errors.Is(err, ErrProductNotFound) {
		t.Errorf("beklenen: ErrProductNotFound, gelen: %v", err)
	}
}

func TestUpdateStatus_NotMember(t *testing.T) {
	repo := &mockRepo{
		findByIDFunc: func(_ context.Context, _ string) (*Product, error) {
			return &Product{ID: "p1", TenantID: testTenantID}, nil
		},
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.UpdateStatus(context.Background(), testOtherUserID, "p1", "var")
	if err == nil {
		t.Fatal("üye olmayan kullanıcı hata vermeli")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("beklenen: ErrForbidden, gelen: %v", err)
	}
}

// ── 4. DeleteProduct Testleri ─────────────────────────────────────────────────

func TestDeleteProduct_Success(t *testing.T) {
	var deleted bool

	repo := &mockRepo{
		findByIDFunc: func(_ context.Context, _ string) (*Product, error) {
			return &Product{ID: "p1", TenantID: testTenantID}, nil
		},
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
		deleteFunc: func(_ context.Context, _ string) error {
			deleted = true
			return nil
		},
	}
	svc := NewService(repo)

	err := svc.DeleteProduct(context.Background(), testUserID, "p1")
	if err != nil {
		t.Fatalf("DeleteProduct başarısız: %v", err)
	}
	if !deleted {
		t.Error("ürün silinmedi")
	}
}

func TestDeleteProduct_NotFound(t *testing.T) {
	repo := &mockRepo{
		findByIDFunc: func(_ context.Context, _ string) (*Product, error) {
			return nil, nil
		},
	}
	svc := NewService(repo)

	err := svc.DeleteProduct(context.Background(), testUserID, "nonexistent")
	if err == nil {
		t.Fatal("var olmayan ürün hata vermeli")
	}
	if !errors.Is(err, ErrProductNotFound) {
		t.Errorf("beklenen: ErrProductNotFound, gelen: %v", err)
	}
}

func TestDeleteProduct_NotMember(t *testing.T) {
	repo := &mockRepo{
		findByIDFunc: func(_ context.Context, _ string) (*Product, error) {
			return &Product{ID: "p1", TenantID: testTenantID}, nil
		},
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo)

	err := svc.DeleteProduct(context.Background(), testOtherUserID, "p1")
	if err == nil {
		t.Fatal("üye olmayan kullanıcı hata vermeli")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("beklenen: ErrForbidden, gelen: %v", err)
	}
}

// ── 5. checkMembership Testleri ───────────────────────────────────────────────

func TestCheckMembership_Success(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
	}
	svc := NewService(repo)

	err := svc.checkMembership(context.Background(), testTenantID, testUserID, "test")
	if err != nil {
		t.Errorf("beklenen: nil, gelen: %v", err)
	}
}

func TestCheckMembership_NotMember(t *testing.T) {
	repo := &mockRepo{
		isTenantMemberFunc: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo)

	err := svc.checkMembership(context.Background(), testTenantID, testUserID, "test")
	if err == nil {
		t.Fatal("üye olmayan kullanıcı hata vermeli")
	}
}

func TestCheckMembership_EmptyTenantID(t *testing.T) {
	svc := NewService(nil) // repo çağrılmaz, validation önceden yapılır

	err := svc.checkMembership(context.Background(), "", testUserID, "test")
	if err == nil {
		t.Fatal("boş tenant_id hata vermeli")
	}
	if err.Error() != "tenant_id boş olamaz" {
		t.Errorf("beklenen: 'tenant_id boş olamaz', gelen: %s", err.Error())
	}
}
