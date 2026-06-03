package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"classified-vault/internal/domain"
)

type APIClient struct {
	baseURL string
	client  *http.Client
	Token   string
	User    *domain.User
}

func New(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) do(method, path string, body interface{}) (*http.Response, []byte, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	return resp, respBody, nil
}

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

func (c *APIClient) ListDocuments() ([]*domain.Document, error) {
	_, body, err := c.do("GET", "/api/documents", nil)
	if err != nil {
		return nil, err
	}
	var docs []*domain.Document
	if err := json.Unmarshal(body, &docs); err != nil {
		return nil, fmt.Errorf("unmarshal documents: %w", err)
	}
	return docs, nil
}

func (c *APIClient) GetDocument(id string) (*domain.Document, error) {
	_, body, err := c.do("GET", "/api/documents/"+id, nil)
	if err != nil {
		return nil, err
	}
	var doc domain.Document
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal document: %w", err)
	}
	return &doc, nil
}

func (c *APIClient) CreateDocument(title, content string, classification domain.ClearanceLevel, tags []string) (*domain.Document, error) {
	doc := domain.Document{
		Title:          title,
		Content:        content,
		Classification: classification,
		Tags:           tags,
	}
	_, body, err := c.do("POST", "/api/documents", doc)
	if err != nil {
		return nil, err
	}
	var created domain.Document
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("unmarshal document: %w", err)
	}
	return &created, nil
}

func (c *APIClient) UpdateDocument(id, title, content string, classification domain.ClearanceLevel, tags []string) (*domain.Document, error) {
	doc := domain.Document{
		Title:          title,
		Content:        content,
		Classification: classification,
		Tags:           tags,
	}
	_, body, err := c.do("PUT", "/api/documents/"+id, doc)
	if err != nil {
		return nil, err
	}
	var updated domain.Document
	if err := json.Unmarshal(body, &updated); err != nil {
		return nil, fmt.Errorf("unmarshal document: %w", err)
	}
	return &updated, nil
}

func (c *APIClient) DeleteDocument(id string) error {
	_, _, err := c.do("DELETE", "/api/documents/"+id, nil)
	return err
}

func (c *APIClient) ListUsers() ([]*domain.User, error) {
	_, body, err := c.do("GET", "/api/users", nil)
	if err != nil {
		return nil, err
	}
	var users []*domain.User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("unmarshal users: %w", err)
	}
	return users, nil
}

func (c *APIClient) CreateUser(username, email, password string, role domain.Role) (*domain.User, error) {
	u := domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: password,
		Role:         role,
	}
	_, body, err := c.do("POST", "/api/users", u)
	if err != nil {
		return nil, err
	}
	var created domain.User
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &created, nil
}

func (c *APIClient) UpdateUser(id, username, email, password string, role domain.Role, clearance domain.ClearanceLevel) (*domain.User, error) {
	u := domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: password,
		Role:         role,
		Clearance:    clearance,
	}
	_, body, err := c.do("PUT", "/api/users/"+id, u)
	if err != nil {
		return nil, err
	}
	var updated domain.User
	if err := json.Unmarshal(body, &updated); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &updated, nil
}

func (c *APIClient) DeleteUser(id string) error {
	_, _, err := c.do("DELETE", "/api/users/"+id, nil)
	return err
}

func (c *APIClient) ListAuditLogs() ([]*domain.AuditLog, error) {
	_, body, err := c.do("GET", "/api/audit", nil)
	if err != nil {
		return nil, err
	}
	var logs []*domain.AuditLog
	if err := json.Unmarshal(body, &logs); err != nil {
		return nil, fmt.Errorf("unmarshal audit logs: %w", err)
	}
	return logs, nil
}
