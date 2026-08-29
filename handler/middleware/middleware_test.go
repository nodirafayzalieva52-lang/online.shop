package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shop/handler/middleware"
	"shop/internal/models"
	"shop/pkg/jwt"
)

func TestAuthMiddleware(t *testing.T) {
	jwtSvc, _ := jwt.NewService("test-secret-key-12345", time.Hour)
	token, _ := jwtSvc.GenerateToken(10, "user@test.com", string(models.RoleSeller))

	authMW := middleware.AuthMiddleware(jwtSvc)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID != 10 {
			t.Errorf("expected userID 10, got %d", userID)
		}
		userRole, ok := middleware.GetUserRole(r.Context())
		if !ok || userRole != string(models.RoleSeller) {
			t.Errorf("expected seller role, got %s", userRole)
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := authMW(nextHandler)

	// 1. Valid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	// 2. Missing authorization header
	reqNoAuth := httptest.NewRequest(http.MethodGet, "/test", nil)
	rrNoAuth := httptest.NewRecorder()
	wrapped.ServeHTTP(rrNoAuth, reqNoAuth)
	if rrNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rrNoAuth.Code)
	}

	// 3. Invalid authorization header
	reqBadAuth := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqBadAuth.Header.Set("Authorization", "Bearer invalid-token-string")
	rrBadAuth := httptest.NewRecorder()
	wrapped.ServeHTTP(rrBadAuth, reqBadAuth)
	if rrBadAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rrBadAuth.Code)
	}
}

func TestRequireRole(t *testing.T) {
	jwtSvc, _ := jwt.NewService("test-secret-key-12345", time.Hour)
	sellerToken, _ := jwtSvc.GenerateToken(10, "seller@test.com", string(models.RoleSeller))
	customerToken, _ := jwtSvc.GenerateToken(20, "customer@test.com", string(models.RoleCustomer))

	authMW := middleware.AuthMiddleware(jwtSvc)
	roleMW := middleware.RequireRole(models.RoleSeller, models.RoleAdmin)

	testHandler := authMW(roleMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Seller should pass
	req1 := httptest.NewRequest(http.MethodPost, "/seller-only", nil)
	req1.Header.Set("Authorization", "Bearer "+sellerToken)
	rr1 := httptest.NewRecorder()
	testHandler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected seller to access route, got %d", rr1.Code)
	}

	// Customer should be forbidden (403)
	req2 := httptest.NewRequest(http.MethodPost, "/seller-only", nil)
	req2.Header.Set("Authorization", "Bearer "+customerToken)
	rr2 := httptest.NewRecorder()
	testHandler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("expected customer to be forbidden (403), got %d", rr2.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	recoveryMW := middleware.RecoveryMiddleware(nil)

	panicHandler := recoveryMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went critically wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()
	panicHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on recovered panic, got %d", rr.Code)
	}
}
