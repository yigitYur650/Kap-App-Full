package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"assignment-backend/internal/middleware"
	"assignment-backend/internal/modules/auth"
	"assignment-backend/internal/modules/product"
	"assignment-backend/internal/modules/tenant"
)

type testSuite struct {
	server       *httptest.Server
	supabaseURL  string
	supabaseMock *auth.SupabaseMock
	tenantRepo   *mockTenantRepo
	productRepo  *mockProductRepo
	aliToken     string
	aliUserID    string
	ayseToken    string
	ayseUserID   string
	ayseSlugID   string
	tenantID     string
	productIDs   []string
}

func setupSuite(t *testing.T) *testSuite {
	t.Helper()
	t.Setenv("SUPABASE_JWT_SECRET", "integration-test-secret-key-for-hs256-at-least-32-chars!")
	middleware.InitJWTSecret()
	supabaseMock := auth.NewSupabaseMock()
	supabaseMock.Start()
	t.Cleanup(supabaseMock.Close)
	t.Setenv("SUPABASE_URL", supabaseMock.URL())
	t.Setenv("SUPABASE_ANON_KEY", "test-anon-key")
	t.Setenv("SUPABASE_SERVICE_ROLE_KEY", "test-service-role-key")
	tenantRepo := &mockTenantRepo{
		tenants:  make(map[string]*tenant.Tenant),
		members:  make(map[string]map[string]bool),
		profiles: make(map[string]*tenant.ProfileRef),
	}
	productRepo := &mockProductRepo{
		products: make(map[string]*product.Product),
		members:  make(map[string]map[string]bool),
	}
	authSvc := auth.NewService(auth.NewRepository(nil))
	tenantSvc := tenant.NewService(tenantRepo)
	productSvc := product.NewService(productRepo)
	authHandler := auth.NewHandler(authSvc)
	tenantHandler := tenant.NewHandler(tenantSvc)
	productHandler := product.NewHandler(productSvc)
	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)
	tenantHandler.RegisterRoutes(mux)
	productHandler.RegisterRoutes(mux)
	server := httptest.NewServer(middleware.CORSMiddleware(mux))
	t.Cleanup(server.Close)
	return &testSuite{
		server: server, supabaseURL: supabaseMock.URL(),
		supabaseMock: supabaseMock,
		tenantRepo:   tenantRepo, productRepo: productRepo,
	}
}

func (s *testSuite) url(path string) string { return s.server.URL + path }

func (s *testSuite) doRequest(t *testing.T, method, path string, body any, token string) *http.Response {
	t.Helper()
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("JSON marshal: %v", err)
		}
	}
	req, err := http.NewRequest(method, s.url(path), bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("HTTP istek: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP istek basarisiz: %v", err)
	}
	return resp
}

func (s *testSuite) post(t *testing.T, path string, body any, token string) *http.Response {
	return s.doRequest(t, http.MethodPost, path, body, token)
}
func (s *testSuite) get(t *testing.T, path string, token string) *http.Response {
	return s.doRequest(t, http.MethodGet, path, nil, token)
}
func (s *testSuite) patch(t *testing.T, path string, body any, token string) *http.Response {
	return s.doRequest(t, http.MethodPatch, path, body, token)
}
func (s *testSuite) deleteReq(t *testing.T, path string, token string) *http.Response {
	return s.doRequest(t, http.MethodDelete, path, nil, token)
}

func (s *testSuite) decode(t *testing.T, resp *http.Response, v any) string {
	t.Helper()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Yanit (status=%d): %s", resp.StatusCode, string(body))
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}
	return string(body)
}

func (s *testSuite) registerUser(t *testing.T, email, password, name string) (token, userID string) {
	t.Helper()
	resp := s.post(t, "/api/v1/auth/register", map[string]string{
		"email": email, "password": password, "name": name,
	}, "")
	var reg map[string]any
	s.decode(t, resp, &reg)
	return reg["access_token"].(string), reg["user_id"].(string)
}

func (s *testSuite) createTenant(t *testing.T, name, color, token string) string {
	t.Helper()
	resp := s.post(t, "/api/v1/tenants", map[string]string{"name": name, "theme_color": color}, token)
	var ten map[string]any
	s.decode(t, resp, &ten)
	return ten["id"].(string)
}

func (s *testSuite) addProduct(t *testing.T, name string, price any, quantity int) {
	t.Helper()
	body := map[string]any{"tenant_id": s.tenantID, "name": name, "quantity": quantity}
	if price != nil {
		body["price"] = price
	}
	resp := s.post(t, "/api/v1/products", body, s.aliToken)
	var p map[string]any
	s.decode(t, resp, &p)
	s.productIDs = append(s.productIDs, p["id"].(string))
}

func (s *testSuite) ensureProductMembership() {
	if s.productRepo.members[s.tenantID] == nil {
		s.productRepo.members[s.tenantID] = make(map[string]bool)
	}
	s.productRepo.members[s.tenantID][s.aliUserID] = true
}

func (s *testSuite) addProfileToMock(displayName, slugID, userID string) {
	s.tenantRepo.profiles[slugID] = &tenant.ProfileRef{
		UserID: userID, DisplayName: displayName,
	}
}

