package client

import (
	"encoding/json"
	"fmt"

	"classified-vault/internal/domain"
)

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
