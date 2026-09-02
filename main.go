package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shop/handler"
	"shop/internal/config"
	"shop/internal/repository/postgres"
	"shop/internal/service"
	"shop/migrations"
	"shop/pkg/jwt"
	"shop/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New("config.env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger, err := logger.New(true)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer func() {
		_ = appLogger.Sync()
	}()

	ttl, err := time.ParseDuration(cfg.JWT.TTL)
	if err != nil {
		log.Fatalf("invalid JWT_TTL: %v", err)
	}

	jwtService, err := jwt.NewService(cfg.JWT.Secret, ttl)
	if err != nil {
		log.Fatalf("failed to init jwt service: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()


	// ---------- AUTO MIGRATIONS ----------
	if err := migrations.Run(ctx, pool); err != nil {
		appLogger.Warn("Database auto-migration failed or skipped", zap.Error(err))
	}

	// ---------- REFRESH TOKEN & USER ----------
	userRepo := postgres.NewUserRepository(pool)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(pool)
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtService)
	userHandler := handler.NewUserHandler(userService, appLogger)

	// ---------- STORE ----------
	storeRepo := postgres.NewStoreRepository(pool)
	storeService := service.NewStoreService(storeRepo)
	storeHandler := handler.NewStoreHandler(storeService)

	// ---------- CATEGORY ----------
	categoryRepo := postgres.NewCategoryRepository(pool)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	// ---------- PRODUCT ----------
	productRepo := postgres.NewProductRepository(pool)
	productService := service.NewProductService(productRepo, storeRepo)
	productHandler := handler.NewProductHandler(productService)

	// ---------- ORDER ----------
	orderRepo := postgres.NewOrderRepository(pool)
	orderService := service.NewOrderService(orderRepo, productRepo, storeRepo)
	orderHandler := handler.NewOrderHandler(orderService, appLogger)

	router := handler.NewRouter(handler.Deps{
		JWTService:      jwtService,
		Logger:          appLogger,
		UserHandler:     userHandler,
		StoreHandler:    storeHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
	})

	server := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		appLogger.Info("Server is running", zap.String("port", cfg.HTTPPort))
		appLogger.Info(fmt.Sprintf("Swagger UI available at http://localhost%s/swagger/", cfg.HTTPPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		appLogger.Fatal("Server error", zap.Error(err))
	case <-ctx.Done():
		appLogger.Info("Shutting down server gracefully...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("Server forced to shutdown", zap.Error(err))
			_ = server.Close()
		}
		appLogger.Info("Server exited cleanly")
	}
}