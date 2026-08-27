package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"shop/handler"
	"shop/internal/config"
	"shop/internal/repository/postgres"
	"shop/internal/service"
	"shop/migrations"
	"shop/pkg/jwt"
	"shop/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New("config.env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger, err := logger.New(true)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	ttl, err := time.ParseDuration(cfg.JWT.TTL)
	if err != nil {
		log.Fatalf("invalid JWT_TTL: %v", err)
	}

	jwtService, err := jwt.NewService(cfg.JWT.Secret, ttl)
	if err != nil {
		log.Fatalf("failed to init jwt service: %v", err)
	}

	migratorDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	// ---------- AUTO MIGRATIONS ----------
	if err := migrations.Run(migratorDSN); err != nil {
		log.Printf("Warning: Database auto-migration failed or skipped: %v", err)
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
		UserHandler:     userHandler,
		StoreHandler:    storeHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
	})

	server := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: router,
	}

	log.Printf("Server is running on %s", cfg.HTTPPort)
	log.Printf("Swagger UI available at http://localhost%s/swagger/", cfg.HTTPPort)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
