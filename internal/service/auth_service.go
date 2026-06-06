package service

import (
	"fmt"
	"sync"
	"time"

	"classified-vault/config"
	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/ds"
	"classified-vault/internal/repository"
	"classified-vault/internal/validate"
)

type LockoutError struct {
	Duration time.Duration
}

func (e *LockoutError) Error() string {
	return fmt.Sprintf("account locked: %s remaining", e.Duration.Truncate(time.Second))
}

type LockoutState struct {
	attempts     int
	firstAttempt time.Time
	lockedUntil  time.Time
}

type AuthService struct {
	userRepo     *repository.UserRepository
	sessionCache *ds.HashMap[auth.Session]
	cfg          *config.Config
	lockoutsMu   sync.Mutex
	lockouts     map[string]*LockoutState
}

func NewAuthService(userRepo *repository.UserRepository, sessionCache *ds.HashMap[auth.Session], cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionCache: sessionCache,
		cfg:          cfg,
		lockouts:     make(map[string]*LockoutState),
	}
}

func (s *AuthService) Login(username, password, ip string) (auth.Session, string, *domain.User, error) {
	s.lockoutsMu.Lock()
	state, exists := s.lockouts[ip]
	if exists && time.Now().Before(state.lockedUntil) {
		remaining := state.lockedUntil.Sub(time.Now())
		s.lockoutsMu.Unlock()
		return auth.Session{}, "", nil, &LockoutError{Duration: remaining}
	}
	if exists && !time.Now().Before(state.lockedUntil) {
		delete(s.lockouts, ip)
	}
	s.lockoutsMu.Unlock()

	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return auth.Session{}, "", nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil || !user.Active {
		return auth.Session{}, "", nil, fmt.Errorf("invalid credentials")
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		s.lockoutsMu.Lock()
		st, ok := s.lockouts[ip]
		if !ok {
			st = &LockoutState{firstAttempt: time.Now()}
			s.lockouts[ip] = st
		}
		st.attempts++
		if st.attempts >= 5 {
			st.lockedUntil = time.Now().Add(15 * time.Minute)
		}
		s.lockoutsMu.Unlock()
		return auth.Session{}, "", nil, fmt.Errorf("invalid credentials")
	}

	s.lockoutsMu.Lock()
	delete(s.lockouts, ip)
	s.lockoutsMu.Unlock()

	token := auth.NewToken()
	session := auth.Session{
		UserID:     user.ID,
		Username:   user.Username,
		Role:       user.Role,
		RoleName:   user.RoleName,
		Clearance:  user.Clearance,
		Department: user.Department,
		ExpiresAt:  time.Now().Add(s.cfg.SessionTTL),
	}

	s.sessionCache.Set(token, session)

	return session, token, user, nil
}

func (s *AuthService) Register(username, password string, department domain.Department) (*domain.User, error) {
	if err := validate.Password(password); err != nil {
		return nil, fmt.Errorf("password: %w", err)
	}

	existing, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("username already taken")
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: password,
		Role:         domain.RoleAssociate,
		RoleName:     string(domain.RoleAssociate),
		Department:   department,
	}

	userID := auth.NewToken()[:12]
	user.ID = "usr_" + userID

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = hash
	user.Active = true
	user.Clearance = domain.MaxClearanceForRole(user.Role)

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
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

func (s *AuthService) RefreshSession(token string) error {
	session, ok := s.sessionCache.Get(token)
	if !ok {
		return fmt.Errorf("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		s.sessionCache.Delete(token)
		return fmt.Errorf("session expired")
	}
	session.ExpiresAt = time.Now().Add(s.cfg.SessionTTL)
	s.sessionCache.Set(token, session)
	return nil
}

func (s *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	if err := validate.Password(newPassword); err != nil {
		return fmt.Errorf("new password: %w", err)
	}

	if currentPassword == newPassword {
		return fmt.Errorf("new password must differ from current password")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if !auth.CheckPassword(currentPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = hash

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

func (s *AuthService) GetLockoutRemaining(ip string) time.Duration {
	s.lockoutsMu.Lock()
	defer s.lockoutsMu.Unlock()
	state, exists := s.lockouts[ip]
	if !exists || !time.Now().Before(state.lockedUntil) {
		if exists {
			delete(s.lockouts, ip)
		}
		return 0
	}
	return state.lockedUntil.Sub(time.Now())
}
