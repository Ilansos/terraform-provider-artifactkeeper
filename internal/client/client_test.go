package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNormalizeBaseURL(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"https://artifactkeeper.example.com":         "https://artifactkeeper.example.com/api/v1",
		"https://artifactkeeper.example.com/":        "https://artifactkeeper.example.com/api/v1",
		"https://artifactkeeper.example.com/api/v1":  "https://artifactkeeper.example.com/api/v1",
		"https://artifactkeeper.example.com/custom":  "https://artifactkeeper.example.com/custom/api/v1",
		"https://artifactkeeper.example.com/api/v1/": "https://artifactkeeper.example.com/api/v1",
	}

	for input, expected := range tests {
		input, expected := input, expected
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeBaseURL(input)
			if err != nil {
				t.Fatalf("NormalizeBaseURL returned error: %v", err)
			}
			if got.String() != expected {
				t.Fatalf("got %q, want %q", got.String(), expected)
			}
		})
	}
}

func TestTokenAuthAndQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/repositories" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("per_page") != "100" {
			t.Fatalf("per_page query = %q", r.URL.Query().Get("per_page"))
		}
		if got := r.Header.Get("Authorization"); got != "Bearer direct-token" {
			t.Fatalf("authorization = %q", got)
		}
		writeJSON(t, w, RepositoryListResponse{Items: []Repository{}, Pagination: Pagination{Page: 1, PerPage: 100}})
	}))
	defer server.Close()

	c, err := New(context.Background(), Config{BaseURL: server.URL, Token: "direct-token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := c.ListRepositories(context.Background(), RepositoryListOptions{PerPage: 100}); err != nil {
		t.Fatalf("ListRepositories returned error: %v", err)
	}
}

func TestLoginFlow(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			var req loginRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode login request: %v", err)
			}
			if req.Username != "alice" || req.Password != "secret" {
				t.Fatalf("unexpected credentials: %#v", req)
			}
			writeJSON(t, w, loginResponse{AccessToken: "login-token", TokenType: "Bearer", ExpiresIn: 900})
		case "/api/v1/repositories/demo":
			if got := r.Header.Get("Authorization"); got != "Bearer login-token" {
				t.Fatalf("authorization = %q", got)
			}
			writeJSON(t, w, Repository{
				ID:               "repo-id",
				Key:              "demo",
				Name:             "Demo",
				Format:           "docker",
				RepoType:         "local",
				IsPublic:         true,
				StorageUsedBytes: 42,
				CreatedAt:        "2026-01-01T00:00:00Z",
				UpdatedAt:        "2026-01-01T00:00:00Z",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c, err := New(context.Background(), Config{BaseURL: server.URL, Username: "alice", Password: "secret"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := c.GetRepository(context.Background(), "demo"); err != nil {
		t.Fatalf("GetRepository returned error: %v", err)
	}
}

func TestErrorDecoding(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "3")
		w.WriteHeader(http.StatusTooManyRequests)
		writeJSON(t, w, map[string]any{
			"error": map[string]any{
				"code":    "RATE_LIMITED",
				"message": "slow down",
				"details": map[string]any{"limit": 1},
			},
		})
	}))
	defer server.Close()

	c, err := New(context.Background(), Config{BaseURL: server.URL, Token: "token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = c.GetRepository(context.Background(), "demo")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests || apiErr.Code != "RATE_LIMITED" || apiErr.Message != "slow down" {
		t.Fatalf("unexpected API error: %#v", apiErr)
	}
	if apiErr.RetryAfter == nil || *apiErr.RetryAfter != 3*time.Second {
		t.Fatalf("RetryAfter = %v", apiErr.RetryAfter)
	}
}

func TestFlatErrorDecoding(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		writeJSON(t, w, map[string]any{
			"code":    "CONFLICT",
			"message": "already exists",
		})
	}))
	defer server.Close()

	c, err := New(context.Background(), Config{BaseURL: server.URL, Token: "token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	err = c.DeleteRepository(context.Background(), "demo")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusConflict || apiErr.Code != "CONFLICT" || apiErr.Message != "already exists" {
		t.Fatalf("unexpected API error: %#v", apiErr)
	}
}

