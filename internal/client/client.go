package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

type Config struct {
	BaseURL            string
	Username           string
	Password           string
	Token              string
	InsecureSkipVerify bool
	HTTPClient         *http.Client
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      string
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	baseURL, err := NormalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		if cfg.InsecureSkipVerify {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Explicit lab/dev provider option.
		}

		httpClient = &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		}
	}

	c := &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		token:      cfg.Token,
	}

	if c.token == "" && cfg.Username != "" && cfg.Password != "" {
		if err := c.Login(ctx, cfg.Username, cfg.Password); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func NormalizeBaseURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("artifact keeper url is required")
	}

	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("parse artifact keeper url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("artifact keeper url must include scheme and host")
	}

	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "" {
		u.Path = "/api/v1"
	} else if !strings.HasSuffix(u.Path, "/api/v1") {
		u.Path = strings.TrimRight(u.Path, "/") + "/api/v1"
	}

	return u, nil
}

func (c *Client) do(ctx context.Context, method, endpoint string, body any, out any, expected ...int) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	reqURL := *c.baseURL
	rel, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse request endpoint: %w", err)
	}
	reqURL.Path = path.Join(c.baseURL.Path, rel.Path)
	reqURL.RawQuery = rel.RawQuery
	if strings.HasSuffix(rel.Path, "/") {
		reqURL.Path += "/"
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	if statusExpected(res.StatusCode, expected) {
		if out == nil || res.StatusCode == http.StatusNoContent {
			io.Copy(io.Discard, res.Body)
			return nil
		}
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		return nil
	}

	return decodeAPIError(res)
}

func statusExpected(status int, expected []int) bool {
	for _, code := range expected {
		if status == code {
			return true
		}
	}
	return false
}
