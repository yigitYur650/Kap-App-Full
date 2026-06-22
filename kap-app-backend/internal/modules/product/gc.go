package product

import (
	"context"
	"log"
	"time"
)

// StartGarbageCollector, 30 günden eski soft-deleted ürünleri temizleyen
// background goroutine'i başlatır.
//
// Çalışma zamanlaması:
//   - Uygulama başlangıcında bir kez (30 sn gecikme — DB bağlantısı hazır olsun)
//   - Sonrasında her 24 saatte bir
//
// ctx iptal edildiğinde (örn. OS SIGTERM) goroutine düzgünce durur.
func StartGarbageCollector(ctx context.Context, repo *Repository) {
	go func() {
		log.Println("product.GarbageCollector: başlatıldı (ilk çalışma 30 sn sonra).")

		// İlk çalışma için kısa bekleme — sunucu tam ayağa kalksın.
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}

		// İlk GC çalışması.
		runGC(ctx, repo)

		// Sonraki çalışmalar: her 24 saatte bir.
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("product.GarbageCollector: context iptal edildi, durduruluyor.")
				return
			case <-ticker.C:
				runGC(ctx, repo)
			}
		}
	}()
}

// runGC, 30 günden eski soft-deleted ürünleri kalıcı olarak siler.
func runGC(ctx context.Context, repo *Repository) {
	deleted, err := repo.DeleteExpiredSoftDeleted(ctx)
	if err != nil {
		log.Printf("product.GarbageCollector: GC hatası: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("product.GarbageCollector: %d adet 30 günü geçmiş soft-deleted ürün temizlendi.", deleted)
	} else {
		log.Println("product.GarbageCollector: Temizlenecek süresi dolmuş ürün bulunamadı.")
	}
}
