package embedding

// Storage defines the interface for document embedding storage backends.
// Embeddings are represented as []float32 throughout the interface; each
// backend implementation converts to its native vector type internally.
type Storage interface {
	Upsert(doc *Document) error
	UpsertBatch(docs []*Document) error
	FindByID(id string) (*Document, error)
	FindByParentID(parentID string) ([]Document, error)
	Delete(id string) error
	DeleteByParentID(parentID string) error
	Search(embedding []float32, filter SearchFilter, opts SearchOptions) ([]SearchResult, error)
}
