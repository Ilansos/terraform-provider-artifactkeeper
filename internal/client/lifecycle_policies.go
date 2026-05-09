package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type LifecyclePolicy struct {
	ID                  string          `json:"id"`
	RepositoryID        *string         `json:"repository_id"`
	Name                string          `json:"name"`
	Description         *string         `json:"description"`
	Enabled             bool            `json:"enabled"`
	PolicyType          string          `json:"policy_type"`
	Config              json.RawMessage `json:"config"`
	Priority            int             `json:"priority"`
	CronSchedule        *string         `json:"cron_schedule"`
	LastRunAt           *string         `json:"last_run_at"`
	LastRunItemsRemoved *int64          `json:"last_run_items_removed"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
}

type CreateLifecyclePolicyRequest struct {
	RepositoryID *string         `json:"repository_id,omitempty"`
	Name         string          `json:"name"`
	Description  *string         `json:"description,omitempty"`
	Enabled      *bool           `json:"enabled,omitempty"`
	PolicyType   string          `json:"policy_type"`
	Config       json.RawMessage `json:"config"`
	Priority     int             `json:"priority"`
}

type UpdateLifecyclePolicyRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Enabled     bool            `json:"enabled"`
	Config      json.RawMessage `json:"config"`
	Priority    int             `json:"priority"`
}

func (c *Client) CreateLifecyclePolicy(ctx context.Context, req CreateLifecyclePolicyRequest) (*LifecyclePolicy, error) {
	var out LifecyclePolicy
	return &out, c.do(ctx, http.MethodPost, "/admin/lifecycle", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) GetLifecyclePolicy(ctx context.Context, id string) (*LifecyclePolicy, error) {
	var out LifecyclePolicy
	return &out, c.do(ctx, http.MethodGet, "/admin/lifecycle/"+url.PathEscape(id), nil, &out, http.StatusOK)
}

func (c *Client) UpdateLifecyclePolicy(ctx context.Context, id string, req UpdateLifecyclePolicyRequest) (*LifecyclePolicy, error) {
	var out LifecyclePolicy
	return &out, c.do(ctx, http.MethodPatch, "/admin/lifecycle/"+url.PathEscape(id), req, &out, http.StatusOK)
}

func (c *Client) DeleteLifecyclePolicy(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/admin/lifecycle/"+url.PathEscape(id), nil, nil, http.StatusOK, http.StatusNoContent)
}