func TestRepositoryJSONMapping(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/repositories" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req CreateRepositoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Key != "docker-local" || req.Name != "Docker Local" || req.Format != "docker" || req.RepoType != "local" {
			t.Fatalf("unexpected create request: %#v", req)
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(t, w, Repository{
			ID:               "repo-id",
			Key:              req.Key,
			Name:             req.Name,
			Format:           req.Format,
			RepoType:         req.RepoType,
			IsPublic:         true,
			StorageUsedBytes: 1024,
			CreatedAt:        "2026-01-01T00:00:00Z",
			UpdatedAt:        "2026-01-02T00:00:00Z",
		})
	}))
	defer server.Close()

	c, err := New(context.Background(), Config{BaseURL: server.URL, Token: "token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	public := true
	repo, err := c.CreateRepository(context.Background(), CreateRepositoryRequest{
		Key:      "docker-local",
		Name:     "Docker Local",
		Format:   "docker",
		RepoType: "local",
		IsPublic: &public,
	})
	if err != nil {
		t.Fatalf("CreateRepository returned error: %v", err)
	}
	if repo.ID != "repo-id" || repo.StorageUsedBytes != 1024 || !repo.IsPublic {
		t.Fatalf("unexpected repository: %#v", repo)
	}
}

func TestSecurityPolicyJSONMapping(t *testing.T) {
	t.Parallel()

	var created bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/security/policies":
			var req CreateSecurityPolicyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if req.Name != "strict-production" || req.MaxSeverity != "high" || !req.BlockOnFail || req.BlockUnscanned || !req.RequireSignature {
				t.Fatalf("unexpected create request: %#v", req)
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			writeJSON(t, w, SecurityPolicy{
				ID:               "policy-id",
				Name:             "strict-production",
				MaxSeverity:      "high",
				BlockOnFail:      true,
				BlockUnscanned:   false,
				Enabled:          true,
				RequireSignature: true,
				CreatedAt:        "2026-01-01T00:00:00Z",
				UpdatedAt:        "2026-01-01T00:00:00Z",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/security/policies/policy-id":
			if !created {
				t.Fatal("policy read before create")
			}
			writeJSON(t, w, SecurityPolicy{
				ID:               "policy-id",
				Name:             "strict-production",
				MaxSeverity:      "high",
				BlockOnFail:      true,
				BlockUnscanned:   false,
				Enabled:          true,
				RequireSignature: true,
				CreatedAt:        "2026-01-01T00:00:00Z",
				UpdatedAt:        "2026-01-01T00:00:00Z",
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c, err := New(context.Background(), Config{BaseURL: server.URL, Token: "token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	createdPolicy, err := c.CreateSecurityPolicy(context.Background(), CreateSecurityPolicyRequest{
		Name:             "strict-production",
		MaxSeverity:      "high",
		BlockOnFail:      true,
		BlockUnscanned:   false,
		RequireSignature: true,
	})
	if err != nil {
		t.Fatalf("CreateSecurityPolicy returned error: %v", err)
	}
	policy, err := c.GetSecurityPolicy(context.Background(), createdPolicy.ID)
	if err != nil {
		t.Fatalf("GetSecurityPolicy returned error: %v", err)
	}
	if policy.ID != "policy-id" || policy.Name != "strict-production" || !policy.Enabled || policy.MaxSeverity != "high" || !policy.RequireSignature {
		t.Fatalf("unexpected policy: %#v", policy)
	}
}

func TestLicensePolicyJSONMapping(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sbom/license-policies":
			var req UpsertLicensePolicyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if req.Name != "no-copyleft" || req.Action != "block" || req.AllowUnknown || !req.Enabled {
				t.Fatalf("unexpected upsert request: %#v", req)
			}
			if len(req.DeniedLicenses) != 2 || req.DeniedLicenses[0] != "AGPL-3.0-only" || req.DeniedLicenses[1] != "GPL-3.0-only" {
				t.Fatalf("denied licenses = %#v", req.DeniedLicenses)
			}
			writeJSON(t, w, LicensePolicy{
				ID:              "license-policy-id",
				Name:            req.Name,
				Description:     strPtr("Block copyleft"),
				AllowedLicenses: req.AllowedLicenses,
				DeniedLicenses:  req.DeniedLicenses,
				AllowUnknown:    req.AllowUnknown,
				Action:          req.Action,
				Enabled:         req.Enabled,
				CreatedAt:       "2026-01-01T00:00:00Z",
				UpdatedAt:       strPtr("2026-01-02T00:00:00Z"),
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/sbom/license-policies/license-policy-id":
			writeJSON(t, w, LicensePolicy{
				ID:              "license-policy-id",
				Name:            "no-copyleft",
				Description:     strPtr("Block copyleft"),
				AllowedLicenses: []string{},
				DeniedLicenses:  []string{"AGPL-3.0-only", "GPL-3.0-only"},
				AllowUnknown:    false,
				Action:          "block",
				Enabled:         true,
				CreatedAt:       "2026-01-01T00:00:00Z",
				UpdatedAt:       strPtr("2026-01-02T00:00:00Z"),
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c, err := New(context.Background(), Config{BaseURL: server.URL, Token: "token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	policy, err := c.UpsertLicensePolicy(context.Background(), UpsertLicensePolicyRequest{
		Name:            "no-copyleft",
		Description:     strPtr("Block copyleft"),
		AllowedLicenses: []string{},
		DeniedLicenses:  []string{"AGPL-3.0-only", "GPL-3.0-only"},
		AllowUnknown:    false,
		Action:          "block",
		Enabled:         true,
	})
	if err != nil {
		t.Fatalf("UpsertLicensePolicy returned error: %v", err)
	}
	if policy.ID != "license-policy-id" || policy.Description == nil || *policy.Description != "Block copyleft" {
		t.Fatalf("unexpected created policy: %#v", policy)
	}

	policy, err = c.GetLicensePolicy(context.Background(), policy.ID)
	if err != nil {
		t.Fatalf("GetLicensePolicy returned error: %v", err)
	}
	if policy.Name != "no-copyleft" || len(policy.DeniedLicenses) != 2 || policy.UpdatedAt == nil {
		t.Fatalf("unexpected policy: %#v", policy)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func strPtr(value string) *string {
	return &value
}
