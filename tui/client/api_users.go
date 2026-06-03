package client

import (
	"encoding/json"
	"fmt"

	"classified-vault/internal/domain"
)

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

func (c *APIClient) CreateUser(username, email, password string, role domain.Role, faction domain.Faction) (*domain.User, error) {
	u := domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: password,
		Role:         role,
		Faction:      faction,
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

func (c *APIClient) UpdateUser(id, username, email, password string, role domain.Role, clearance domain.ClearanceLevel, faction domain.Faction) (*domain.User, error) {
	u := domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: password,
		Role:         role,
		Clearance:    clearance,
		Faction:      faction,
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
