package client

import (
	"context"
	"net/http"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	ExpiresIn          int64  `json:"expires_in"`
	TokenType          string `json:"token_type"`
	MustChangePassword bool   `json:"must_change_password"`
}

func (c *Client) Login(ctx context.Context, username, password string) error {
	var out loginResponse
	if err := c.do(ctx, http.MethodPost, "/auth/login", loginRequest{
		Username: username,
		Password: password,
	}, &out, http.StatusOK); err != nil {
		return err
	}
	c.token = out.AccessToken
	return nil
}
