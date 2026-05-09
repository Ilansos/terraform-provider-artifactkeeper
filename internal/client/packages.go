package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type PackageListOptions struct {
	Page          int
	PerPage       int
	RepositoryKey string
	Format        string
	Search        string
}

type PackageListResponse struct {
	Items      []Package  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PackageVersionsResponse struct {
	Versions []PackageVersion `json:"versions"`
}

func (c *Client) GetPackage(ctx context.Context, id string) (*Package, error) {
	var out Package
	return &out, c.do(ctx, http.MethodGet, "/packages/"+url.PathEscape(id), nil, &out, http.StatusOK)
}

func (c *Client) ListPackages(ctx context.Context, opts PackageListOptions) (*PackageListResponse, error) {
	var out PackageListResponse
	return &out, c.do(ctx, http.MethodGet, "/packages"+encodePackageListOptions(opts), nil, &out, http.StatusOK)
}

func (c *Client) GetPackageVersions(ctx context.Context, id string) (*PackageVersionsResponse, error) {
	var out PackageVersionsResponse
	return &out, c.do(ctx, http.MethodGet, "/packages/"+url.PathEscape(id)+"/versions", nil, &out, http.StatusOK)
}

func encodePackageListOptions(opts PackageListOptions) string {
	values := url.Values{}
	if opts.Page > 0 {
		values.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(opts.PerPage))
	}
	if opts.RepositoryKey != "" {
		values.Set("repository_key", opts.RepositoryKey)
	}
	if opts.Format != "" {
		values.Set("format", opts.Format)
	}
	if opts.Search != "" {
		values.Set("search", opts.Search)
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
