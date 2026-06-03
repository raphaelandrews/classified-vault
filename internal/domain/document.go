package domain

import "time"

type DocumentStatus string

const (
	StatusActive   DocumentStatus = "active"
	StatusArchived DocumentStatus = "archived"
	StatusSealed   DocumentStatus = "sealed"
)

type Document struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Content        string         `json:"content"`
	Classification ClearanceLevel `json:"tier"`
	Status         DocumentStatus `json:"status"`
	Department        Department        `json:"department"`
	Folder         string         `json:"folder,omitempty"`
	Tags           []string       `json:"tags"`
	ReferenceIDs   []string       `json:"references,omitempty"`
	CreatedBy      string         `json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
