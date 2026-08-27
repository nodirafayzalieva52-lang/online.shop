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

type StoreHandler struct {
	storeService *service.StoreService
}

func NewStoreHandler(storeService *service.StoreService) *StoreHandler {
	return &StoreHandler{
		storeService: storeService,
	}
}

type createStoreRequest struct {
	SellerID    int64  `json:"seller_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *StoreHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	var req createStoreRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sellerID := req.SellerID
	if sellerID <= 0 {
		if authID, ok := middleware.GetUserID(r.Context()); ok && authID > 0 {
			sellerID = authID
		} else {
			respondWithError(w, http.StatusBadRequest, "seller_id is required")
			return
		}
	}

	store, err := h.storeService.CreateStore(r.Context(), sellerID, req.Name, req.Description)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, store)
}

func (h *StoreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "invalid store id")
		return
	}

	store, err := h.storeService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerr.ErrStoreNotFound) {
			respondWithError(w, http.StatusNotFound, "store not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, store)
}

func (h *StoreHandler) GetBySellerID(w http.ResponseWriter, r *http.Request) {
	sellerIDStr := r.PathValue("seller_id")
	if sellerIDStr == "" {
		sellerIDStr = r.URL.Query().Get("seller_id")
	}
	if sellerIDStr == "" {
		if authID, ok := middleware.GetUserID(r.Context()); ok && authID > 0 {
			sellerIDStr = strconv.FormatInt(authID, 10)
		}
	}

	sellerID, err := strconv.ParseInt(sellerIDStr, 10, 64)
	if err != nil || sellerID <= 0 {
		respondWithError(w, http.StatusBadRequest, "invalid seller id")
		return
	}

	store, err := h.storeService.GetBySellerID(r.Context(), sellerID)
	if err != nil {
		if errors.Is(err, pkgerr.ErrStoreNotFound) {
			respondWithError(w, http.StatusNotFound, "store not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, store)
}