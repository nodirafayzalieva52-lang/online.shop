package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"shop/handler/middleware"
	"shop/internal/service"
	pkgerr "shop/pkg/errors"
)

type ProductHandler struct {
	ProductService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		ProductService: productService,
	}
}

type CreateProductRequest struct {
	StoreID     int64   `json:"store_id"`
	CategoryID  int64   `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body format")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	product, err := h.ProductService.CreateProduct(
		r.Context(),
		userID,
		userRole,
		req.StoreID,
		req.CategoryID,
		req.Name,
		req.Description,
		req.Price,
		req.Stock,
	)
	if err != nil {
		if errors.Is(err, pkgerr.ErrAccessDenied) {
			respondWithError(w, http.StatusForbidden, "you do not have permission to add products to this store")
			return
		}
		if errors.Is(err, pkgerr.ErrStoreNotFound) {
			respondWithError(w, http.StatusNotFound, "store not found")
			return
		}
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.ProductService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerr.ErrProductNotFound) {
			respondWithError(w, http.StatusNotFound, "product not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) GetByStoreID(w http.ResponseWriter, r *http.Request) {
	storeIDStr := r.PathValue("store_id")
	if storeIDStr == "" {
		storeIDStr = r.URL.Query().Get("store_id")
	}

	storeID, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil || storeID <= 0 {
		respondWithError(w, http.StatusBadRequest, "invalid store id")
		return
	}

	products, err := h.ProductService.GetByStoreID(r.Context(), storeID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	products, err := h.ProductService.GetAll(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}
