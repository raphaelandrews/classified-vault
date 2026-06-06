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
		Department:     department,
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
		Department:     department,
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

func (c *APIClient) TransitionDocument(id, status string) (*domain.Document, error) {
	_, body, err := c.do("PUT", "/api/documents/"+id+"/transition", map[string]string{"status": status})
	if err != nil {
		return nil, err
	}
	var doc domain.Document
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal document: %w", err)
	}
	return &doc, nil
}

type CatalogEntry struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Classification int      `json:"classification"`
	Status         string   `json:"status"`
	Department     string   `json:"department"`
	Folder         string   `json:"folder"`
	Tags           []string `json:"tags"`
	CreatedBy      string   `json:"created_by"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

func (c *APIClient) ListCatalog() ([]CatalogEntry, error) {
	return c.ListCatalogPage(1000, 0)
}

func (c *APIClient) ListCatalogPage(limit, offset int) ([]CatalogEntry, error) {
	_, body, err := c.do("GET", fmt.Sprintf("/api/catalog?limit=%d&offset=%d", limit, offset), nil)
	if err != nil {
		return nil, err
	}
	var entries []CatalogEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal catalog: %w", err)
	}
	return entries, nil
}

func (c *APIClient) CountDocuments() (int, error) {
	resp, _, err := c.do("GET", "/api/catalog?limit=1&offset=0", nil)
	if err != nil {
		return 0, err
	}
	total := resp.Header.Get("X-Total-Count")
	if total == "" {
		return 0, nil
	}
	var count int
	fmt.Sscanf(total, "%d", &count)
	return count, nil
}

func (c *APIClient) SearchDocuments(query string) ([]CatalogEntry, error) {
	_, body, err := c.do("GET", "/api/documents/search?q="+query, nil)
	if err != nil {
		return nil, err
	}
	var entries []CatalogEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal search results: %w", err)
	}
	return entries, nil
}

func (c *APIClient) ExportDocument(id string) (string, error) {
	resp, body, err := c.do("GET", "/api/documents/"+id+"/export?format=md", nil)
	if err != nil {
		return "", err
	}
	_ = resp
	return string(body), nil
}

type StatsResponse struct {
	TierCounts       map[string]int `json:"tier_counts"`
	DepartmentCounts map[string]int `json:"department_counts"`
	MostActive       string         `json:"most_active"`
	MostActiveCount  int            `json:"most_active_count"`
	CreatedThisMonth int            `json:"created_this_month"`
	TotalScrolls     int            `json:"total_scrolls"`
	TotalVillagers   int            `json:"total_villagers"`
}

func (c *APIClient) GetStats() (*StatsResponse, error) {
	_, body, err := c.do("GET", "/api/stats", nil)
	if err != nil {
		return nil, err
	}
	var stats StatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	return &stats, nil
}
