package repository

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"classified-vault/internal/domain"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) FindAll() ([]*domain.Document, error) {
	rows, err := r.db.Query(
		`SELECT id, title, content, classification, status, tags, created_by, created_at, updated_at
		 FROM documents ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var d domain.Document
		var tagsJSON string
		if err := rows.Scan(&d.ID, &d.Title, &d.Content, &d.Classification, &d.Status, &tagsJSON, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &d.Tags)
		if d.Tags == nil {
			d.Tags = []string{}
		}
		docs = append(docs, &d)
	}
	return docs, rows.Err()
}

func (r *DocumentRepository) FindByIDs(ids []string) ([]*domain.Document, error) {
	if len(ids) == 0 {
		return []*domain.Document{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id, title, content, classification, status, tags, created_by, created_at, updated_at
		 FROM documents WHERE id IN (` + strings.Join(placeholders, ",") + `) ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var d domain.Document
		var tagsJSON string
		if err := rows.Scan(&d.ID, &d.Title, &d.Content, &d.Classification, &d.Status, &tagsJSON, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &d.Tags)
		if d.Tags == nil {
			d.Tags = []string{}
		}
		docs = append(docs, &d)
	}
	return docs, rows.Err()
}

func (r *DocumentRepository) FindByID(id string) (*domain.Document, error) {
	var d domain.Document
	var tagsJSON string
	err := r.db.QueryRow(
		`SELECT id, title, content, classification, status, tags, created_by, created_at, updated_at
		 FROM documents WHERE id = ?`, id,
	).Scan(&d.ID, &d.Title, &d.Content, &d.Classification, &d.Status, &tagsJSON, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(tagsJSON), &d.Tags)
	if d.Tags == nil {
		d.Tags = []string{}
	}
	return &d, nil
}

func (r *DocumentRepository) Create(d *domain.Document) error {
	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now

	tagsJSON, _ := json.Marshal(d.Tags)
	if string(tagsJSON) == "null" {
		tagsJSON = []byte("[]")
	}

	_, err := r.db.Exec(
		`INSERT INTO documents (id, title, content, classification, status, tags, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.Title, d.Content, d.Classification, d.Status, string(tagsJSON), d.CreatedBy, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DocumentRepository) Update(d *domain.Document) error {
	d.UpdatedAt = time.Now()

	tagsJSON, _ := json.Marshal(d.Tags)
	if string(tagsJSON) == "null" {
		tagsJSON = []byte("[]")
	}

	_, err := r.db.Exec(
		`UPDATE documents SET title=?, content=?, classification=?, status=?, tags=?, updated_at=?
		 WHERE id=?`,
		d.Title, d.Content, d.Classification, d.Status, string(tagsJSON), d.UpdatedAt, d.ID,
	)
	return err
}

func (r *DocumentRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM documents WHERE id = ?`, id)
	return err
}

type DocMetadata struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Classification int       `json:"classification"`
	Status         string    `json:"status"`
	Tags           []string  `json:"tags"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (r *DocumentRepository) FindAllMetadata() ([]DocMetadata, error) {
	rows, err := r.db.Query(
		`SELECT id, title, classification, status, tags, created_by, created_at, updated_at
		 FROM documents ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []DocMetadata
	for rows.Next() {
		var d DocMetadata
		var tagsJSON string
		if err := rows.Scan(&d.ID, &d.Title, &d.Classification, &d.Status, &tagsJSON, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &d.Tags)
		if d.Tags == nil {
			d.Tags = []string{}
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}
