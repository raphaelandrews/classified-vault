package handler

import (
	"encoding/json"
	"net/http"

	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/ds"
	"classified-vault/internal/middleware"
)

type AuthService interface {
	Login(username, password, ip string) (auth.Session, string, *domain.User, error)
	Logout(token string) error
	GetSession(token string) (*auth.Session, error)
}

type AuthHandler struct {
	cache   *ds.HashMap[auth.Session]
	service AuthService
}

func NewAuthHandler(service AuthService, cache *ds.HashMap[auth.Session]) *AuthHandler {
	return &AuthHandler{service: service, cache: cache}
}

// Login authenticates a villager and returns a session token.
// @Summary      Sign In
// @Description  Authenticate with username and password to access Pelican Town Archives
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{username=string,password=string} true "Credentials"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	session, token, user, err := h.service.Login(req.Username, req.Password, r.RemoteAddr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	h.cache.Set(token, session)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := middleware.TokenFromContext(r.Context())
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing token"})
		return
	}

	if err := h.service.Logout(token); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "logout failed"})
		return
	}

	h.cache.Delete(token)
	writeJSON(w, http.StatusOK, map[string]string{"status": "signed out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	writeJSON(w, http.StatusOK, domain.User{
		ID:        session.UserID,
		Username:  session.Username,
		Role:      session.Role,
		Clearance: session.Clearance,
		Department:   session.Department,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
