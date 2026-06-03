package service

import (
	"fmt"
	"time"

	"classified-vault/config"
	"classified-vault/internal/auth"
	"classified-vault/internal/ds"
	"classified-vault/internal/repository"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	sessionCache *ds.HashMap[auth.Session]
	cfg          *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, sessionCache *ds.HashMap[auth.Session], cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionCache: sessionCache,
		cfg:          cfg,
	}
}

func (s *AuthService) Login(username, password, ip string) (auth.Session, string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return auth.Session{}, "", fmt.Errorf("find user: %w", err)
	}
	if user == nil || !user.Active {
		return auth.Session{}, "", fmt.Errorf("invalid credentials")
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		return auth.Session{}, "", fmt.Errorf("invalid credentials")
	}

	token := auth.NewToken()
	session := auth.Session{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Clearance: user.Clearance,
		Faction:   user.Faction,
		ExpiresAt: time.Now().Add(s.cfg.SessionTTL),
	}

	s.sessionCache.Set(token, session)

	return session, token, nil
}

func (s *AuthService) Logout(token string) error {
	s.sessionCache.Delete(token)
	return nil
}

func (s *AuthService) GetSession(token string) (*auth.Session, error) {
	session, ok := s.sessionCache.Get(token)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		s.sessionCache.Delete(token)
		return nil, fmt.Errorf("session expired")
	}
	return &session, nil
}
