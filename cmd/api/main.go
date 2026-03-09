package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	authapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/auth"
	categoryapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/category"
	listingapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/listing"
	orderapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/order"
	userapp "github.com/jiazaiwanbi/second-hand-platform/internal/application/user"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/auth"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/config"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/database"
	platformhttp "github.com/jiazaiwanbi/second-hand-platform/internal/platform/httpserver"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/logger"
	postgresrepo "github.com/jiazaiwanbi/second-hand-platform/internal/repository/postgres"
	"github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/handler"
	httprouter "github.com/jiazaiwanbi/second-hand-platform/internal/transport/http/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.New()
	appLogger.Info("application starting", map[string]any{
		"app_env":       cfg.AppEnv,
		"http_addr":     cfg.HTTPAddr,
		"postgres_host": cfg.Postgres.Host,
		"postgres_db":   cfg.Postgres.Database,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Postgres.URL())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	userRepo := postgresrepo.NewUserRepository(pool)
	categoryRepo := postgresrepo.NewCategoryRepository(pool)
	listingRepo := postgresrepo.NewListingRepository(pool)
	orderRepo := postgresrepo.NewOrderRepository(pool)
	tokenManager := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)
	authService := authapp.NewService(userRepo, tokenManager)
	userService := userapp.NewService(userRepo)
	categoryService := categoryapp.NewService(categoryRepo)
	listingService := listingapp.NewService(listingRepo)
	orderService := orderapp.NewService(orderRepo)

	readyHandler := handler.NewReadyHandler(pool)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	listingHandler := handler.NewListingHandler(listingService, categoryService)
	orderHandler := handler.NewOrderHandler(orderService)

	server := platformhttp.New(
		cfg.HTTPAddr,
		cfg.ShutdownWait,
		httprouter.New(appLogger, cfg.HTTPTimeout, tokenManager, readyHandler, authHandler, userHandler, listingHandler, orderHandler),
	)

	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
