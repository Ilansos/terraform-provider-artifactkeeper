package client

import (
	"context"
	"net/http"
	"net/url"
)

type LicensePolicy struct {
	ID              string   `json:"id"`
	RepositoryID    *string  `json:"repository_id"`
	Name            string   `json:"name"`
	Description     *string  `json:"description"`
	AllowedLicenses []string `json:"allowed_licenses"`
	DeniedLicenses  []string `json:"denied_licenses"`
	AllowUnknown    bool     `json:"allow_unknown"`
	Action          string   `json:"action"`
	Enabled         bool     `json:"is_enabled"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       *string  `json:"updated_at"`
}

type UpsertLicensePolicyRequest struct {
	RepositoryID    *string  `json:"repository_id,omitempty"`
	Name            string   `json:"name"`
	Description     *string  `json:"description,omitempty"`
	AllowedLicenses []string `json:"allowed_licenses"`
	DeniedLicenses  []string `json:"denied_licenses"`
	AllowUnknown    bool     `json:"allow_unknown"`
	Action          string   `json:"action"`
	Enabled         bool     `json:"is_enabled"`
}

func (c *Client) ListLicensePolicies(ctx context.Context) ([]LicensePolicy, error) {
	var out []LicensePolicy
	return out, c.do(ctx, http.MethodGet, "/sbom/license-policies", nil, &out, http.StatusOK)
}

func (c *Client) UpsertLicensePolicy(ctx context.Context, req UpsertLicensePolicyRequest) (*LicensePolicy, error) {
	var out LicensePolicy
	return &out, c.do(ctx, http.MethodPost, "/sbom/license-policies", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) GetLicensePolicy(ctx context.Context, id string) (*LicensePolicy, error) {
	var out LicensePolicy
	return &out, c.do(ctx, http.MethodGet, "/sbom/license-policies/"+url.PathEscape(id), nil, &out, http.StatusOK)
}

func (c *Client) DeleteLicensePolicy(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/sbom/license-policies/"+url.PathEscape(id), nil, nil, http.StatusOK, http.StatusNoContent)
}
