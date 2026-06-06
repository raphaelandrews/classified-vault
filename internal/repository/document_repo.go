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

func scanDoc(rows *sql.Rows) (*domain.Document, error) {
	var d domain.Document
	var tagsJSON string
	var department string
	var folder, refIDs sql.NullString
	if err := rows.Scan(&d.ID, &d.Title, &d.Content, &d.Classification, &d.Status, &department, &folder, &tagsJSON, &refIDs, &d.ContentHash, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, err
	}
	d.Department = domain.Department(department)
	d.Folder = folder.String
	json.Unmarshal([]byte(tagsJSON), &d.Tags)
	json.Unmarshal([]byte(refIDs.String), &d.ReferenceIDs)
	if d.Tags == nil {
		d.Tags = []string{}
	}
	if d.ReferenceIDs == nil {
		d.ReferenceIDs = []string{}
	}
	return &d, nil
}

func scanDocRow(row *sql.Row) (*domain.Document, error) {
	var d domain.Document
	var tagsJSON string
	var department string
	var folder, refIDs sql.NullString
	err := row.Scan(&d.ID, &d.Title, &d.Content, &d.Classification, &d.Status, &department, &folder, &tagsJSON, &refIDs, &d.ContentHash, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d.Department = domain.Department(department)
	d.Folder = folder.String
	json.Unmarshal([]byte(tagsJSON), &d.Tags)
	json.Unmarshal([]byte(refIDs.String), &d.ReferenceIDs)
	if d.Tags == nil {
		d.Tags = []string{}
	}
	if d.ReferenceIDs == nil {
		d.ReferenceIDs = []string{}
	}
	return &d, nil
}

const docColumns = `id, title, content, classification, status, department, folder, tags, reference_ids, content_hash, created_by, created_at, updated_at`

func (r *DocumentRepository) FindAll() ([]*domain.Document, error) {
	rows, err := r.db.Query(
		`SELECT ` + docColumns + ` FROM documents ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		d, err := scanDoc(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, d)
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

	query := `SELECT ` + docColumns + ` FROM documents WHERE id IN (` + strings.Join(placeholders, ",") + `) ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		d, err := scanDoc(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *DocumentRepository) FindByID(id string) (*domain.Document, error) {
	row := r.db.QueryRow(`SELECT `+docColumns+` FROM documents WHERE id = ?`, id)
	return scanDocRow(row)
}

func (r *DocumentRepository) Create(d *domain.Document) error {
	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now

	tagsJSON, _ := json.Marshal(d.Tags)
	if string(tagsJSON) == "null" {
		tagsJSON = []byte("[]")
	}
	refsJSON, _ := json.Marshal(d.ReferenceIDs)
	if string(refsJSON) == "null" {
		refsJSON = []byte("[]")
	}
	if d.Department == "" {
		d.Department = "public"
	}

	_, err := r.db.Exec(
		`INSERT INTO documents (id, title, content, classification, status, department, folder, tags, reference_ids, content_hash, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.Title, d.Content, d.Classification, d.Status, string(d.Department), d.Folder, string(tagsJSON), string(refsJSON), d.ContentHash, d.CreatedBy, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DocumentRepository) Update(d *domain.Document) error {
	d.UpdatedAt = time.Now()

	tagsJSON, _ := json.Marshal(d.Tags)
	if string(tagsJSON) == "null" {
		tagsJSON = []byte("[]")
	}
	refsJSON, _ := json.Marshal(d.ReferenceIDs)
	if string(refsJSON) == "null" {
		refsJSON = []byte("[]")
	}

	_, err := r.db.Exec(
		`UPDATE documents SET title=?, content=?, classification=?, status=?, department=?, folder=?, tags=?, reference_ids=?, content_hash=?, updated_at=?
		 WHERE id=?`,
		d.Title, d.Content, d.Classification, d.Status, string(d.Department), d.Folder, string(tagsJSON), string(refsJSON), d.ContentHash, d.UpdatedAt, d.ID,
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
	Department     string    `json:"department"`
	Folder         string    `json:"folder"`
	Tags           []string  `json:"tags"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (r *DocumentRepository) FindAllMetadata(limit, offset int) ([]DocMetadata, error) {
	rows, err := r.db.Query(
		`SELECT id, title, classification, status, department, folder, tags, created_by, created_at, updated_at
		 FROM documents ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []DocMetadata
	for rows.Next() {
		var d DocMetadata
		var tagsJSON string
		var folder sql.NullString
		if err := rows.Scan(&d.ID, &d.Title, &d.Classification, &d.Status, &d.Department, &folder, &tagsJSON, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.Folder = folder.String
		json.Unmarshal([]byte(tagsJSON), &d.Tags)
		if d.Tags == nil {
			d.Tags = []string{}
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *DocumentRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&count)
	return count, err
}

func (r *DocumentRepository) SearchContent(query string) ([]DocMetadata, error) {
	rows, err := r.db.Query(
		`SELECT d.id, d.title, d.classification, d.status, d.department, d.folder, d.tags, d.created_by, d.created_at, d.updated_at
		 FROM documents d
		 INNER JOIN documents_fts fts ON d.rowid = fts.rowid
		 WHERE documents_fts MATCH ?
		 ORDER BY rank LIMIT 50`,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []DocMetadata
	for rows.Next() {
		var d DocMetadata
		var tagsJSON string
		var folder sql.NullString
		if err := rows.Scan(&d.ID, &d.Title, &d.Classification, &d.Status, &d.Department, &folder, &tagsJSON, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.Folder = folder.String
		json.Unmarshal([]byte(tagsJSON), &d.Tags)
		if d.Tags == nil {
			d.Tags = []string{}
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}
