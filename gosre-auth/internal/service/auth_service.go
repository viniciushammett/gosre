// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/gosre/gosre-auth/internal/domain"
	"github.com/gosre/gosre-auth/internal/store"
)

const refreshTokenTTL = 7 * 24 * time.Hour

var (
	// ErrInvalidCredentials is returned when email or password do not match.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken is returned when a JWT is malformed, expired, or has wrong signature.
	ErrInvalidToken = errors.New("invalid token")
	// ErrInvalidRefreshToken is returned when a refresh token is missing, expired, or already used.
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

// AuthService handles registration, login, token validation and session management.
type AuthService struct {
	users     store.UserStore
	sessions  store.SessionStore
	jwtSecret []byte
}

// New returns an AuthService backed by the given stores and JWT secret.
func New(users store.UserStore, sessions store.SessionStore, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new user with a bcrypt-hashed password.
func (a *AuthService) Register(ctx context.Context, email, password string, role domain.Role) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	u := domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}

	if err := a.users.Create(ctx, u); err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Login verifies credentials and returns a short-lived access token (15 min)
// and a long-lived refresh token (7 days) on success.
func (a *AuthService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	u, lookupErr := a.users.GetByEmail(ctx, email)
	if lookupErr != nil {
		return "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return a.issueTokenPair(ctx, u)
}

// ValidateToken parses and validates a JWT, returning the embedded claims.
func (a *AuthService) ValidateToken(_ context.Context, tokenString string) (domain.Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return a.jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		return domain.Claims{}, ErrInvalidToken
	}

	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return domain.Claims{}, ErrInvalidToken
	}

	userID, ok1 := mc["user_id"].(string)
	email, ok2 := mc["email"].(string)
	role, ok3 := mc["role"].(string)

	if !ok1 || !ok2 || !ok3 || userID == "" || email == "" || role == "" {
		return domain.Claims{}, ErrInvalidToken
	}

	return domain.Claims{
		UserID: userID,
		Email:  email,
		Role:   domain.Role(role),
	}, nil
}

// FindOrCreate returns the user with the given email, or creates one with the
// given role if not found. Used by OAuth flows where there is no password.
func (a *AuthService) FindOrCreate(ctx context.Context, email string, role domain.Role) (domain.User, error) {
	u, err := a.users.GetByEmail(ctx, email)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, store.ErrUserNotFound) {
		return domain.User{}, fmt.Errorf("lookup user: %w", err)
	}

	u = domain.User{
		ID:        uuid.NewString(),
		Email:     email,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}
	if err := a.users.Create(ctx, u); err != nil {
		return domain.User{}, fmt.Errorf("create oauth user: %w", err)
	}
	return u, nil
}

// IssueToken creates and signs an access + refresh token pair for the given user.
// Used by OAuth flows after FindOrCreate.
func (a *AuthService) IssueToken(ctx context.Context, u domain.User) (accessToken, refreshToken string, err error) {
	return a.issueTokenPair(ctx, u)
}

// Refresh validates a refresh token, rotates it, and returns a new token pair.
// The old refresh token is deleted on success (single-use rotation).
func (a *AuthService) Refresh(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	userID, getErr := a.sessions.Get(ctx, refreshToken)
	if getErr != nil {
		return "", "", ErrInvalidRefreshToken
	}

	if delErr := a.sessions.Delete(ctx, refreshToken); delErr != nil {
		return "", "", fmt.Errorf("revoke refresh token: %w", delErr)
	}

	u, lookupErr := a.users.GetByID(ctx, userID)
	if lookupErr != nil {
		return "", "", fmt.Errorf("lookup user: %w", lookupErr)
	}

	return a.issueTokenPair(ctx, u)
}

// Logout invalidates a refresh token, ending the session.
func (a *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return a.sessions.Delete(ctx, refreshToken)
}

// issueTokenPair signs a 15-min access JWT and generates a 7-day refresh token.
func (a *AuthService) issueTokenPair(ctx context.Context, u domain.User) (string, string, error) {
	accessToken, err := a.sign(u)
	if err != nil {
		return "", "", err
	}

	refreshToken := uuid.NewString()
	if err := a.sessions.Save(ctx, refreshToken, u.ID, refreshTokenTTL); err != nil {
		return "", "", fmt.Errorf("save session: %w", err)
	}

	return accessToken, refreshToken, nil
}

// sign builds and returns a signed HS256 JWT valid for 15 minutes.
func (a *AuthService) sign(u domain.User) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"role":    string(u.Role),
		"exp":     now.Add(15 * time.Minute).Unix(),
		"iat":     now.Unix(),
	})

	signed, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
