package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/repository"
)

type DocumentService interface {
	List(session auth.Session) ([]*domain.Document, error)
	GetByID(session auth.Session, id string) (*domain.Document, error)
	Create(session auth.Session, doc *domain.Document) (*domain.Document, error)
	Update(session auth.Session, id string, doc *domain.Document) (*domain.Document, error)
	Delete(session auth.Session, id string) error
	Transition(session auth.Session, id string, to domain.DocumentStatus) (*domain.Document, error)
	Catalog(limit, offset int) ([]repository.DocMetadata, error)
	CountDocuments() (int, error)
	Search(query string, session auth.Session) ([]repository.DocMetadata, error)
	ExportToMarkdown(doc *domain.Document) (string, error)
	RecentlyViewed(userID string) []string
	TrieSearch(prefix string) []struct {
		Word  string
		DocID string
	}
	FeaturedScrolls(n int) []struct {
		DocID string
		Score int
	}
}

type DocumentHandler struct {
	service DocumentService
}

func NewDocumentHandler(service DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

// List returns scrolls accessible to the authenticated villager based on tier and department.
// @Summary      List accessible scrolls
// @Tags         scrolls
// @Produce      json
// @Param        Authorization header string true "Bearer {token}"
// @Success      200  {array}   domain.Document
// @Failure      401  {object}  map[string]string
// @Router       /api/documents [get]
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	docs, err := h.service.List(*session)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list scrolls"})
		return
	}

	writeJSON(w, http.StatusOK, docs)
}

// Get returns a single scroll by ID. Access denied if tier or department insufficient.
// @Summary      Get scroll by ID
// @Tags         scrolls
// @Produce      json
// @Param        Authorization header string true "Bearer {token}"
// @Param        id path string true "Scroll ID"
// @Success      200  {object}  domain.Document
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/documents/{id} [get]
func (h *DocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	id := r.PathValue("id")
	doc, err := h.service.GetByID(*session, id)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	var doc domain.Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	created, err := h.service.Create(*session, &doc)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create document"})
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *DocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	id := r.PathValue("id")
	var doc domain.Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updated, err := h.service.Update(*session, id, &doc)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update document"})
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	id := r.PathValue("id")
	if err := h.service.Delete(*session, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete document"})
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *DocumentHandler) Transition(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	id := r.PathValue("id")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	doc, err := h.service.Transition(*session, id, domain.DocumentStatus(req.Status))
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *DocumentHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	docs, err := h.service.Catalog(limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list catalog"})
		return
	}

	total, err := h.service.CountDocuments()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to count documents"})
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	writeJSON(w, http.StatusOK, docs)
}

func (h *DocumentHandler) Search(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'q' required"})
		return
	}

	docs, err := h.service.Search(q, *session)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
		return
	}

	writeJSON(w, http.StatusOK, docs)
}

func (h *DocumentHandler) Export(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	id := r.PathValue("id")
	doc, err := h.service.GetByID(*session, id)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "access denied or not found"})
		return
	}

	md, err := h.service.ExportToMarkdown(doc)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "export failed"})
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+doc.Title+".md\"")
	w.Write([]byte(md))
}

func (h *DocumentHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	q := r.URL.Query().Get("q")
	matching := h.service.TrieSearch(q)
	writeJSON(w, http.StatusOK, matching)
}

func (h *DocumentHandler) Featured(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	featured := h.service.FeaturedScrolls(5)
	writeJSON(w, http.StatusOK, featured)
}

func (h *DocumentHandler) Recent(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	docIDs := h.service.RecentlyViewed(session.UserID)
	if docIDs == nil {
		docIDs = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document_ids": docIDs,
	})
}
