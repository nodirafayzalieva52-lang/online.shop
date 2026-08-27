package handler

import (
	"net/http"

	"shop/docs"
	"shop/handler/middleware"
	"shop/internal/models"
	"shop/pkg/jwt"
)

type Deps struct {
	JWTService      *jwt.Service
	UserHandler     *UserHandler
	StoreHandler    *StoreHandler
	CategoryHandler *CategoryHandler
	ProductHandler  *ProductHandler
	OrderHandler    *OrderHandler
}

func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	auth := middleware.AuthMiddleware(d.JWTService)

	requireSellerOrAdmin := middleware.RequireRole(models.RoleSeller, models.RoleAdmin)
	requireAdmin := middleware.RequireRole(models.RoleAdmin)

	// ---------- SWAGGER ----------
	mux.HandleFunc("GET /swagger/", docs.UIHandler)
	mux.HandleFunc("GET /swagger/doc.json", docs.JSONHandler)

	// ---------- PUBLIC ----------
	mux.HandleFunc("POST /register", d.UserHandler.Register)
	mux.HandleFunc("POST /login", d.UserHandler.Login)
	mux.HandleFunc("POST /refresh", d.UserHandler.Refresh)
	mux.HandleFunc("POST /logout", d.UserHandler.Logout)

	mux.HandleFunc("GET /categories", d.CategoryHandler.GetAll)
	mux.HandleFunc("GET /categories/{id}", d.CategoryHandler.GetByID)

	mux.HandleFunc("GET /products", d.ProductHandler.GetAll)
	mux.HandleFunc("GET /products/{id}", d.ProductHandler.GetByID)
	mux.HandleFunc("GET /products/store/{store_id}", d.ProductHandler.GetByStoreID)

	mux.HandleFunc("GET /stores/{id}", d.StoreHandler.GetByID)
	mux.HandleFunc("GET /stores/seller/{seller_id}", d.StoreHandler.GetBySellerID)

	// ---------- PROTECTED WITH RBAC ----------
	mux.Handle("POST /stores", auth(requireSellerOrAdmin(http.HandlerFunc(d.StoreHandler.CreateStore))))
	mux.Handle("POST /products", auth(requireSellerOrAdmin(http.HandlerFunc(d.ProductHandler.CreateProduct))))

	mux.Handle("POST /categories", auth(requireAdmin(http.HandlerFunc(d.CategoryHandler.CreateCategory))))

	mux.Handle("GET /orders/store/{store_id}", auth(requireSellerOrAdmin(http.HandlerFunc(d.OrderHandler.GetStoreOrders))))

	mux.Handle("POST /orders", auth(http.HandlerFunc(d.OrderHandler.CreateOrder)))
	mux.Handle("GET /orders", auth(http.HandlerFunc(d.OrderHandler.GetUserOrders)))
	mux.Handle("GET /orders/{id}", auth(http.HandlerFunc(d.OrderHandler.GetByID)))

	return mux
}