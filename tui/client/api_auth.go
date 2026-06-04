package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"classified-vault/internal/domain"
)

type LoginResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func (c *APIClient) Login(username, password string) (*LoginResponse, error) {
	resp, body, err := c.do("POST", "/auth/login", map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr map[string]string
		json.Unmarshal(body, &apiErr)
		msg := apiErr["error"]
		if msg == "" {
			msg = fmt.Sprintf("login failed (status %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", msg)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	c.Token = loginResp.Token
	c.User = &loginResp.User
	return &loginResp, nil
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

func (c *APIClient) Register(username, password, department string) (*domain.User, error) {
	resp, body, err := c.do("POST", "/auth/register", map[string]string{
		"username":   username,
		"password":   password,
		"department": department,
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		var apiErr map[string]string
		json.Unmarshal(body, &apiErr)
		msg := apiErr["error"]
		if msg == "" {
			msg = fmt.Sprintf("register failed (status %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", msg)
	}

	var user domain.User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}
