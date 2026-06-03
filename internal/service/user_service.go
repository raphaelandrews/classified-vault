package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/ds"
	"classified-vault/internal/repository"
)

type UserService struct {
	userRepo    *repository.UserRepository
	auditBuffer *ds.LinkedList[domain.AuditLog]
	auditRepo   *repository.AuditRepository
}

func NewUserService(
	userRepo *repository.UserRepository,
	auditBuffer *ds.LinkedList[domain.AuditLog],
	auditRepo *repository.AuditRepository,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		auditBuffer: auditBuffer,
		auditRepo:   auditRepo,
	}
}

func (s *UserService) List() ([]*domain.User, error) {
	return s.userRepo.FindAll()
}

func (s *UserService) GetByID(id string) (*domain.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) Create(user *domain.User) (*domain.User, error) {
	user.ID = "usr_" + uuid.New().String()[:8]

	hash, err := auth.HashPassword("changeme")
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = hash

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(id string, user *domain.User) (*domain.User, error) {
	existing, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.PasswordHash != "" && user.PasswordHash != existing.PasswordHash {
		hash, err := auth.HashPassword(user.PasswordHash)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		user.PasswordHash = hash
	} else {
		user.PasswordHash = existing.PasswordHash
	}

	user.ID = id
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(id string) error {
	existing, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("user not found")
	}

	if err := s.userRepo.Delete(id); err != nil {
		return err
	}

	return nil
}

func (s *UserService) logAudit(log domain.AuditLog) {
	log.ID = uuid.New().String()
	log.Timestamp = time.Now()
	s.auditBuffer.Append(log)
	go s.auditRepo.Save(&log)
}
