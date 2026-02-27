package embedding

import (
	"gorm.io/gorm"
)

const createEmbeddingTableSQL = `
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS embedding_document (
    id            VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id     VARCHAR(36),
    user_id       VARCHAR(36) NOT NULL,
    project_id    VARCHAR(36),
    session_id    VARCHAR(36),
    scope_type    VARCHAR(64),
    scope_id      VARCHAR(36),
    content       TEXT NOT NULL,
    content_type  VARCHAR(32) NOT NULL DEFAULT 'text',
    embedding     vector,
    dimensions    INT NOT NULL,
    model         VARCHAR(128) NOT NULL,
    metadata      JSONB,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_embedding_doc_user_project
    ON embedding_document (user_id, project_id);

CREATE INDEX IF NOT EXISTS idx_embedding_doc_scope
    ON embedding_document (scope_type, scope_id);

CREATE INDEX IF NOT EXISTS idx_embedding_doc_session
    ON embedding_document (session_id);

CREATE INDEX IF NOT EXISTS idx_embedding_doc_parent
    ON embedding_document (parent_id);

CREATE INDEX IF NOT EXISTS idx_embedding_doc_hnsw_384
    ON embedding_document
    USING hnsw ((embedding::vector(384)) vector_cosine_ops)
    WHERE dimensions = 384;

CREATE INDEX IF NOT EXISTS idx_embedding_doc_hnsw_768
    ON embedding_document
    USING hnsw ((embedding::vector(768)) vector_cosine_ops)
    WHERE dimensions = 768;

CREATE INDEX IF NOT EXISTS idx_embedding_doc_hnsw_1536
    ON embedding_document
    USING hnsw ((embedding::vector(1536)) vector_cosine_ops)
    WHERE dimensions = 1536;
`

// EnsureTable creates the embedding_document table and indexes if they do not exist.
func EnsureTable(db *gorm.DB) error {
	return db.Exec(createEmbeddingTableSQL).Error
}
