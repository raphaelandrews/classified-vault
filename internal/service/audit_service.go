package service

import (
	"classified-vault/internal/domain"
	"classified-vault/internal/ds"
	"classified-vault/internal/repository"
)

type AuditService struct {
	auditRepo   *repository.AuditRepository
	auditBuffer *ds.LinkedList[domain.AuditLog]
}

func NewAuditService(
	auditRepo *repository.AuditRepository,
	auditBuffer *ds.LinkedList[domain.AuditLog],
) *AuditService {
	return &AuditService{
		auditRepo:   auditRepo,
		auditBuffer: auditBuffer,
	}
}

func (s *AuditService) List(limit int) ([]*domain.AuditLog, error) {
	dbLogs, err := s.auditRepo.FindAll(limit)
	if err != nil {
		return nil, err
	}

	return dbLogs, nil
}

func (s *AuditService) ListByUser(userID string, limit int) ([]*domain.AuditLog, error) {
	return s.auditRepo.FindByUser(userID, limit)
}
