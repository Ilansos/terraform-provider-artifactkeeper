package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type CreateGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type MembersRequest struct {
	UserIDs []string `json:"user_ids"`
}

type groupDetailResponse struct {
	Group
	Members      []GroupMember `json:"members"`
	MembersTotal int64         `json:"members_total"`
}

type GroupListOptions struct {
	Page    int
	PerPage int
	Search  string
}

type GroupListResponse struct {
	Items      []Group    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

func (c *Client) CreateGroup(ctx context.Context, req CreateGroupRequest) (*Group, error) {
	var out Group
	return &out, c.do(ctx, http.MethodPost, "/groups", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) GetGroup(ctx context.Context, id string) (*Group, error) {
	var out groupDetailResponse
	if err := c.do(ctx, http.MethodGet, "/groups/"+url.PathEscape(id), nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	group := out.Group
	group.Members = out.Members
	return &group, nil
}

func (c *Client) UpdateGroup(ctx context.Context, id string, req CreateGroupRequest) (*Group, error) {
	var out Group
	return &out, c.do(ctx, http.MethodPut, "/groups/"+url.PathEscape(id), req, &out, http.StatusOK)
}

func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/groups/"+url.PathEscape(id), nil, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) AddGroupMembers(ctx context.Context, id string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	return c.do(ctx, http.MethodPost, "/groups/"+url.PathEscape(id)+"/members", MembersRequest{UserIDs: userIDs}, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) RemoveGroupMembers(ctx context.Context, id string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	return c.do(ctx, http.MethodDelete, "/groups/"+url.PathEscape(id)+"/members", MembersRequest{UserIDs: userIDs}, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) ListGroups(ctx context.Context, opts GroupListOptions) (*GroupListResponse, error) {
	var out GroupListResponse
	return &out, c.do(ctx, http.MethodGet, "/groups"+encodeGroupListOptions(opts), nil, &out, http.StatusOK)
}

func encodeGroupListOptions(opts GroupListOptions) string {
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
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
