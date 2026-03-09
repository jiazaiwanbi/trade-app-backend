package middleware

import (
	"fmt"
	"net/http"

	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/logger"
	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
)

func Recovery(appLogger *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					appLogger.ErrorContext(r.Context(), "panic recovered", map[string]any{
						"panic": fmt.Sprint(recovered),
						"path":  r.URL.Path,
					})

					sharedhttp.WriteError(w, r, sharedhttp.NewInternalError())
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
