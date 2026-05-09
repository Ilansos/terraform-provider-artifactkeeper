package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type CreateWebhookRequest struct {
	Name         string         `json:"name"`
	URL          string         `json:"url"`
	Events       []string       `json:"events"`
	RepositoryID *string        `json:"repository_id,omitempty"`
	Secret       *string        `json:"secret,omitempty"`
	Headers      map[string]any `json:"headers,omitempty"`
}

type WebhookListOptions struct {
	Page         int
	PerPage      int
	RepositoryID string
	Enabled      *bool
}

type WebhookListResponse struct {
	Items []Webhook `json:"items"`
	Total int64     `json:"total"`
}

func (c *Client) CreateWebhook(ctx context.Context, req CreateWebhookRequest) (*Webhook, error) {
	var out Webhook
	return &out, c.do(ctx, http.MethodPost, "/webhooks", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	var out Webhook
	return &out, c.do(ctx, http.MethodGet, "/webhooks/"+url.PathEscape(id), nil, &out, http.StatusOK)
}

func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/webhooks/"+url.PathEscape(id), nil, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) EnableWebhook(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, "/webhooks/"+url.PathEscape(id)+"/enable", nil, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) DisableWebhook(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, "/webhooks/"+url.PathEscape(id)+"/disable", nil, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) ListWebhooks(ctx context.Context, opts WebhookListOptions) (*WebhookListResponse, error) {
	var out WebhookListResponse
	return &out, c.do(ctx, http.MethodGet, "/webhooks"+encodeWebhookListOptions(opts), nil, &out, http.StatusOK)
}

func encodeWebhookListOptions(opts WebhookListOptions) string {
	values := url.Values{}
	if opts.Page > 0 {
		values.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(opts.PerPage))
	}
	if opts.RepositoryID != "" {
		values.Set("repository_id", opts.RepositoryID)
	}
	if opts.Enabled != nil {
		values.Set("enabled", strconv.FormatBool(*opts.Enabled))
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
