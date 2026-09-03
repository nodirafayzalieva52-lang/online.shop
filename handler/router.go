package handler

import (
	"io/fs"
	"net/http"

	"shop/docs"
	"shop/handler/middleware"
	"shop/internal/models"
	"shop/pkg/jwt"
	"shop/pkg/logger"
	"shop/web"
)

type Deps struct {
	JWTService      *jwt.Service
	Logger          *logger.Logger
	UserHandler     *UserHandler
	StoreHandler    *StoreHandler
	CategoryHandler *CategoryHandler
	ProductHandler  *ProductHandler
	OrderHandler    *OrderHandler
	ReviewHandler   *ReviewHandler
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

	mux.HandleFunc("GET /stores", d.StoreHandler.GetAll)
	mux.HandleFunc("GET /stores/{id}", d.StoreHandler.GetByID)
	mux.HandleFunc("GET /stores/seller/{seller_id}", d.StoreHandler.GetBySellerID)

	mux.HandleFunc("POST /ai/reviews", d.ReviewHandler.GenerateReview)

	// ---------- PROTECTED WITH RBAC ----------
	mux.Handle("PATCH /users/me/role", auth(http.HandlerFunc(d.UserHandler.SetRole)))

	mux.Handle("POST /stores", auth(requireSellerOrAdmin(http.HandlerFunc(d.StoreHandler.CreateStore))))
	mux.Handle("POST /products", auth(requireSellerOrAdmin(http.HandlerFunc(d.ProductHandler.CreateProduct))))
	mux.Handle("PUT /products/{id}", auth(requireSellerOrAdmin(http.HandlerFunc(d.ProductHandler.UpdateProduct))))

	mux.Handle("POST /categories", auth(requireAdmin(http.HandlerFunc(d.CategoryHandler.CreateCategory))))

	mux.Handle("GET /orders/store/{store_id}", auth(requireSellerOrAdmin(http.HandlerFunc(d.OrderHandler.GetStoreOrders))))

	mux.Handle("POST /orders", auth(http.HandlerFunc(d.OrderHandler.CreateOrder)))
	mux.Handle("GET /orders", auth(http.HandlerFunc(d.OrderHandler.GetUserOrders)))
	mux.Handle("GET /orders/{id}", auth(http.HandlerFunc(d.OrderHandler.GetByID)))

	registerStatic(mux)

	var handler http.Handler = mux
	handler = middleware.CORSMiddleware(handler)
	if d.Logger != nil {
		handler = middleware.LoggingMiddleware(d.Logger)(handler)
		handler = middleware.RecoveryMiddleware(d.Logger)(handler)
	}

	return handler
}

func registerStatic(mux *http.ServeMux) {
	fileServer := http.FileServer(http.FS(web.Files))
	mux.Handle("GET /css/", fileServer)
	mux.Handle("GET /js/", fileServer)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		serveIndex(w)
	})
	mux.HandleFunc("GET /index.html", func(w http.ResponseWriter, r *http.Request) {
		serveIndex(w)
	})
}

func serveIndex(w http.ResponseWriter) {
	data, err := fs.ReadFile(web.Files, "index.html")
	if err != nil {
		http.Error(w, "index not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
