package client

import (
	"encoding/json"
	"fmt"

	"classified-vault/internal/domain"
)

type LoginResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func (c *APIClient) Login(username, password string) (*LoginResponse, error) {
	_, body, err := c.do("POST", "/auth/login", map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	var resp LoginResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	c.Token = resp.Token
	c.User = &resp.User
	return &resp, nil
}

func (c *APIClient) Logout() error {
	_, _, err := c.do("POST", "/auth/logout", nil)
	c.Token = ""
	c.User = nil
	return err
}

func (c *APIClient) GetMe() (*domain.User, error) {
	_, body, err := c.do("GET", "/api/me", nil)
	if err != nil {
		return nil, err
	}
	var user domain.User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}
