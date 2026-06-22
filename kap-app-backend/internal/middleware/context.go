package middleware

import "context"

// contextKey, bu paket içinde context.Context'e yazılan değerlerin
// çakışmasını önlemek için kullanılan özel (unexported) tip.
// String tabanlı anahtarların başka paketlerle çakışma riskini ortadan kaldırır.
type contextKey string

const (
	// ContextKeyUserID, doğrulanmış kullanıcının UUID'sini (JWT sub alanı)
	// context içinde saklayan anahtar.
	ContextKeyUserID contextKey = "userID"
)

// contextWithUserID, verilen userID'yi context'e gömer ve yeni context'i döner.
// Yalnızca bu paket içinden (middleware.go) çağrılır.
func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// UserIDFromContext, context'ten kullanıcı UUID'sini güvenle okur.
// Handler'larda kullanım örneği:
//
//	userID, ok := middleware.UserIDFromContext(r.Context())
//	if !ok {
//	    // Bu noktaya AuthMiddleware bypass edilmeden ulaşılamaz; defensive check.
//	    http.Error(w, "yetkisiz", http.StatusUnauthorized)
//	    return
//	}
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ContextKeyUserID).(string)
	return id, ok && id != ""
}
