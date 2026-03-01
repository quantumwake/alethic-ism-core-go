package embedding

import (
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
)

// ContentType represents the type of content stored in a document.
type ContentType string

const (
	ContentText     = ContentType("text")
	ContentMarkdown = ContentType("markdown")
	ContentCode     = ContentType("code")
	ContentJSON     = ContentType("json")
)

// ScopeType represents the type of scope a document is linked to.
type ScopeType string

const (
	ScopeState        = ScopeType("state")
	ScopeRoute        = ScopeType("route")
	ScopeProcessor    = ScopeType("processor")
	ScopeDocument     = ScopeType("document")
	ScopeConversation = ScopeType("conversation")
	ScopeToolResult   = ScopeType("tool_result")
)

// Document represents a text document with its embedding vector.
type Document struct {
	ID          string      `gorm:"column:id;type:varchar(36);default:gen_random_uuid();primaryKey" json:"id"`
	ParentID    *string     `gorm:"column:parent_id;type:varchar(36);index" json:"parent_id,omitempty"`
	UserID      string      `gorm:"column:user_id;type:varchar(36);not null" json:"user_id"`
	ProjectID   *string     `gorm:"column:project_id;type:varchar(36)" json:"project_id,omitempty"`
	SessionID   *string     `gorm:"column:session_id;type:varchar(36);index" json:"session_id,omitempty"`
	ScopeType   ScopeType   `gorm:"column:scope_type;type:varchar(64)" json:"scope_type,omitempty"`
	ScopeID     *string     `gorm:"column:scope_id;type:varchar(36)" json:"scope_id,omitempty"`
	Content     string      `gorm:"column:content;type:text;not null" json:"content"`
	ContentType ContentType `gorm:"column:content_type;type:varchar(32);not null;default:'text'" json:"content_type"`
	Embedding   []float32   `gorm:"-" json:"embedding,omitempty"`
	Dimensions  int         `gorm:"column:dimensions;type:int;not null" json:"dimensions"`
	Model       string      `gorm:"column:model;type:varchar(128);not null" json:"model"`
	Metadata    data.JSON   `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (Document) TableName() string {
	return "embedding_document"
}

// SearchFilter defines the filtering criteria for a similarity search.
// UserID is required; all other fields are optional additional filters.
type SearchFilter struct {
	UserID    string
	ProjectID *string
	SessionID *string
	ScopeType *ScopeType
	ScopeID   *string
}

// SearchResult pairs a document with its similarity score.
type SearchResult struct {
	Document   Document `json:"document"`
	Similarity float64  `json:"similarity"`
}

// SearchOptions controls search behavior.
type SearchOptions struct {
	Limit         int
	MinSimilarity *float64
}
