package handler

import (
	"encoding/json"
	"net/http"

	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/middleware"
)

type UserService interface {
	List() ([]*domain.User, error)
	GetByID(id string) (*domain.User, error)
	Create(user *domain.User) (*domain.User, error)
	Update(id string, user *domain.User) (*domain.User, error)
	Delete(id string) error
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	users, err := h.service.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	created, err := h.service.Create(&user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	id := r.PathValue("id")
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updated, err := h.service.Update(id, &user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user"})
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	id := r.PathValue("id")
	if err := h.service.Delete(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete user"})
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func requireSession(w http.ResponseWriter, r *http.Request) *auth.Session {
	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return nil
	}
	return session
}
