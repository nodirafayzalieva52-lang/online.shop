package service

import (
	"context"
	"fmt"
	"time"

	"shop/internal/models"
	"shop/internal/repository"
	pkgerr "shop/pkg/errors"
	"shop/pkg/hash"
	"shop/pkg/jwt"
)

type UserService struct {
	UserRepo         repository.UserRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	JWTService       *jwt.Service
}

func NewUserService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService *jwt.Service,
) *UserService {
	return &UserService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		JWTService:       jwtService,
	}
}

func (s *UserService) Register(ctx context.Context, email, password string) error {
	existing, err := s.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("userRepo.GetByEmail: %w", err)
	}
	if existing != nil {
		return pkgerr.ErrUserAlreadyExists
	}

	hashedPassword, err := hash.Generate(password)
	if err != nil {
		return fmt.Errorf("hash.Generate: %w", err)
	}

	user := &models.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         models.RoleCustomer,
	}

	if err := s.UserRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("userRepo.Create: %w", err)
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("userRepo.GetByEmail: %w", err)
	}
	if user == nil {
		return "", "", pkgerr.ErrAccessDenied
	}
	if !hash.Compare(user.PasswordHash, password) {
		return "", "", pkgerr.ErrAccessDenied
	}

	accessToken, err := s.JWTService.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", "", fmt.Errorf("access token creation failed: %w", err)
	}

	refreshTokenStr, err := s.JWTService.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("refresh token creation failed: %w", err)
	}

	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}

	if err := s.RefreshTokenRepo.Create(ctx, refreshTokenModel); err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return accessToken, refreshTokenStr, nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	if refreshTokenStr == "" {
		return "", "", pkgerr.ErrInvalidToken
	}

	tokenModel, err := s.RefreshTokenRepo.GetByToken(ctx, refreshTokenStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to get refresh token: %w", err)
	}
	if tokenModel == nil || time.Now().After(tokenModel.ExpiresAt) {
		if tokenModel != nil {
			_ = s.RefreshTokenRepo.DeleteByToken(ctx, refreshTokenStr)
		}
		return "", "", pkgerr.ErrInvalidToken
	}

	user, err := s.UserRepo.GetByID(ctx, tokenModel.UserID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", "", pkgerr.ErrUserNotFound
	}

	// Token rotation: delete old refresh token
	_ = s.RefreshTokenRepo.DeleteByToken(ctx, refreshTokenStr)

	newAccessToken, err := s.JWTService.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshTokenStr, err := s.JWTService.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	newRefreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshTokenStr,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	if err := s.RefreshTokenRepo.Create(ctx, newRefreshTokenModel); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return newAccessToken, newRefreshTokenStr, nil
}

func (s *UserService) Logout(ctx context.Context, refreshTokenStr string) error {
	if refreshTokenStr == "" {
		return nil
	}
	return s.RefreshTokenRepo.DeleteByToken(ctx, refreshTokenStr)
}
