package router

import (
	"net/http"
	"time"

	platformauth "github.com/jiazaiwanbi/second-hand-platform/internal/platform/auth"
	platformmiddleware "github.com/jiazaiwanbi/second-hand-platform/internal/platform/http/middleware"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/logger"
	"github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/handler"
)

func New(
	appLogger *logger.Logger,
	timeout time.Duration,
	tokenManager *platformauth.TokenManager,
	readyHandler http.Handler,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	listingHandler *handler.ListingHandler,
	orderHandler *handler.OrderHandler,
) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("/healthz", handler.Healthz)
	r.Handle("/readyz", readyHandler)
	r.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	r.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	r.HandleFunc("GET /api/v1/categories", listingHandler.ListCategories)
	r.HandleFunc("GET /api/v1/listings", listingHandler.List)
	r.HandleFunc("GET /api/v1/listings/{id}", listingHandler.Get)
	r.Handle("/api/v1/listings", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			listingHandler.Create(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}), platformmiddleware.Auth(tokenManager)))
	r.Handle("/api/v1/listings/{id}", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			listingHandler.Update(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}), platformmiddleware.Auth(tokenManager)))
	r.Handle("/api/v1/orders", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			orderHandler.Create(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}), platformmiddleware.Auth(tokenManager)))
	r.Handle("/api/v1/orders/{id}/cancel", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			orderHandler.Cancel(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}), platformmiddleware.Auth(tokenManager)))
	r.Handle("/api/v1/orders/{id}/complete", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			orderHandler.Complete(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}), platformmiddleware.Auth(tokenManager)))
	r.Handle("/api/v1/users/me", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetMe(w, r)
		case http.MethodPatch:
			userHandler.UpdateMe(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), platformmiddleware.Auth(tokenManager)))
	r.Handle("/api/v1/users/me/listings", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			listingHandler.ListMine(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}), platformmiddleware.Auth(tokenManager)))
	r.Handle("/api/v1/users/me/orders", platformmiddleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderHandler.ListMine(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}), platformmiddleware.Auth(tokenManager)))

	return platformmiddleware.Chain(
		r,
		platformmiddleware.Timeout(timeout),
		platformmiddleware.RequestID(),
		platformmiddleware.Recovery(appLogger),
		platformmiddleware.RequestLogger(appLogger),
	)
}
