package middleware

import "net/http"

// CORSMiddleware, tüm HTTP yanıtlarına CORS başlıklarını ekler.
// Web (Flutter Web / tarayıcı) isteklerindeki preflight (OPTIONS) kontrollerini
// de karşılar; böylece farklı origin'den gelen fetch istekleri engellenmez.
//
// Geliştirme ortamı için tüm origin'lere (*) açık bırakılmıştır.
// Prodüksiyona geçişte ALLOWED_ORIGINS env değişkeniyle kısıtlanabilir.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// İzin verilen origin — geliştirme için herkese açık
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 saat önbellek

		// Preflight isteğini (OPTIONS) burada bitir — body gerekmez
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
