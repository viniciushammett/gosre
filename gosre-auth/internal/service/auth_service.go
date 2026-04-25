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

var (
	// ErrInvalidCredentials is returned when email or password do not match.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken is returned when a JWT is malformed, expired, or has wrong signature.
	ErrInvalidToken = errors.New("invalid token")
)

// AuthService handles registration, login and token validation.
type AuthService struct {
	store     store.UserStore
	jwtSecret []byte
}

// New returns an AuthService backed by the given store and JWT secret.
func New(s store.UserStore, jwtSecret string) *AuthService {
	return &AuthService{
		store:     s,
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

	if err := a.store.Create(ctx, u); err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Login verifies credentials and returns a signed JWT on success.
func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := a.store.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return a.sign(u)
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

// sign builds and returns a signed HS256 JWT for the given user.
func (a *AuthService) sign(u domain.User) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"role":    string(u.Role),
		"exp":     now.Add(24 * time.Hour).Unix(),
		"iat":     now.Unix(),
	})

	signed, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
