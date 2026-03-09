package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/logger"
)

const requestIDHeader = "X-Request-Id"

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = newRequestID()
			}

			ctx := logger.WithRequestID(r.Context(), requestID)
			w.Header().Set(requestIDHeader, requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func newRequestID() string {
	buf := make([]byte, 12)
	_, err := rand.Read(buf)
	if err != nil {
		return "request-id-unavailable"
	}

	return hex.EncodeToString(buf)
}
