package service

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"classified-vault/internal/auth"
	vaultcrypto "classified-vault/internal/crypto"
	"classified-vault/internal/domain"
	"classified-vault/internal/ds"
	"classified-vault/internal/repository"
)

type DocumentService struct {
	repo        *repository.DocumentRepository
	index       *ds.AVLTree
	auditBuffer *ds.LinkedList[domain.AuditLog]
	auditRepo   *repository.AuditRepository
	lruCache    *ds.LRUCache
	trie        *ds.Trie
}

func NewDocumentService(
	repo *repository.DocumentRepository,
	index *ds.AVLTree,
	auditBuffer *ds.LinkedList[domain.AuditLog],
	auditRepo *repository.AuditRepository,
	lruCache *ds.LRUCache,
	trie *ds.Trie,
) *DocumentService {
	return &DocumentService{
		repo:        repo,
		index:       index,
		auditBuffer: auditBuffer,
		auditRepo:   auditRepo,
		lruCache:    lruCache,
		trie:        trie,
	}
}

func (s *DocumentService) List(session auth.Session) ([]*domain.Document, error) {
	maxTier := int(session.Clearance)
	if session.Department == domain.DepartmentMayorsOffice && session.Clearance >= domain.TierArcane {
		maxTier = int(domain.TierJunimo)
	}
	if session.Department == domain.DepartmentWizardsTower && maxTier < int(domain.TierArcane) {
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
			doc.Content = decryptContent(doc.Content)
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
			Details:  fmt.Sprintf("tier %d (%s) < %d (%s) department=%s", session.Clearance, session.Clearance, doc.Classification, doc.Classification, doc.Department),
		})
		return nil, fmt.Errorf("access denied")
	}

	doc.Content = decryptContent(doc.Content)

	s.lruCache.Put(session.UserID+":"+docID, docID)

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
	if doc.Status == "" {
		doc.Status = domain.StatusDraft
	}
	if doc.Department == "" {
		doc.Department = session.Department
	}
	doc.ContentHash = doc.ComputeHash()
	doc.Content = encryptContent(doc.Content)

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
		Details:  fmt.Sprintf("title=%s tier=%s department=%s", doc.Title, doc.Classification, doc.Department),
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
			Details:  "update denied: insufficient tier or wrong department",
		})
		return nil, fmt.Errorf("access denied")
	}

	if existing.Classification != doc.Classification {
		s.index.Remove(int(existing.Classification), id)
		s.index.Insert(int(doc.Classification), id)
	}

	doc.ID = id
	doc.ContentHash = doc.ComputeHash()
	doc.Content = encryptContent(doc.Content)
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

func (s *DocumentService) Transition(session auth.Session, id string, to domain.DocumentStatus) (*domain.Document, error) {
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	if !canAccess(session, doc) {
		return nil, fmt.Errorf("access denied")
	}

	if !domain.CanTransition(doc.Status, to) {
		return nil, fmt.Errorf("cannot transition from %s to %s", doc.Status, to)
	}

	if domain.TransitionRequiresMayor(to) && session.Role != domain.RoleMayor {
		return nil, fmt.Errorf("only the Mayor can transition to %s", to)
	}

	if to == domain.StatusFrozen {
		doc.Content = decryptContent(doc.Content)
		doc.ContentHash = doc.ComputeHash()
		doc.Content = encryptContent(doc.Content)
	}

	doc.Status = to
	if err := s.repo.Update(doc); err != nil {
		return nil, err
	}

	s.logAudit(domain.AuditLog{
		UserID:   session.UserID,
		Username: session.Username,
		Action:   domain.ActionScrollUpdate,
		Resource: "scroll:" + id,
		Success:  true,
		Details:  fmt.Sprintf("status %s → %s", doc.Status, to),
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
				Details:  "delete denied: insufficient tier or wrong department",
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

func (s *DocumentService) Catalog(limit, offset int) ([]repository.DocMetadata, error) {
	return s.repo.FindAllMetadata(limit, offset)
}

func (s *DocumentService) CountDocuments() (int, error) {
	return s.repo.Count()
}

func (s *DocumentService) Search(query string, session auth.Session) ([]repository.DocMetadata, error) {
	return s.repo.SearchContent(query)
}

func (s *DocumentService) ExportToMarkdown(doc *domain.Document) (string, error) {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("title: " + doc.Title + "\n")
	sb.WriteString("id: " + doc.ID + "\n")
	sb.WriteString("tier: " + doc.Classification.String() + "\n")
	sb.WriteString("status: " + string(doc.Status) + "\n")
	sb.WriteString("department: " + string(doc.Department) + "\n")
	if doc.Folder != "" {
		sb.WriteString("folder: " + doc.Folder + "\n")
	}
	sb.WriteString("created_by: " + doc.CreatedBy + "\n")
	sb.WriteString("created_at: " + doc.CreatedAt.Format(time.RFC3339) + "\n")
	sb.WriteString("updated_at: " + doc.UpdatedAt.Format(time.RFC3339) + "\n")
	if len(doc.Tags) > 0 {
		sb.WriteString("tags: [" + strings.Join(doc.Tags, ", ") + "]\n")
	}
	sb.WriteString("---\n\n")
	sb.WriteString(doc.Content)
	sb.WriteString("\n")

	return sb.String(), nil
}

func (s *DocumentService) RecentlyViewed(userID string) []string {
	allKeys := s.lruCache.Keys()
	prefix := userID + ":"
	var docIDs []string
	for i := len(allKeys) - 1; i >= 0; i-- {
		if strings.HasPrefix(allKeys[i], prefix) {
			if id, ok := s.lruCache.Get(allKeys[i]); ok {
				if idStr, ok2 := id.(string); ok2 {
					docIDs = append(docIDs, idStr)
				}
			}
		}
	}
	return docIDs
}

func (s *DocumentService) FeaturedScrolls(n int) []struct {
	DocID string
	Score int
} {
	allDocs, err := s.repo.FindAll()
	if err != nil {
		return nil
	}

	heap := ds.NewMaxHeap()
	now := time.Now()
	for _, doc := range allDocs {
		recency := int(now.Sub(doc.CreatedAt).Hours() / 24)
		score := int(doc.Classification)*10 + recency
		heap.Insert(doc.ID, score)
	}

	return heap.TopN(n)
}

func (s *DocumentService) RebuildTrie() {
	allDocs, err := s.repo.FindAll()
	if err != nil {
		return
	}
	s.trie.Clear()
	for _, doc := range allDocs {
		s.trie.Insert(doc.Title, doc.ID)
	}
}

func (s *DocumentService) TrieSearch(prefix string) []struct {
	Word  string
	DocID string
} {
	return s.trie.SearchWithIDs(prefix)
}

func encryptContent(content string) string {
	enc, err := vaultcrypto.Encrypt(content)
	if err != nil {
		return content
	}
	return enc
}

func decryptContent(content string) string {
	if !vaultcrypto.IsEncrypted(content) {
		return content
	}
	dec, err := vaultcrypto.Decrypt(content)
	if err != nil {
		return "[VAULT ERROR: cannot decrypt]"
	}
	return dec
}

func canAccess(session auth.Session, doc *domain.Document) bool {
	if doc.Classification == domain.TierPublic {
		return true
	}

	if session.Department == doc.Department && session.Clearance >= doc.Classification {
		return true
	}

	if session.Department == domain.DepartmentMayorsOffice && session.Clearance >= domain.TierArcane {
		return true
	}

	if session.Department == domain.DepartmentWizardsTower && slices.Contains(doc.Tags, "arcane") {
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
