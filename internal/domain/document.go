package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type DocumentStatus string

const (
	StatusDraft    DocumentStatus = "draft"
	StatusReview   DocumentStatus = "review"
	StatusFrozen   DocumentStatus = "frozen"
	StatusArchived DocumentStatus = "archived"
	StatusPublic   DocumentStatus = "public"
)

type Document struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Content        string         `json:"content"`
	Classification ClearanceLevel `json:"tier"`
	Status         DocumentStatus `json:"status"`
	Department     Department     `json:"department"`
	Folder         string         `json:"folder,omitempty"`
	Tags           []string       `json:"tags"`
	ReferenceIDs   []string       `json:"references,omitempty"`
	ContentHash    string         `json:"content_hash,omitempty"`
	CreatedBy      string         `json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (d *Document) ComputeHash() string {
	payload := struct {
		Title          string         `json:"title"`
		Content        string         `json:"content"`
		Classification ClearanceLevel `json:"tier"`
		Department     Department     `json:"department"`
		Tags           []string       `json:"tags"`
	}{
		Title:          d.Title,
		Content:        d.Content,
		Classification: d.Classification,
		Department:     d.Department,
		Tags:           d.Tags,
	}
	data, _ := json.Marshal(payload)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (d *Document) VerifyIntegrity() bool {
	if d.ContentHash == "" {
		return true
	}
	return d.ComputeHash() == d.ContentHash
}

var workflowTransitions = map[DocumentStatus][]DocumentStatus{
	StatusDraft:    {StatusReview},
	StatusReview:   {StatusFrozen, StatusDraft},
	StatusFrozen:   {StatusArchived, StatusDraft},
	StatusArchived: {StatusPublic, StatusDraft},
	StatusPublic:   {StatusDraft},
}

func CanTransition(from, to DocumentStatus) bool {
	targets, ok := workflowTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

func TransitionRequiresMayor(to DocumentStatus) bool {
	return to == StatusFrozen || to == StatusPublic
}
