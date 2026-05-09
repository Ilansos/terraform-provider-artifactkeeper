package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type CreateRepositoryRequest struct {
	Key              string                     `json:"key"`
	Name             string                     `json:"name"`
	Format           string                     `json:"format"`
	RepoType         string                     `json:"repo_type"`
	Description      *string                    `json:"description,omitempty"`
	IsPublic         *bool                      `json:"is_public,omitempty"`
	QuotaBytes       *int64                     `json:"quota_bytes,omitempty"`
	MemberRepos      []CreateVirtualMemberInput `json:"member_repos,omitempty"`
	UpstreamURL      *string                    `json:"upstream_url,omitempty"`
	IndexUpstreamURL *string                    `json:"index_upstream_url,omitempty"`
	UpstreamAuthType *string                    `json:"upstream_auth_type,omitempty"`
	UpstreamUsername *string                    `json:"upstream_username,omitempty"`
	UpstreamPassword *string                    `json:"upstream_password,omitempty"`
}

type CreateVirtualMemberInput struct {
	RepoKey  string `json:"repo_key"`
	Priority int    `json:"priority,omitempty"`
}

type VirtualMemberPriority struct {
	MemberKey string `json:"member_key"`
	Priority  int    `json:"priority"`
}

type VirtualMember struct {
	ID             string `json:"id"`
	MemberRepoID   string `json:"member_repo_id"`
	MemberRepoKey  string `json:"member_repo_key"`
	MemberRepoName string `json:"member_repo_name"`
	MemberRepoType string `json:"member_repo_type"`
	Priority       int    `json:"priority"`
	CreatedAt      string `json:"created_at"`
}

type VirtualMembersListResponse struct {
	Members []VirtualMember `json:"members"`
}

type UpdateVirtualMembersRequest struct {
	Members []VirtualMemberPriority `json:"members"`
}

type RepositorySecurity struct {
	Config *ScanConfig `json:"config"`
}

type ScanConfig struct {
	ID                     string `json:"id"`
	RepositoryID           string `json:"repository_id"`
	ScanEnabled            bool   `json:"scan_enabled"`
	ScanOnUpload           bool   `json:"scan_on_upload"`
	ScanOnProxy            bool   `json:"scan_on_proxy"`
	BlockOnPolicyViolation bool   `json:"block_on_policy_violation"`
	SeverityThreshold      string `json:"severity_threshold"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

type UpsertScanConfigRequest struct {
	ScanEnabled            bool   `json:"scan_enabled"`
	ScanOnUpload           bool   `json:"scan_on_upload"`
	ScanOnProxy            bool   `json:"scan_on_proxy"`
	BlockOnPolicyViolation bool   `json:"block_on_policy_violation"`
	SeverityThreshold      string `json:"severity_threshold"`
}

type UpdateRepositoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
	QuotaBytes  *int64  `json:"quota_bytes,omitempty"`
}

type RepositoryListOptions struct {
	Page    int
	PerPage int
	Format  string
	Type    string
	Query   string
}

type RepositoryListResponse struct {
	Items      []Repository `json:"items"`
	Pagination Pagination   `json:"pagination"`
}

func (c *Client) CreateRepository(ctx context.Context, req CreateRepositoryRequest) (*Repository, error) {
	var out Repository
	return &out, c.do(ctx, http.MethodPost, "/repositories", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) GetRepository(ctx context.Context, key string) (*Repository, error) {
	var out Repository
	return &out, c.do(ctx, http.MethodGet, "/repositories/"+url.PathEscape(key), nil, &out, http.StatusOK)
}

func (c *Client) UpdateRepository(ctx context.Context, key string, req UpdateRepositoryRequest) (*Repository, error) {
	var out Repository
	return &out, c.do(ctx, http.MethodPatch, "/repositories/"+url.PathEscape(key), req, &out, http.StatusOK)
}

func (c *Client) DeleteRepository(ctx context.Context, key string) error {
	return c.do(ctx, http.MethodDelete, "/repositories/"+url.PathEscape(key), nil, nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) GetVirtualMembers(ctx context.Context, key string) (*VirtualMembersListResponse, error) {
	var out VirtualMembersListResponse
	return &out, c.do(ctx, http.MethodGet, "/repositories/"+url.PathEscape(key)+"/members", nil, &out, http.StatusOK)
}

func (c *Client) UpdateVirtualMembers(ctx context.Context, key string, members []VirtualMemberPriority) (*VirtualMembersListResponse, error) {
	var out VirtualMembersListResponse
	return &out, c.do(ctx, http.MethodPut, "/repositories/"+url.PathEscape(key)+"/members", UpdateVirtualMembersRequest{Members: members}, &out, http.StatusOK)
}

func (c *Client) GetRepositorySecurity(ctx context.Context, key string) (*RepositorySecurity, error) {
	var out RepositorySecurity
	return &out, c.do(ctx, http.MethodGet, "/repositories/"+url.PathEscape(key)+"/security", nil, &out, http.StatusOK)
}

func (c *Client) UpsertRepositorySecurity(ctx context.Context, key string, req UpsertScanConfigRequest) (*ScanConfig, error) {
	var out ScanConfig
	return &out, c.do(ctx, http.MethodPut, "/repositories/"+url.PathEscape(key)+"/security", req, &out, http.StatusOK, http.StatusCreated)
}

func (c *Client) ListRepositories(ctx context.Context, opts RepositoryListOptions) (*RepositoryListResponse, error) {
	var out RepositoryListResponse
	return &out, c.do(ctx, http.MethodGet, "/repositories"+encodeRepositoryListOptions(opts), nil, &out, http.StatusOK)
}

func encodeRepositoryListOptions(opts RepositoryListOptions) string {
	values := url.Values{}
	if opts.Page > 0 {
		values.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(opts.PerPage))
	}
	if opts.Format != "" {
		values.Set("format", opts.Format)
	}
	if opts.Type != "" {
		values.Set("type", opts.Type)
	}
	if opts.Query != "" {
		values.Set("q", opts.Query)
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
