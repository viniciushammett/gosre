// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gosre/gosre-auth/internal/domain"
	"github.com/gosre/gosre-auth/internal/service"
	"github.com/gosre/gosre-auth/internal/store"
)

const testSecret = "test-secret-key"

func newSvc() *service.AuthService {
	return service.New(store.NewMemoryStore(), store.NewMemorySessionStore(), testSecret)
}

func TestRegister_Success(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	u, err := svc.Register(ctx, "alice@example.com", "password123", domain.RoleOperator)
	if err != nil {
		t.Fatalf("Register: unexpected error: %v", err)
	}
	if u.ID == "" {
		t.Error("Register: expected non-empty ID")
	}
	if u.Email != "alice@example.com" {
		t.Errorf("Register: email = %q, want %q", u.Email, "alice@example.com")
	}
	if u.PasswordHash == "password123" {
		t.Error("Register: password must be hashed")
	}
	if u.Role != domain.RoleOperator {
		t.Errorf("Register: role = %q, want %q", u.Role, domain.RoleOperator)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, err := svc.Register(ctx, "bob@example.com", "s3cret!", domain.RoleViewer)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	accessToken, refreshToken, err := svc.Login(ctx, "bob@example.com", "s3cret!")
	if err != nil {
		t.Fatalf("Login: unexpected error: %v", err)
	}
	if accessToken == "" {
		t.Error("Login: expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("Login: expected non-empty refresh token")
	}

	claims, err := svc.ValidateToken(ctx, accessToken)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.Email != "bob@example.com" {
		t.Errorf("claims.Email = %q, want %q", claims.Email, "bob@example.com")
	}
	if claims.Role != domain.RoleViewer {
		t.Errorf("claims.Role = %q, want %q", claims.Role, domain.RoleViewer)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, err := svc.Register(ctx, "carol@example.com", "correct-password", domain.RoleAdmin)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, _, err = svc.Login(ctx, "carol@example.com", "wrong-password")
	if err == nil {
		t.Fatal("Login: expected error for wrong password, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	svc := newSvc()

	// craft an already-expired token signed with the same secret
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test-id",
		"email":   "exp@example.com",
		"role":    string(domain.RoleViewer),
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
		"iat":     time.Now().Add(-2 * time.Hour).Unix(),
	})
	tokenStr, err := expired.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	_, err = svc.ValidateToken(context.Background(), tokenStr)
	if err == nil {
		t.Fatal("ValidateToken: expected error for expired token, got nil")
	}
}

func TestRefresh_Success(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, err := svc.Register(ctx, "dave@example.com", "password123", domain.RoleViewer)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, refreshToken, err := svc.Login(ctx, "dave@example.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	newAccess, newRefresh, err := svc.Refresh(ctx, refreshToken)
	if err != nil {
		t.Fatalf("Refresh: unexpected error: %v", err)
	}
	if newAccess == "" {
		t.Error("Refresh: expected non-empty access token")
	}
	if newRefresh == "" {
		t.Error("Refresh: expected non-empty refresh token")
	}

	_, err = svc.ValidateToken(ctx, newAccess)
	if err != nil {
		t.Fatalf("ValidateToken after Refresh: %v", err)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, _, err := svc.Refresh(ctx, "bogus-token-that-does-not-exist")
	if err == nil {
		t.Fatal("Refresh: expected error for invalid token, got nil")
	}
}

func TestRefresh_Rotation(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, err := svc.Register(ctx, "eve@example.com", "password123", domain.RoleViewer)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, oldRefresh, err := svc.Login(ctx, "eve@example.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// first Refresh succeeds and rotates the token
	_, _, err = svc.Refresh(ctx, oldRefresh)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	// second Refresh with the same (now consumed) token must fail
	_, _, err = svc.Refresh(ctx, oldRefresh)
	if err == nil {
		t.Fatal("Refresh: expected error when reusing a consumed refresh token, got nil")
	}
}

func TestLogout_Success(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, err := svc.Register(ctx, "frank@example.com", "password123", domain.RoleViewer)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, refreshToken, err := svc.Login(ctx, "frank@example.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if err := svc.Logout(ctx, refreshToken); err != nil {
		t.Fatalf("Logout: unexpected error: %v", err)
	}

	// after logout, refresh must be rejected
	_, _, err = svc.Refresh(ctx, refreshToken)
	if err == nil {
		t.Fatal("Refresh after Logout: expected error, got nil")
	}
}