type mockTenantRepo struct {
	tenants  map[string]*tenant.Tenant
	members  map[string]map[string]bool
	profiles map[string]*tenant.ProfileRef
}

func (r *mockTenantRepo) CreateTenant(ctx context.Context, ownerID, name, themeColor string) (*tenant.Tenant, error) {
	id := fmt.Sprintf("tenant-%d", len(r.tenants)+1)
	t := &tenant.Tenant{ID: id, Name: name, ThemeColor: themeColor, OwnerID: ownerID, CreatedAt: time.Now()}
	r.tenants[id] = t
	if r.members[id] == nil {
		r.members[id] = make(map[string]bool)
	}
	r.members[id][ownerID] = true
	return t, nil
}

func (r *mockTenantRepo) ListByUserID(ctx context.Context, userID string) ([]tenant.Tenant, error) {
	var list []tenant.Tenant
	for _, t := range r.tenants {
		if r.members[t.ID] != nil && r.members[t.ID][userID] {
			list = append(list, *t)
		}
	}
	return list, nil
}

func (r *mockTenantRepo) FindUserBySlugID(ctx context.Context, slugID string) (*tenant.ProfileRef, error) {
	p, ok := r.profiles[slugID]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (r *mockTenantRepo) AddMember(ctx context.Context, tenantID, userID string) error {
	if r.members[tenantID] == nil {
		r.members[tenantID] = make(map[string]bool)
	}
	if r.members[tenantID][userID] {
		return tenant.ErrAlreadyMember
	}
	r.members[tenantID][userID] = true
	return nil
}

func (r *mockTenantRepo) IsTenantMember(ctx context.Context, tenantID, userID string) (bool, error) {
	if r.members[tenantID] == nil {
		return false, nil
	}
	return r.members[tenantID][userID], nil
}

type mockProductRepo struct {
	products map[string]*product.Product
	members  map[string]map[string]bool
	nextID   int
}

func (r *mockProductRepo) Create(ctx context.Context, in product.CreateInput) (*product.Product, error) {
	r.nextID++
	id := fmt.Sprintf("product-%d", r.nextID)
	p := &product.Product{
		ID: id, TenantID: in.TenantID, Name: in.Name,
		Price: in.Price, MarketName: in.MarketName,
		Category: in.Category, Unit: in.Unit,
		Quantity: in.Quantity, Status: "yok", AddedBy: in.AddedBy,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	r.products[id] = p
	return p, nil
}

func (r *mockProductRepo) FindByID(ctx context.Context, productID string) (*product.Product, error) {
	p, ok := r.products[productID]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (r *mockProductRepo) ListByTenantID(ctx context.Context, tenantID string) ([]product.Product, error) {
	var list []product.Product
	for _, p := range r.products {
		if p.TenantID == tenantID {
			list = append(list, *p)
		}
	}
	return list, nil
}

func (r *mockProductRepo) UpdateStatus(ctx context.Context, productID, status string) (*product.Product, error) {
	p, ok := r.products[productID]
	if !ok {
		return nil, nil
	}
	p.Status = status
	return p, nil
}

func (r *mockProductRepo) Delete(ctx context.Context, productID string) error {
	delete(r.products, productID)
	return nil
}

func (r *mockProductRepo) IsTenantMember(ctx context.Context, tenantID, userID string) (bool, error) {
	if r.members[tenantID] == nil {
		return false, nil
	}
	return r.members[tenantID][userID], nil
}

func (r *mockProductRepo) DeleteExpiredSoftDeleted(ctx context.Context) (int64, error) {
	return 0, nil
}

// TESTS

func TestFullUserFlow(t *testing.T) {
	s := setupSuite(t)
	token, uid := s.registerUser(t, "ali@example.com", "pass123", "Ali")
	s.aliToken = token
	s.aliUserID = uid
	s.tenantID = s.createTenant(t, "Ali'nin Evi", "#4F46E5", s.aliToken)
	s.ensureProductMembership()
	s.addProduct(t, "Sut", 29.90, 2)
	s.addProduct(t, "Ekmek", nil, 3)
	resp := s.get(t, "/api/v1/products?tenant_id="+s.tenantID, s.aliToken)
	var products []any
	s.decode(t, resp, &products)
	if len(products) != 2 {
		t.Errorf("beklenen: 2 urun, gelen: %d", len(products))
	}
}

func TestUpdateAndDeleteProduct(t *testing.T) {
	s := setupSuite(t)
	token, uid := s.registerUser(t, "ali@example.com", "pass123", "Ali")
	s.aliToken = token
	s.aliUserID = uid
	s.tenantID = s.createTenant(t, "Ali'nin Evi", "#4F46E5", s.aliToken)
	s.ensureProductMembership()
	s.addProduct(t, "Sut", 29.90, 2)
	s.addProduct(t, "Ekmek", nil, 3)
	pid := s.productIDs[0]
	resp := s.patch(t, "/api/v1/products/"+pid+"/status", map[string]string{"status": "var"}, s.aliToken)
	var updated map[string]any
	s.decode(t, resp, &updated)
	if updated["status"] != "var" {
		t.Errorf("beklenen 'var', gelen %s", updated["status"])
	}
	resp = s.deleteReq(t, "/api/v1/products/"+pid, s.aliToken)
	if resp.StatusCode != 204 {
		t.Errorf("beklenen 204, gelen %d", resp.StatusCode)
	}
	resp = s.get(t, "/api/v1/products?tenant_id="+s.tenantID, s.aliToken)
	var products []any
	s.decode(t, resp, &products)
	if len(products) != 1 {
		t.Errorf("beklenen 1 urun, gelen %d", len(products))
	}
}

func TestInviteMemberAndAccessProducts(t *testing.T) {
	s := setupSuite(t)
	token, uid := s.registerUser(t, "ali@example.com", "pass123", "Ali")
	s.aliToken = token
	s.aliUserID = uid
	s.tenantID = s.createTenant(t, "Ali'nin Evi", "#4F46E5", s.aliToken)
	s.ensureProductMembership()
	s.addProduct(t, "Sut", 29.90, 2)

	ayseToken, ayseUserID := s.registerUser(t, "ayse@example.com", "pass456", "Ayse")
	s.ayseToken = ayseToken
	s.ayseUserID = ayseUserID

	// PreRegisterUser ile slug_id'yi ayarla
	s.supabaseMock.PreRegisterUser("ayse@example.com", "pass456", s.ayseUserID, "Ayse", "test-slug-ayse")
	s.ayseSlugID = "test-slug-ayse"
	s.addProfileToMock("Ayse", "test-slug-ayse", s.ayseUserID)

	resp := s.post(t, "/api/v1/tenants/"+s.tenantID+"/members",
		map[string]string{"slug_id": s.ayseSlugID}, s.aliToken)
	var mr map[string]any
	s.decode(t, resp, &mr)
	s.productRepo.members[s.tenantID] = map[string]bool{s.aliUserID: true, s.ayseUserID: true}
	s.tenantRepo.members[s.tenantID] = map[string]bool{s.aliUserID: true, s.ayseUserID: true}

	resp = s.get(t, "/api/v1/products?tenant_id="+s.tenantID, s.ayseToken)
	var products []any
	s.decode(t, resp, &products)
	if len(products) != 1 {
		t.Errorf("beklenen 1 urun, gelen %d", len(products))
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	s := setupSuite(t)
	token, uid := s.registerUser(t, "ali@example.com", "pass123", "Ali")
	s.aliToken = token
	s.aliUserID = uid
	s.tenantID = s.createTenant(t, "Ali'nin Evi", "#4F46E5", s.aliToken)
	s.ensureProductMembership()
	s.addProduct(t, "Sut", 29.90, 2)

	resp := s.get(t, "/api/v1/products?tenant_id="+s.tenantID, "")
	if resp.StatusCode != 401 {
		t.Errorf("beklenen 401, gelen %d", resp.StatusCode)
	}
	resp = s.get(t, "/api/v1/tenants", "invalid-token")
	if resp.StatusCode != 401 {
		t.Errorf("beklenen 401, gelen %d", resp.StatusCode)
	}
	resp = s.patch(t, "/api/v1/products/nonexistent/status", map[string]string{"status": "var"}, s.aliToken)
	if resp.StatusCode != 404 {
		t.Errorf("beklenen 404, gelen %d", resp.StatusCode)
	}
}

func TestValidationErrors(t *testing.T) {
	s := setupSuite(t)
	token, uid := s.registerUser(t, "ali@example.com", "pass123", "Ali")
	s.aliToken = token
	s.aliUserID = uid
	s.tenantID = s.createTenant(t, "Ali'nin Evi", "#4F46E5", s.aliToken)
	s.ensureProductMembership()

	resp := s.post(t, "/api/v1/products", map[string]any{"tenant_id": s.tenantID, "name": ""}, s.aliToken)
	if resp.StatusCode != 400 {
		t.Errorf("beklenen 400, gelen %d", resp.StatusCode)
	}
	resp = s.post(t, "/api/v1/products", map[string]any{"tenant_id": s.tenantID, "name": "Test", "price": -10}, s.aliToken)
	if resp.StatusCode != 400 {
		t.Errorf("beklenen 400, gelen %d", resp.StatusCode)
	}
	resp = s.post(t, "/api/v1/tenants", map[string]string{"name": "", "theme_color": "#4F46E5"}, s.aliToken)
	if resp.StatusCode != 400 {
		t.Errorf("beklenen 400, gelen %d", resp.StatusCode)
	}
	resp = s.post(t, "/api/v1/tenants", map[string]string{"name": "Ev", "theme_color": "#XYZ"}, s.aliToken)
	if resp.StatusCode != 400 {
		t.Errorf("beklenen 400, gelen %d", resp.StatusCode)
	}
	resp = s.post(t, "/api/v1/auth/login", map[string]string{"email": "ali@example.com", "password": "wrong"}, "")
	var errResp map[string]string
	s.decode(t, resp, &errResp)
	if resp.StatusCode != 400 {
		t.Errorf("beklenen 400, gelen %d: %s", resp.StatusCode, errResp["error"])
	}
}
