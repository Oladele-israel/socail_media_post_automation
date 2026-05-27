// internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Oladele-israel/socialmedia-post-automation/pkg/cache"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/token"
)

type contextKey string

const (
	userIDKey      contextKey = "user_id"
	userEmailKey   contextKey = "user_email"
	tokenIDKey     contextKey = "token_id"
	tokenExpiryKey contextKey = "token_expiry"
)

func RequireAuth(cache *cache.RedisClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := token.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			blacklisted, err := cache.IsTokenBlacklisted(r.Context(), claims.ID)
			if err != nil || blacklisted {
				http.Error(w, `{"error":"token has been revoked"}`, http.StatusUnauthorized)
				return
			}

			// Store everything downstream handlers might need
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userEmailKey, claims.Email)
			ctx = context.WithValue(ctx, tokenIDKey, claims.ID)
			ctx = context.WithValue(ctx, tokenExpiryKey, claims.ExpiresAt.Time)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper functions — handlers call these instead of touching context directly

func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(userIDKey).(string)
	return id
}

func GetUserEmail(r *http.Request) string {
	email, _ := r.Context().Value(userEmailKey).(string)
	return email
}

func GetTokenID(r *http.Request) string {
	id, _ := r.Context().Value(tokenIDKey).(string)
	return id
}

func GetTokenExpiry(r *http.Request) time.Time {
	expiry, _ := r.Context().Value(tokenExpiryKey).(time.Time)
	return expiry
}
