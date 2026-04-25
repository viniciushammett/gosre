// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// githubEndpoint defines GitHub's OAuth2 authorization and token URLs.
var githubEndpoint = oauth2.Endpoint{
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
}

// GitHubUser holds the identity fields returned by the GitHub API.
type GitHubUser struct {
	Email string
	Login string
}

// GitHubProvider implements GitHub OAuth2 authentication.
type GitHubProvider struct {
	cfg *oauth2.Config
}

// NewGitHubProvider returns a GitHubProvider for the given app credentials.
func NewGitHubProvider(clientID, clientSecret, redirectURL string) *GitHubProvider {
	return &GitHubProvider{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     githubEndpoint,
		},
	}
}

// AuthURL returns the GitHub authorization URL embedding the given CSRF state.
func (p *GitHubProvider) AuthURL(state string) string {
	return p.cfg.AuthCodeURL(state)
}

// Exchange trades the OAuth2 authorization code for a GitHubUser.
// Fetches the user profile from https://api.github.com/user.
func (p *GitHubProvider) Exchange(ctx context.Context, code string) (GitHubUser, error) {
	token, err := p.cfg.Exchange(ctx, code)
	if err != nil {
		return GitHubUser{}, fmt.Errorf("exchange code: %w", err)
	}

	client := p.cfg.Client(ctx, token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return GitHubUser{}, fmt.Errorf("build user request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return GitHubUser{}, fmt.Errorf("fetch github user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return GitHubUser{}, fmt.Errorf("github API status %d", resp.StatusCode)
	}

	var raw struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return GitHubUser{}, fmt.Errorf("decode github user: %w", err)
	}
	if raw.Email == "" {
		return GitHubUser{}, fmt.Errorf("github user has no public email; grant user:email scope")
	}

	return GitHubUser{Email: raw.Email, Login: raw.Login}, nil
}
