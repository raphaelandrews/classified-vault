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

type DocumentService struct {
	repo        *repository.DocumentRepository
	index       *ds.AVLTree
	auditBuffer *ds.LinkedList[domain.AuditLog]
	auditRepo   *repository.AuditRepository
}

func NewDocumentService(
	repo *repository.DocumentRepository,
	index *ds.AVLTree,
	auditBuffer *ds.LinkedList[domain.AuditLog],
	auditRepo *repository.AuditRepository,
) *DocumentService {
	return &DocumentService{
		repo:        repo,
		index:       index,
		auditBuffer: auditBuffer,
		auditRepo:   auditRepo,
	}
}

func (s *DocumentService) List(session auth.Session) ([]*domain.Document, error) {
	accessibleIDs := s.index.QueryUpTo(int(session.Clearance))
	return s.repo.FindByIDs(accessibleIDs)
}

func (s *DocumentService) GetByID(session auth.Session, docID string) (*domain.Document, error) {
	doc, err := s.repo.FindByID(docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	if session.Clearance < doc.Classification {
		s.logAudit(domain.AuditLog{
			UserID:   session.UserID,
			Username: session.Username,
			Action:   domain.ActionAccessDenied,
			Resource: "document:" + docID,
			Success:  false,
			Details:  fmt.Sprintf("clearance %s < %s", session.Clearance, doc.Classification),
		})
		return nil, fmt.Errorf("access denied")
	}

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionDocumentRead,
		Resource: "document:" + docID,
		Success:  true,
	})

	return doc, nil
}

func (s *DocumentService) Create(session auth.Session, doc *domain.Document) (*domain.Document, error) {
	doc.ID = "doc_" + uuid.New().String()[:8]
	doc.CreatedBy = session.UserID

	if err := s.repo.Create(doc); err != nil {
		return nil, err
	}

	s.index.Insert(int(doc.Classification), doc.ID)

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionDocumentCreate,
		Resource: "document:" + doc.ID,
		Success:  true,
		Details:  fmt.Sprintf("title=%s classification=%s", doc.Title, doc.Classification),
	})

	return doc, nil
}

func (s *DocumentService) Update(session auth.Session, id string, doc *domain.Document) (*domain.Document, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("document not found")
	}

	if session.Clearance < existing.Classification {
		s.logAudit(domain.AuditLog{
			UserID:   session.UserID,
			Username: session.Username,
			Action:   domain.ActionAccessDenied,
			Resource: "document:" + id,
			Success:  false,
			Details:  "update denied: insufficient clearance",
		})
		return nil, fmt.Errorf("access denied")
	}

	if existing.Classification != doc.Classification {
		s.index.Remove(int(existing.Classification), id)
		s.index.Insert(int(doc.Classification), id)
	}

	doc.ID = id
	if err := s.repo.Update(doc); err != nil {
		return nil, err
	}

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionDocumentUpdate,
		Resource: "document:" + id,
		Success:  true,
	})

	return doc, nil
}

func (s *DocumentService) Delete(session auth.Session, id string) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("document not found")
	}

	if session.Clearance < existing.Classification {
		s.logAudit(domain.AuditLog{
			UserID:   session.UserID,
			Username: session.Username,
			Action:   domain.ActionAccessDenied,
			Resource: "document:" + id,
			Success:  false,
			Details:  "delete denied: insufficient clearance",
		})
		return fmt.Errorf("access denied")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	s.index.Remove(int(existing.Classification), id)

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionDocumentDelete,
		Resource: "document:" + id,
		Success:  true,
	})

	return nil
}

func (s *DocumentService) Catalog() ([]repository.DocMetadata, error) {
	return s.repo.FindAllMetadata()
}

func (s *DocumentService) logAudit(log domain.AuditLog) {
	log.ID = uuid.New().String()
	log.Timestamp = time.Now()
	s.auditBuffer.Append(log)
	go s.auditRepo.Save(&log)
}
