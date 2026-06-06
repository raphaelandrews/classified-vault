package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/ds"
)

type contextKey string

const (
	ctxKeySession contextKey = "session"
	ctxKeyToken   contextKey = "token"
)

func RequireAuth(cache *ds.HashMap[auth.Session]) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			session, ok := cache.Get(token)
			if !ok || time.Now().After(session.ExpiresAt) {
				if ok {
					cache.Delete(token)
				}
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = contextWithSession(ctx, session)
			ctx = contextWithToken(ctx, token)

			w.Header().Set("X-Session-Expires", session.ExpiresAt.Format(time.RFC3339))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext(r.Context())
			if session == nil || session.Role != role {
				http.Error(w, `{"error":"permission denied"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAnyRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext(r.Context())
			if session == nil {
				http.Error(w, `{"error":"permission denied"}`, http.StatusForbidden)
				return
			}
			for _, role := range roles {
				if session.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"permission denied"}`, http.StatusForbidden)
		})
	}
}

func RequireClearance(level domain.ClearanceLevel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext(r.Context())
			if session == nil || session.Clearance < level {
				http.Error(w, `{"error":"insufficient tier"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func contextWithSession(ctx context.Context, session auth.Session) context.Context {
	return context.WithValue(ctx, ctxKeySession, session)
}

func SessionFromContext(ctx context.Context) *auth.Session {
	s, ok := ctx.Value(ctxKeySession).(auth.Session)
	if !ok {
		return nil
	}
	return &s
}

func TokenFromContext(ctx context.Context) string {
	t, ok := ctx.Value(ctxKeyToken).(string)
	if !ok {
		return ""
	}
	return t
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxKeyToken, token)
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
