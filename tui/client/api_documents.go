package client

import (
	"encoding/json"
	"fmt"

	"classified-vault/internal/domain"
)

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

func (c *APIClient) CreateDocument(title, content string, classification domain.ClearanceLevel, department domain.Department, tags []string) (*domain.Document, error) {
	doc := domain.Document{
		Title:          title,
		Content:        content,
		Classification: classification,
		Department:        department,
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

func (c *APIClient) UpdateDocument(id, title, content string, classification domain.ClearanceLevel, department domain.Department, tags []string) (*domain.Document, error) {
	doc := domain.Document{
		Title:          title,
		Content:        content,
		Classification: classification,
		Department:        department,
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

type CatalogEntry struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Classification int      `json:"classification"`
	Status         string   `json:"status"`
	Department        string   `json:"department"`
	Folder         string   `json:"folder"`
	Tags           []string `json:"tags"`
	CreatedBy      string   `json:"created_by"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

func (c *APIClient) ListCatalog() ([]CatalogEntry, error) {
	_, body, err := c.do("GET", "/api/catalog", nil)
	if err != nil {
		return nil, err
	}
	var entries []CatalogEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal catalog: %w", err)
	}
	return entries, nil
}
