package middleware

import (
	"context"
	"net/http"
	"strings"

	authdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/auth"
	platformauth "github.com/jiazaiwanbi/second-hand-platform/internal/platform/auth"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
)

type contextKey string

const userIDKey contextKey = "user_id"

func Auth(tokenManager *platformauth.TokenManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: authdomain.ErrInvalidToken.Error()})
				return
			}

			claims, err := tokenManager.Parse(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				sharedhttp.WriteError(w, r, sharedhttp.AppError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: authdomain.ErrInvalidToken.Error()})
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}
