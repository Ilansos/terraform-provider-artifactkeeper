package client

import (
	"context"
	"net/http"
	"net/url"
)

type SecurityPolicy struct {
	ID                 string  `json:"id"`
	RepositoryID       *string `json:"repository_id"`
	Name               string  `json:"name"`
	MaxSeverity        string  `json:"max_severity"`
	BlockUnscanned     bool    `json:"block_unscanned"`
	BlockOnFail        bool    `json:"block_on_fail"`
	Enabled            bool    `json:"is_enabled"`
	RequireSignature   bool    `json:"require_signature"`
	MaxArtifactAgeDays *int    `json:"max_artifact_age_days"`
	MinStagingHours    *int    `json:"min_staging_hours"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type CreateSecurityPolicyRequest struct {
	RepositoryID       *string `json:"repository_id,omitempty"`
	Name               string  `json:"name"`
	MaxSeverity        string  `json:"max_severity"`
	BlockUnscanned     bool    `json:"block_unscanned"`
	BlockOnFail        bool    `json:"block_on_fail"`
	RequireSignature   bool    `json:"require_signature"`
	MaxArtifactAgeDays *int    `json:"max_artifact_age_days,omitempty"`
	MinStagingHours    *int    `json:"min_staging_hours,omitempty"`
}

type UpdateSecurityPolicyRequest struct {
	Name               string `json:"name"`
	MaxSeverity        string `json:"max_severity"`
	BlockUnscanned     bool   `json:"block_unscanned"`
	BlockOnFail        bool   `json:"block_on_fail"`
	Enabled            bool   `json:"is_enabled"`
	RequireSignature   bool   `json:"require_signature"`
	MaxArtifactAgeDays *int   `json:"max_artifact_age_days"`
	MinStagingHours    *int   `json:"min_staging_hours"`
}

func (c *Client) CreateSecurityPolicy(ctx context.Context, req CreateSecurityPolicyRequest) (*SecurityPolicy, error) {
	var out SecurityPolicy
	return &out, c.do(ctx, http.MethodPost, "/security/policies", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) GetSecurityPolicy(ctx context.Context, id string) (*SecurityPolicy, error) {
	var out SecurityPolicy
	return &out, c.do(ctx, http.MethodGet, "/security/policies/"+url.PathEscape(id), nil, &out, http.StatusOK)
}

func (c *Client) UpdateSecurityPolicy(ctx context.Context, id string, req UpdateSecurityPolicyRequest) (*SecurityPolicy, error) {
	var out SecurityPolicy
	return &out, c.do(ctx, http.MethodPut, "/security/policies/"+url.PathEscape(id), req, &out, http.StatusOK)
}

func (c *Client) DeleteSecurityPolicy(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/security/policies/"+url.PathEscape(id), nil, nil, http.StatusOK, http.StatusNoContent)
}
