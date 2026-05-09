package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type CreatePermissionRequest struct {
	PrincipalType string   `json:"principal_type"`
	PrincipalID   string   `json:"principal_id"`
	TargetType    string   `json:"target_type"`
	TargetID      string   `json:"target_id"`
	Actions       []string `json:"actions"`
}

type PermissionListOptions struct {
	Page          int
	PerPage       int
	PrincipalType string
	PrincipalID   string
	TargetType    string
	TargetID      string
}

type PermissionListResponse struct {
	Items      []Permission `json:"items"`
	Pagination Pagination   `json:"pagination"`
}

func (c *Client) CreatePermission(ctx context.Context, req CreatePermissionRequest) (*Permission, error) {
	var out Permission
	return &out, c.do(ctx, http.MethodPost, "/permissions", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) GetPermission(ctx context.Context, id string) (*Permission, error) {
	var out Permission
	return &out, c.do(ctx, http.MethodGet, "/permissions/"+url.PathEscape(id), nil, &out, http.StatusOK)
}

func (c *Client) UpdatePermission(ctx context.Context, id string, req CreatePermissionRequest) (*Permission, error) {
	var out Permission
	return &out, c.do(ctx, http.MethodPut, "/permissions/"+url.PathEscape(id), req, &out, http.StatusOK)
}

func (c *Client) DeletePermission(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/permissions/"+url.PathEscape(id), nil, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) ListPermissions(ctx context.Context, opts PermissionListOptions) (*PermissionListResponse, error) {
	var out PermissionListResponse
	return &out, c.do(ctx, http.MethodGet, "/permissions"+encodePermissionListOptions(opts), nil, &out, http.StatusOK)
}

func encodePermissionListOptions(opts PermissionListOptions) string {
	values := url.Values{}
	if opts.Page > 0 {
		values.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(opts.PerPage))
	}
	if opts.PrincipalType != "" {
		values.Set("principal_type", opts.PrincipalType)
	}
	if opts.PrincipalID != "" {
		values.Set("principal_id", opts.PrincipalID)
	}
	if opts.TargetType != "" {
		values.Set("target_type", opts.TargetType)
	}
	if opts.TargetID != "" {
		values.Set("target_id", opts.TargetID)
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
