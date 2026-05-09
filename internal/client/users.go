package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type CreateUserRequest struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password *string `json:"password,omitempty"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
}

type createUserResponse struct {
	User User `json:"user"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
}

type UserListOptions struct {
	Page     int
	PerPage  int
	Search   string
	IsActive *bool
	IsAdmin  *bool
}

type UserListResponse struct {
	Items      []User     `json:"items"`
	Pagination Pagination `json:"pagination"`
}

func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	var out createUserResponse
	if err := c.do(ctx, http.MethodPost, "/users", req, &out, http.StatusOK, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out.User, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var out User
	return &out, c.do(ctx, http.MethodGet, "/users/"+url.PathEscape(id), nil, &out, http.StatusOK)
}

func (c *Client) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	var out User
	return &out, c.do(ctx, http.MethodPatch, "/users/"+url.PathEscape(id), req, &out, http.StatusOK)
}

func (c *Client) DeleteUser(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/users/"+url.PathEscape(id), nil, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) ListUsers(ctx context.Context, opts UserListOptions) (*UserListResponse, error) {
	var out UserListResponse
	return &out, c.do(ctx, http.MethodGet, "/users"+encodeUserListOptions(opts), nil, &out, http.StatusOK)
}

func encodeUserListOptions(opts UserListOptions) string {
	values := url.Values{}
	if opts.Page > 0 {
		values.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(opts.PerPage))
	}
	if opts.Search != "" {
		values.Set("search", opts.Search)
	}
	if opts.IsActive != nil {
		values.Set("is_active", strconv.FormatBool(*opts.IsActive))
	}
	if opts.IsAdmin != nil {
		values.Set("is_admin", strconv.FormatBool(*opts.IsAdmin))
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
