package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"shop/handler/middleware"
	"shop/internal/models"
	"shop/internal/service"
	pkgerr "shop/pkg/errors"
	"shop/pkg/logger"
)

type OrderHandler struct {
	OrderService *service.OrderService
	log          *logger.Logger
}

func NewOrderHandler(orderService *service.OrderService, log *logger.Logger) *OrderHandler {
	return &OrderHandler{
		OrderService: orderService,
		log:          log,
	}
}

type OrderItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type CreateOrderRequest struct {
	CustomerID int64              `json:"customer_id"`
	ProductID  int64              `json:"product_id"`
	Quantity   int                `json:"quantity"`
	Items      []OrderItemRequest `json:"items"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body format")
		return
	}

	authID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	customerID := authID
	if userRole == string(models.RoleAdmin) && req.CustomerID > 0 {
		customerID = req.CustomerID
	}
	if customerID <= 0 {
		respondWithError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	var items []models.OrderItem
	if len(req.Items) > 0 {
		for _, item := range req.Items {
			items = append(items, models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
	} else if req.ProductID > 0 && req.Quantity > 0 {
		items = append(items, models.OrderItem{
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		})
	}

	order, err := h.OrderService.Create(r.Context(), customerID, items)
	if err != nil {
		if errors.Is(err, pkgerr.ErrInsufficientStock) ||
			errors.Is(err, pkgerr.ErrEmptyOrder) ||
			errors.Is(err, pkgerr.ErrProductNotFound) ||
			errors.Is(err, pkgerr.ErrMultiStoreOrder) ||
			errors.Is(err, pkgerr.ErrInvalidOrder) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	order, err := h.OrderService.GetByID(r.Context(), id, userID, userRole)
	if err != nil {
		if errors.Is(err, pkgerr.ErrOrderNotFound) {
			respondWithError(w, http.StatusNotFound, "order not found")
			return
		}
		if errors.Is(err, pkgerr.ErrAccessDenied) {
			respondWithError(w, http.StatusForbidden, "access denied to this order")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	authID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	customerID := authID

	targetIDStr := r.URL.Query().Get("user_id")
	if targetIDStr == "" {
		targetIDStr = r.URL.Query().Get("customer_id")
	}

	if targetIDStr != "" {
		targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
		if err != nil || targetID <= 0 {
			respondWithError(w, http.StatusBadRequest, "invalid user_id parameter")
			return
		}
		if userRole != string(models.RoleAdmin) && targetID != authID {
			respondWithError(w, http.StatusForbidden, "access denied to other users' orders")
			return
		}
		customerID = targetID
	}

	if customerID <= 0 {
		respondWithError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	orders, err := h.OrderService.GetByCustomerID(r.Context(), customerID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) GetStoreOrders(w http.ResponseWriter, r *http.Request) {
	storeIDStr := r.PathValue("store_id")
	if storeIDStr == "" {
		storeIDStr = r.URL.Query().Get("store_id")
	}

	storeID, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil || storeID <= 0 {
		respondWithError(w, http.StatusBadRequest, "invalid store id")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	orders, err := h.OrderService.GetByStoreID(r.Context(), storeID, userID, userRole)
	if err != nil {
		if errors.Is(err, pkgerr.ErrAccessDenied) {
			respondWithError(w, http.StatusForbidden, "access denied to this store orders")
			return
		}
		if errors.Is(err, pkgerr.ErrStoreNotFound) {
			respondWithError(w, http.StatusNotFound, "store not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}