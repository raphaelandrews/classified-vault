package handler

import (
	"encoding/json"
	"net/http"

	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
)

type DocumentService interface {
	List(session auth.Session) ([]*domain.Document, error)
	GetByID(session auth.Session, id string) (*domain.Document, error)
	Create(session auth.Session, doc *domain.Document) (*domain.Document, error)
	Update(session auth.Session, id string, doc *domain.Document) (*domain.Document, error)
	Delete(session auth.Session, id string) error
}

type DocumentHandler struct {
	service DocumentService
}

func NewDocumentHandler(service DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

// List returns documents accessible to the authenticated user based on clearance.
// @Summary      List accessible documents
// @Tags         documents
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list documents"})
		return
	}

	writeJSON(w, http.StatusOK, docs)
}

// Get returns a single document by ID. Access denied if clearance is insufficient.
// @Summary      Get document by ID
// @Tags         documents
// @Produce      json
// @Param        Authorization header string true "Bearer {token}"
// @Param        id path string true "Document ID"
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
