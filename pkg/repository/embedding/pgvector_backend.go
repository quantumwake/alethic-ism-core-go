package embedding

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"gorm.io/gorm"
)

// documentRow is an internal struct that adds the pgvector.Vector field for
// database scanning. The public Document keeps Embedding as []float32 so
// the Storage interface stays backend-agnostic.
type documentRow struct {
	ID          string          `gorm:"column:id;primaryKey"`
	ParentID    *string         `gorm:"column:parent_id"`
	UserID      string          `gorm:"column:user_id"`
	ProjectID   *string         `gorm:"column:project_id"`
	SessionID   *string         `gorm:"column:session_id"`
	ScopeType   ScopeType       `gorm:"column:scope_type"`
	ScopeID     *string         `gorm:"column:scope_id"`
	Content     string          `gorm:"column:content"`
	ContentType ContentType     `gorm:"column:content_type"`
	EmbeddingV  pgvector.Vector `gorm:"column:embedding;type:vector"`
	Dimensions  int             `gorm:"column:dimensions"`
	Model       string          `gorm:"column:model"`
	Metadata    data.JSON       `gorm:"column:metadata;type:jsonb"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

func (documentRow) TableName() string {
	return "embedding_document"
}

func (r *documentRow) toDocument() Document {
	return Document{
		ID:          r.ID,
		ParentID:    r.ParentID,
		UserID:      r.UserID,
		ProjectID:   r.ProjectID,
		SessionID:   r.SessionID,
		ScopeType:   r.ScopeType,
		ScopeID:     r.ScopeID,
		Content:     r.Content,
		ContentType: r.ContentType,
		Embedding:   r.EmbeddingV.Slice(),
		Dimensions:  r.Dimensions,
		Model:       r.Model,
		Metadata:    r.Metadata,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// PgVectorStorage implements Storage backed by PostgreSQL with pgvector.
type PgVectorStorage struct {
	Storage
	*repository.Access
}

// NewPgVectorStorage creates a new PgVectorStorage using the given DSN.
func NewPgVectorStorage(dsn string) *PgVectorStorage {
	return &PgVectorStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// NewPgVectorStorageFromDB creates a PgVectorStorage from an existing gorm.DB.
func NewPgVectorStorageFromDB(db *gorm.DB) *PgVectorStorage {
	return &PgVectorStorage{
		Access: &repository.Access{DB: db},
	}
}

func (s *PgVectorStorage) Upsert(doc *Document) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	vec := pgvector.NewVector(doc.Embedding)

	return s.DB.Exec(`
		INSERT INTO embedding_document
			(id, parent_id, user_id, project_id, session_id, scope_type, scope_id,
			 content, content_type, embedding, dimensions, model, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			parent_id    = EXCLUDED.parent_id,
			user_id      = EXCLUDED.user_id,
			project_id   = EXCLUDED.project_id,
			session_id   = EXCLUDED.session_id,
			scope_type   = EXCLUDED.scope_type,
			scope_id     = EXCLUDED.scope_id,
			content      = EXCLUDED.content,
			content_type = EXCLUDED.content_type,
			embedding    = EXCLUDED.embedding,
			dimensions   = EXCLUDED.dimensions,
			model        = EXCLUDED.model,
			metadata     = EXCLUDED.metadata,
			updated_at   = NOW()`,
		doc.ID, doc.ParentID, doc.UserID, doc.ProjectID, doc.SessionID,
		doc.ScopeType, doc.ScopeID,
		doc.Content, doc.ContentType, vec, doc.Dimensions, doc.Model,
		doc.Metadata, doc.CreatedAt, doc.UpdatedAt,
	).Error
}

func (s *PgVectorStorage) UpsertBatch(docs []*Document) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		txStore := &PgVectorStorage{Access: &repository.Access{DB: tx}}
		for _, doc := range docs {
			if err := txStore.Upsert(doc); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *PgVectorStorage) FindByID(id string) (*Document, error) {
	var row documentRow
	if err := s.DB.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	doc := row.toDocument()
	return &doc, nil
}

func (s *PgVectorStorage) FindByParentID(parentID string) ([]Document, error) {
	var rows []documentRow
	if err := s.DB.Where("parent_id = ?", parentID).Find(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]Document, len(rows))
	for i, r := range rows {
		docs[i] = r.toDocument()
	}
	return docs, nil
}

func (s *PgVectorStorage) Delete(id string) error {
	return s.DB.Where("id = ?", id).Delete(&documentRow{}).Error
}

func (s *PgVectorStorage) DeleteByParentID(parentID string) error {
	return s.DB.Where("parent_id = ?", parentID).Delete(&documentRow{}).Error
}

func (s *PgVectorStorage) Search(emb []float32, filter SearchFilter, opts SearchOptions) ([]SearchResult, error) {
	vec := pgvector.NewVector(emb)
	dims := len(emb)

	var where []string
	var args []interface{}

	// Always filter by user
	where = append(where, "user_id = ?")
	args = append(args, filter.UserID)

	// Dimension filter ensures partial HNSW index usage
	where = append(where, "dimensions = ?")
	args = append(args, dims)

	if filter.ProjectID != nil {
		where = append(where, "project_id = ?")
		args = append(args, *filter.ProjectID)
	}
	if filter.SessionID != nil {
		where = append(where, "session_id = ?")
		args = append(args, *filter.SessionID)
	}
	if filter.ScopeType != nil {
		where = append(where, "scope_type = ?")
		args = append(args, string(*filter.ScopeType))
	}
	if filter.ScopeID != nil {
		where = append(where, "scope_id = ?")
		args = append(args, *filter.ScopeID)
	}

	if opts.MinSimilarity != nil {
		where = append(where, fmt.Sprintf("1 - (embedding::vector(%d) <=> ?) >= ?", dims))
		args = append(args, vec, *opts.MinSimilarity)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	whereClause := strings.Join(where, " AND ")

	query := fmt.Sprintf(`
		SELECT id, parent_id, user_id, project_id, session_id,
		       scope_type, scope_id, content, content_type,
		       embedding, dimensions, model, metadata,
		       created_at, updated_at,
		       1 - (embedding::vector(%d) <=> ?) AS similarity
		FROM embedding_document
		WHERE %s
		ORDER BY embedding::vector(%d) <=> ?
		LIMIT ?`, dims, whereClause, dims)

	// Prepend SELECT similarity vec before WHERE args, then append ORDER BY vec + LIMIT
	// The ? placeholders are ordered: SELECT(vec), WHERE(...), ORDER BY(vec), LIMIT
	args = append([]interface{}{vec}, args...)
	args = append(args, vec, limit)

	type searchRow struct {
		documentRow
		Similarity float64 `gorm:"column:similarity"`
	}

	var rows []searchRow
	if err := s.DB.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("embedding search failed: %w", err)
	}

	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		results[i] = SearchResult{
			Document:   r.documentRow.toDocument(),
			Similarity: r.Similarity,
		}
	}
	return results, nil
}
