package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgUniqueViolationCode, PostgreSQL'in unique constraint ihlali için hata kodu.
const pgUniqueViolationCode = "23505"

// Repository, auth modülünün veritabanı erişim katmanıdır.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository, Repository örneği oluşturur.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindByUserID, Supabase UUID'sine göre profiles tablosundan profil döner.
// Kayıt bulunamazsa (nil, nil) döner; "bulunamadı" ile hata bu şekilde ayrışır.
func (r *Repository) FindByUserID(ctx context.Context, userID string) (*UserProfile, error) {
	const q = `
		SELECT id, slug_id, display_name
		FROM   profiles
		WHERE  id = $1
		LIMIT  1`

	var p UserProfile
	err := r.db.QueryRow(ctx, q, userID).Scan(&p.UserID, &p.SlugID, &p.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // kayıt yok — hata değil
	}
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.FindByUserID: %w", err)
	}

	return &p, nil
}

// Create, profiles tablosuna yeni satır ekler ve oluşturulan profili döner.
// slug_id sütununda UNIQUE kısıtı olduğu varsayılır;
// çakışma olursa pgconn.PgError{Code:"23505"} döner, service katmanı yakalar.
func (r *Repository) Create(ctx context.Context, userID, slugID, displayName string) (*UserProfile, error) {
	const q = `
		INSERT INTO profiles (id, slug_id, display_name, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, slug_id, display_name`

	var p UserProfile
	p.IsNew = true

	err := r.db.QueryRow(ctx, q, userID, slugID, displayName).
		Scan(&p.UserID, &p.SlugID, &p.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.Create: %w", err)
	}

	return &p, nil
}

// isUniqueViolation, hatanın PostgreSQL unique constraint ihlali olup
// olmadığını kontrol eder.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}
