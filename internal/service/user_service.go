package service

import (
	"fmt"
	"log/slog"
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

	password := user.PasswordHash
	if password == "" {
		password = uuid.New().String()[:12]
	}

	if err := s.validateUniqueRole(user.Role, user.Department, ""); err != nil {
		return nil, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = hash
	user.Active = true
	user.Clearance = domain.MaxClearanceForRole(user.Role)
	if user.RoleName == "" {
		user.RoleName = string(user.Role)
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	s.logAudit(domain.AuditLog{
		UserID:   user.ID,
		Username: user.Username,
		Action:   domain.ActionUserCreated,
		Resource: "villager:" + user.ID,
		Success:  true,
		Details:  fmt.Sprintf("role=%s tier=%s department=%s", user.Role, user.Clearance, user.Department),
	})

	return user, nil
}

func (s *UserService) validateUniqueRole(role domain.Role, dept domain.Department, excludeID string) error {
	switch role {
	case domain.RoleMayor:
		exists, err := s.userRepo.ExistsByRole(role, excludeID)
		if err != nil {
			return fmt.Errorf("check mayor uniqueness: %w", err)
		}
		if exists {
			return fmt.Errorf("only one Mayor may exist in Pelican Town")
		}
	case domain.RoleKeeper:
		exists, err := s.userRepo.ExistsByDepartmentAndRole(dept, role, excludeID)
		if err != nil {
			return fmt.Errorf("check director uniqueness: %w", err)
		}
		if exists {
			return fmt.Errorf("department %s already has a Director", dept)
		}
	}
	return nil
}

func (s *UserService) Update(id string, user *domain.User) (*domain.User, error) {
	existing, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.PasswordHash != "" {
		hash, err := auth.HashPassword(user.PasswordHash)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		user.PasswordHash = hash
	} else {
		user.PasswordHash = existing.PasswordHash
	}

	if user.Role != existing.Role || user.Department != existing.Department {
		user.Clearance = domain.MaxClearanceForRole(user.Role)
		if err := s.validateUniqueRole(user.Role, user.Department, id); err != nil {
			return nil, err
		}
	} else {
		user.Clearance = existing.Clearance
	}

	user.ID = id
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	s.logAudit(domain.AuditLog{
		UserID:   user.ID,
		Username: user.Username,
		Action:   domain.ActionUserUpdated,
		Resource: "villager:" + id,
		Success:  true,
		Details:  fmt.Sprintf("role=%s tier=%s department=%s", user.Role, user.Clearance, user.Department),
	})

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

	s.logAudit(domain.AuditLog{
		UserID:   existing.ID,
		Username: existing.Username,
		Action:   domain.ActionUserDeleted,
		Resource: "villager:" + id,
		Success:  true,
	})

	return nil
}

func (s *UserService) SeedMayor(defaultPassword string) error {
	existing, err := s.userRepo.FindByUsername("lewis")
	if err != nil {
		return fmt.Errorf("check mayor existence: %w", err)
	}
	if existing == nil {
		slog.Info("seeding default mayor", "username", "lewis")
		mayor := &domain.User{
			Username:     "lewis",
			Email:        "lewis@pelican.valley",
			Role:         domain.RoleMayor,
			RoleName:     string(domain.RoleMayor),
			Department:   domain.DepartmentMayorsOffice,
			PasswordHash: defaultPassword,
		}
		if _, err := s.Create(mayor); err != nil {
			return fmt.Errorf("create mayor: %w", err)
		}
	}
	return nil
}

func (s *UserService) logAudit(log domain.AuditLog) {
	log.ID = uuid.New().String()
	log.Timestamp = time.Now()
	s.auditBuffer.Append(log)
	go s.auditRepo.Save(&log)
}
