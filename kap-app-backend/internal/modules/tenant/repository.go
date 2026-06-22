package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolationCode = "23505"

// Repository, tenant modulunun veritabani erisim arayuzudur.
type Repository interface {
	CreateTenant(ctx context.Context, ownerID, name, themeColor string) (*Tenant, error)
	ListByUserID(ctx context.Context, userID string) ([]Tenant, error)
	FindUserBySlugID(ctx context.Context, slugID string) (*ProfileRef, error)
	AddMember(ctx context.Context, tenantID, userID string) error
	IsTenantMember(ctx context.Context, tenantID, userID string) (bool, error)
}

// pgRepository, PostgreSQL gercek implementasyonudur.
type pgRepository struct {
	db *pgxpool.Pool
}

// NewRepository, Repository ornegi olusturur.
func NewRepository(db *pgxpool.Pool) *pgRepository {
	return &pgRepository{db: db}
}

// CreateTenat, tek bir transaction icinde iki islem yapar.
func (r *pgRepository) CreateTenant(ctx context.Context, ownerID, name, themeColor string) (*Tenant, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("tenant.Repository.CreateTenant: BeginTx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	const insertTenant = `
		INSERT INTO tenants (name, theme_color, owner_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, name, theme_color, owner_id, created_at`

	var t Tenant
	err = tx.QueryRow(ctx, insertTenant, name, themeColor, ownerID).
		Scan(&t.ID, &t.Name, &t.ThemeColor, &t.OwnerID, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("tenant.Repository.CreateTenant: insert tenant: %w", err)
	}

	const insertMember = `
		INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
		VALUES ($1, $2, 'owner', NOW())`

	if _, err = tx.Exec(ctx, insertMember, t.ID, ownerID); err != nil {
		return nil, fmt.Errorf("tenant.Repository.CreateTenant: insert membership: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("tenant.Repository.CreateTenant: Commit: %w", err)
	}

	return &t, nil
}

// ListByUserID, kullanicinin uye oldugu tum evleri dondurur.
func (r *pgRepository) ListByUserID(ctx context.Context, userID string) ([]Tenant, error) {
	const q = `
		SELECT t.id, t.name, t.theme_color, t.owner_id, t.created_at
		FROM   tenants t
		INNER JOIN tenant_memberships tm ON tm.tenant_id = t.id
		WHERE  tm.user_id = $1
		ORDER BY t.created_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("tenant.Repository.ListByUserID: %w", err)
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.ThemeColor, &t.OwnerID, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("tenant.Repository.ListByUserID: Scan: %w", err)
		}
		tenants = append(tenants, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tenant.Repository.ListByUserID: rows.Err: %w", err)
	}

	return tenants, nil
}

// FindUserBySlugID, profiles tablosunda slug_id'ye gore kullanici arar.
func (r *pgRepository) FindUserBySlugID(ctx context.Context, slugID string) (*ProfileRef, error) {
	const q = `
		SELECT id, display_name
		FROM   profiles
		WHERE  slug_id = $1
		LIMIT  1`

	var p ProfileRef
	err := r.db.QueryRow(ctx, q, slugID).Scan(&p.UserID, &p.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("tenant.Repository.FindUserBySlugID: %w", err)
	}

	return &p, nil
}

// AddMember, kullaniciyi tenant_memberships tablosuna ekler.
func (r *pgRepository) AddMember(ctx context.Context, tenantID, userID string) error {
	const q = `
		INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
		VALUES ($1, $2, 'member', NOW())`

	_, err := r.db.Exec(ctx, q, tenantID, userID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyMember
		}
		return fmt.Errorf("tenant.Repository.AddMember: %w", err)
	}

	return nil
}

// IsTenantMember, userID'nin belirtilen eve uye olup olmadigini kontrol eder.
func (r *pgRepository) IsTenantMember(ctx context.Context, tenantID, userID string) (bool, error) {
	const q = `
		SELECT 1 FROM tenant_memberships
		WHERE  tenant_id = $1 AND user_id = $2
		LIMIT  1`

	var dummy int
	err := r.db.QueryRow(ctx, q, tenantID, userID).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("tenant.Repository.IsTenantMember: %w", err)
	}

	return true, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}
