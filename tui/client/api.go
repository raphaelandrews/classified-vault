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
	baseURL          string
	client           *http.Client
	Token            string
	User             *domain.User
	SessionExpiresAt time.Time
}

func New(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) do(method, path string, body interface{}) (*http.Response, []byte, error) {
	return c.doRequest(method, path, body)
}

func (c *APIClient) doRequest(method, path string, body interface{}) (*http.Response, []byte, error) {
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

	if expiresStr := resp.Header.Get("X-Session-Expires"); expiresStr != "" {
		if t, err := time.Parse(time.RFC3339, expiresStr); err == nil {
			c.SessionExpiresAt = t
		}
	}

	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		msg := "too many login attempts"
		if retryAfter != "" {
			msg = fmt.Sprintf("too many login attempts, wait %s seconds", retryAfter)
		}
		return resp, nil, fmt.Errorf("%s", msg)
	}

	if resp.StatusCode >= 400 {
		var apiErr map[string]string
		json.Unmarshal(respBody, &apiErr)
		msg := apiErr["error"]
		if msg == "" {
			msg = fmt.Sprintf("request failed (status %d)", resp.StatusCode)
		}
		return resp, nil, fmt.Errorf("%s", msg)
	}

	return resp, respBody, nil
}
