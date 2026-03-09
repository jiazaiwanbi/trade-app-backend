package middleware

import (
	"net/http"
	"time"

	sharedhttp "github.com/jiazaiwanbi/second-hand-platform/internal/shared/httpx"
)

func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, sharedhttp.TimeoutResponse())
	}
}
