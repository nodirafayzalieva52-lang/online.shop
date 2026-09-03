package handler

import (
	"encoding/json"
	"net/http"

	"shop/ai"
)

type GenerateReviewRequest struct {
	ProductName string `json:"product_name"`
	Tone        string `json:"tone"`
	Rating      int    `json:"rating"`
}

type GenerateReviewResponse struct {
	Review string `json:"review"`
}

type ReviewHandler struct{}

func NewReviewHandler() *ReviewHandler {
	return &ReviewHandler{}
}

func (h *ReviewHandler) GenerateReview(w http.ResponseWriter, r *http.Request) {
	var req GenerateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body format")
		return
	}

	if req.ProductName == "" {
		respondWithError(w, http.StatusBadRequest, "product_name is required")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		respondWithError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}
	if req.Tone == "" {
		req.Tone = string(ai.Tonepraise)
	}

	review, err := ai.GenerateReview(req.ProductName, ai.ReviewTone(req.Tone), req.Rating)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, GenerateReviewResponse{Review: review})
}