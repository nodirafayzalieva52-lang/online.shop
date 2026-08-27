package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	gjwt.RegisteredClaims
}

type Service struct {
	secret []byte
	ttl    time.Duration
}

func NewService(secret string, ttl time.Duration) (*Service, error) {
	if secret == "" {
		return nil, errors.New("jwt: secret must not be empty")
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &Service{secret: []byte(secret), ttl: ttl}, nil
}

func (s *Service) GenerateToken(userID int64, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: gjwt.RegisteredClaims{
			IssuedAt:  gjwt.NewNumericDate(time.Now()),
			ExpiresAt: gjwt.NewNumericDate(time.Now().Add(s.ttl)),
		},
	}
	token := gjwt.NewWithClaims(gjwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("jwt: generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := gjwt.ParseWithClaims(tokenString, &Claims{}, func(t *gjwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gjwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("jwt: invalid token")
	}
	return claims, nil
}
