package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository, product modulunun veritabani erisim arayuzudur.
type Repository interface {
	Create(ctx context.Context, in CreateInput) (*Product, error)
	FindByID(ctx context.Context, productID string) (*Product, error)
	ListByTenantID(ctx context.Context, tenantID string) ([]Product, error)
	UpdateStatus(ctx context.Context, productID, status string) (*Product, error)
	Delete(ctx context.Context, productID string) error
	IsTenantMember(ctx context.Context, tenantID, userID string) (bool, error)
	DeleteExpiredSoftDeleted(ctx context.Context) (int64, error)
}

// pgRepository, PostgreSQL gercek implementasyonudur.
type pgRepository struct {
	db *pgxpool.Pool
}

// NewRepository, Repository ornegi olusturur.
func NewRepository(db *pgxpool.Pool) *pgRepository {
	return &pgRepository{db: db}
}

// CreateInput, Create metodu icin girdi.
type CreateInput struct {
	TenantID       string
	AddedBy        string
	Name           string
	Price          *float64
	MarketName     *string
	Category       *string
	Unit           *string
	Quantity       int
	ExpirationDate *time.Time
}

func (r *pgRepository) Create(ctx context.Context, in CreateInput) (*Product, error) {
	const q = `
		INSERT INTO products (tenant_id, name, price, market_name, category, unit, quantity, status, added_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'yok', $8, NOW(), NOW())
		RETURNING id, tenant_id, name, price, market_name, category, unit, quantity, status, expiration_date, added_by, created_at, updated_at`

	var p Product
	err := r.db.QueryRow(ctx, q,
		in.TenantID, in.Name, in.Price, in.MarketName, in.Category,
		in.Unit, in.Quantity, in.AddedBy,
	).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Price, &p.MarketName,
		&p.Category, &p.Unit, &p.Quantity, &p.Status, &p.ExpirationDate,
		&p.AddedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("product.Repository.Create: %w", err)
	}
	return &p, nil
}

func (r *pgRepository) FindByID(ctx context.Context, productID string) (*Product, error) {
	const q = `
		SELECT id, tenant_id, name, price, market_name, category, unit, quantity, status, expiration_date, added_by, created_at, updated_at
		FROM   products
		WHERE  id = $1 AND deleted_at IS NULL`

	var p Product
	err := r.db.QueryRow(ctx, q, productID).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Price, &p.MarketName,
		&p.Category, &p.Unit, &p.Quantity, &p.Status, &p.ExpirationDate,
		&p.AddedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("product.Repository.FindByID: %w", err)
	}
	return &p, nil
}

func (r *pgRepository) ListByTenantID(ctx context.Context, tenantID string) ([]Product, error) {
	const q = `
		SELECT id, tenant_id, name, price, market_name, category, unit, quantity, status, expiration_date, added_by, created_at, updated_at
		FROM   products
		WHERE  tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("product.Repository.ListByTenantID: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Name, &p.Price, &p.MarketName,
			&p.Category, &p.Unit, &p.Quantity, &p.Status, &p.ExpirationDate,
			&p.AddedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("product.Repository.ListByTenantID: Scan: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("product.Repository.ListByTenantID: rows.Err: %w", err)
	}

	if products == nil {
		products = []Product{}
	}
	return products, nil
}

func (r *pgRepository) UpdateStatus(ctx context.Context, productID, status string) (*Product, error) {
	const q = `
		UPDATE products
		SET    status = $1, updated_at = NOW()
		WHERE  id = $2 AND deleted_at IS NULL
		RETURNING id, tenant_id, name, price, market_name, category, unit, quantity, status, expiration_date, added_by, created_at, updated_at`

	var p Product
	err := r.db.QueryRow(ctx, q, status, productID).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Price, &p.MarketName,
		&p.Category, &p.Unit, &p.Quantity, &p.Status, &p.ExpirationDate,
		&p.AddedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("product.Repository.UpdateStatus: %w", err)
	}
	return &p, nil
}

func (r *pgRepository) Delete(ctx context.Context, productID string) error {
	const q = `UPDATE products SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, productID)
	if err != nil {
		return fmt.Errorf("product.Repository.Delete: %w", err)
	}
	return nil
}

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
		return false, fmt.Errorf("product.Repository.IsTenantMember: %w", err)
	}
	return true, nil
}

func (r *pgRepository) DeleteExpiredSoftDeleted(ctx context.Context) (int64, error) {
	const q = `DELETE FROM products WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '30 days'`
	res, err := r.db.Exec(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("product.Repository.DeleteExpiredSoftDeleted: %w", err)
	}
	return res.RowsAffected(), nil
}
