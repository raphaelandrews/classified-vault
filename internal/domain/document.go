package domain

import "time"

type DocumentStatus string

const (
	StatusActive   DocumentStatus = "active"
	StatusArchived DocumentStatus = "archived"
	StatusRevoked  DocumentStatus = "revoked"
)

type Document struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Content        string         `json:"content"`
	Classification ClearanceLevel `json:"classification"`
	Status         DocumentStatus `json:"status"`
	Tags           []string       `json:"tags"`
	CreatedBy      string         `json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
