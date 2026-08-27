package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"shop/internal/models"
	"shop/internal/service"
	pkgerr "shop/pkg/errors"
)

type CategoryHandler struct {
	CategoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		CategoryService: categoryService,
	}
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	category, err := h.CategoryService.CreateCategory(r.Context(), req.Name)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	category, err := h.CategoryService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerr.ErrCategoryNotFound) {
			respondWithError(w, http.StatusNotFound, "category not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	categories, err := h.CategoryService.GetAll(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if categories == nil {
		categories = []*models.Category{}
	}

	respondWithJSON(w, http.StatusOK, categories)
}
