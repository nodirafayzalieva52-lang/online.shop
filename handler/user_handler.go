package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"shop/handler/middleware"
	"shop/internal/models"
	"shop/internal/service"
	pkgerr "shop/pkg/errors"
	"shop/pkg/logger"

	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
	logger      *logger.Logger
}

func NewUserHandler(userService *service.UserService, log *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      log,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode register body", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		h.logger.Warn("register validation failed: missing email or password")
		respondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	err := h.userService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, pkgerr.ErrUserAlreadyExists) {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, pkgerr.ErrInvalidEmail) || errors.Is(err, pkgerr.ErrWeakPassword) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to register user",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		respondWithError(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	h.logger.Info("user registered successfully", zap.String("email", req.Email))
	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode login body", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		h.logger.Warn("login validation failed: missing email or password")
		respondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	accessToken, refreshToken, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("failed login attempt",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		respondWithError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	h.logger.Info("user logged in successfully", zap.String("email", req.Email))
	respondWithJSON(w, http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		respondWithError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	accessToken, refreshToken, err := h.userService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.RefreshToken != "" {
		_ = h.userService.Logout(r.Context(), req.RefreshToken)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

type SetRoleRequest struct {
	Role string `json:"role"`
}

func (h *UserHandler) SetRole(w http.ResponseWriter, r *http.Request) {
	var req SetRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID <= 0 {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accessToken, refreshToken, err := h.userService.SetRole(r.Context(), userID, models.Role(req.Role))
	if err != nil {
		if errors.Is(err, pkgerr.ErrInvalidRole) {
			respondWithError(w, http.StatusBadRequest, "role must be customer or seller")
			return
		}
		if errors.Is(err, pkgerr.ErrAccessDenied) {
			respondWithError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, pkgerr.ErrUserNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Error("failed to set role", zap.Int64("user_id", userID), zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}
