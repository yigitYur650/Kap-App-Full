package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database, pgx bağlantı havuzunu saran yapıdır.
// Uygulamanın geri kalanı bu tip üzerinden veritabanıyla konuşur.
type Database struct {
	Pool *pgxpool.Pool
}

// Close, bağlantı havuzunu güvenli bir şekilde kapatır.
// defer ile main()'de çağrılmalıdır.
func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		log.Println("config.Database.Close: bağlantı havuzu kapatıldı.")
	}
}

// NewDatabase, çevre değişkenlerini okuyarak bir pgxpool bağlantı havuzu oluşturur.
// Zorunlu değişkenler eksikse veya bağlantı kurulamazsa log.Fatal ile uygulamayı durdurur.
//
// Gerekli çevre değişkenleri:
//   - SUPABASE_DB_URL : PostgreSQL bağlantı dizesi
//     (örn: postgres://user:pass@host:5432/dbname?sslmode=require)
//   - SUPABASE_API_KEY: Supabase servis anahtarı (ileride HTTP istemcisinde kullanılacak)
func NewDatabase() *Database {
	dbURL := mustGetEnv("SUPABASE_DB_URL")
	// API Key şu an doğrudan pgx tarafından kullanılmıyor;
	// Supabase REST/Auth endpoint'leri için ileriki aşamada devreye girecek.
	_ = mustGetEnv("SUPABASE_API_KEY")

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		// Geliştirici için tam hata; uygulama başlatılamaz.
		log.Fatalf("config.NewDatabase: bağlantı dizesi ayrıştırılamadı: %v", err)
	}

	// Havuz parametreleri — ihtiyaca göre çevre değişkenlerine taşınabilir.
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("config.NewDatabase: bağlantı havuzu oluşturulamadı: %v", err)
	}

	// Bağlantının gerçekten çalıştığını doğrula.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatalf("config.NewDatabase: veritabanına ping gönderilemedi: %v", err)
	}

	log.Println("config.NewDatabase: veritabanı bağlantı havuzu başarıyla oluşturuldu.")
	return &Database{Pool: pool}
}

// mustGetEnv, verilen çevre değişkenini okur; boşsa uygulamayı durdurur.
func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("config.mustGetEnv: zorunlu çevre değişkeni eksik: %s", key)
	}
	return val
}

// DSN, bağlantı dizesini gizleyerek (şifre maskelenerek) döner.
// Loglama ve debug amaçlıdır; asla frontend'e dönülmemelidir.
func DSN(dbURL string) string {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return "<ayrıştırılamadı>"
	}
	if cfg.ConnConfig.Password != "" {
		cfg.ConnConfig.Password = "***"
	}
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.ConnConfig.Host,
		cfg.ConnConfig.Port,
		cfg.ConnConfig.Database,
		cfg.ConnConfig.User,
		cfg.ConnConfig.Password,
		"(pgx varsayılanı)",
	)
}
