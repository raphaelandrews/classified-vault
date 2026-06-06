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

func (s *AuditService) List(limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditRepo.FindAll(limit, offset)
}

func (s *AuditService) Count() (int, error) {
	return s.auditRepo.Count()
}

func (s *AuditService) ListByUser(userID string, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditRepo.FindByUser(userID, limit, offset)
}
