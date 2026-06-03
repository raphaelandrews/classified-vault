package service

import (
	"fmt"
	"slices"
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
	maxTier := int(session.Clearance)
	if session.Faction == domain.FactionMayorsOffice && session.Clearance >= domain.TierArcane {
		maxTier = int(domain.TierJunimo)
	}
	if session.Faction == domain.FactionWizardsTower && maxTier < int(domain.TierArcane) {
		maxTier = int(domain.TierArcane)
	}

	accessibleIDs := s.index.QueryUpTo(maxTier)
	docs, err := s.repo.FindByIDs(accessibleIDs)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Document
	for _, doc := range docs {
		if canAccess(session, doc) {
			filtered = append(filtered, doc)
		}
	}
	return filtered, nil
}

func (s *DocumentService) GetByID(session auth.Session, docID string) (*domain.Document, error) {
	doc, err := s.repo.FindByID(docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	if !canAccess(session, doc) {
		s.logAudit(domain.AuditLog{
			UserID:   session.UserID,
			Username: session.Username,
			Action:   domain.ActionAccessDenied,
			Resource: "scroll:" + docID,
			Success:  false,
			Details:  fmt.Sprintf("tier %d (%s) < %d (%s) faction=%s", session.Clearance, session.Clearance, doc.Classification, doc.Classification, doc.Faction),
		})
		return nil, fmt.Errorf("access denied")
	}

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionScrollRead,
		Resource: "scroll:" + docID,
		Success:  true,
	})

	return doc, nil
}

func (s *DocumentService) Create(session auth.Session, doc *domain.Document) (*domain.Document, error) {
	doc.ID = "scr_" + uuid.New().String()[:8]
	doc.CreatedBy = session.UserID
	if doc.Faction == "" {
		doc.Faction = session.Faction
	}

	if err := s.repo.Create(doc); err != nil {
		return nil, err
	}

	s.index.Insert(int(doc.Classification), doc.ID)

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionScrollCreate,
		Resource: "scroll:" + doc.ID,
		Success:  true,
		Details:  fmt.Sprintf("title=%s tier=%s faction=%s", doc.Title, doc.Classification, doc.Faction),
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

	if !canAccess(session, existing) {
		s.logAudit(domain.AuditLog{
			UserID:   session.UserID,
			Username: session.Username,
			Action:   domain.ActionAccessDenied,
			Resource: "scroll:" + id,
			Success:  false,
			Details:  "update denied: insufficient tier or wrong faction",
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
		Action:   domain.ActionScrollUpdate,
		Resource: "scroll:" + id,
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

	if !canAccess(session, existing) || string(session.Role) != string(domain.RoleMayor) {
		if !canAccess(session, existing) {
			s.logAudit(domain.AuditLog{
				UserID:   session.UserID,
				Username: session.Username,
				Action:   domain.ActionAccessDenied,
				Resource: "scroll:" + id,
				Success:  false,
				Details:  "delete denied: insufficient tier or wrong faction",
			})
			return fmt.Errorf("access denied")
		}
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	s.index.Remove(int(existing.Classification), id)

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionScrollDelete,
		Resource: "scroll:" + id,
		Success:  true,
	})

	return nil
}

func (s *DocumentService) Catalog() ([]repository.DocMetadata, error) {
	return s.repo.FindAllMetadata()
}

func canAccess(session auth.Session, doc *domain.Document) bool {
	if doc.Classification == domain.TierPublic {
		return true
	}

	if session.Faction == doc.Faction && session.Clearance >= doc.Classification {
		return true
	}

	if session.Faction == domain.FactionMayorsOffice && session.Clearance >= domain.TierArcane {
		return true
	}

	if session.Faction == domain.FactionWizardsTower && slices.Contains(doc.Tags, "arcane") {
		return true
	}

	return false
}

func (s *DocumentService) logAudit(log domain.AuditLog) {
	log.ID = uuid.New().String()
	log.Timestamp = time.Now()
	s.auditBuffer.Append(log)
	go s.auditRepo.Save(&log)
}
