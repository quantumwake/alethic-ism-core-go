package embedding

import (
	"context"
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/cache"
)

// CachedStorage wraps any Storage implementation with caching.
// Read-by-ID methods are cached; Search is never cached (contextual, stale-sensitive).
type CachedStorage struct {
	*cache.CachedBackend
	base Storage
}

// NewCachedStorage wraps the given Storage with a local cache and default TTL.
func NewCachedStorage(base Storage, ttl time.Duration) *CachedStorage {
	return &CachedStorage{
		CachedBackend: cache.NewCachedBackend(base, nil, ttl),
		base:          base,
	}
}

// NewCachedStorageWithCache wraps the given Storage with a caller-provided cache.
func NewCachedStorageWithCache(base Storage, c cache.Cache, ttl time.Duration) *CachedStorage {
	return &CachedStorage{
		CachedBackend: cache.NewCachedBackend(base, c, ttl),
		base:          base,
	}
}

func (cs *CachedStorage) Upsert(doc *Document) error {
	if err := cs.base.Upsert(doc); err != nil {
		return err
	}
	ctx := context.Background()
	_ = cs.InvalidateMethod(ctx, "FindByID", doc.ID)
	if doc.ParentID != nil {
		_ = cs.InvalidateMethod(ctx, "FindByParentID", *doc.ParentID)
	}
	return nil
}

func (cs *CachedStorage) UpsertBatch(docs []*Document) error {
	if err := cs.base.UpsertBatch(docs); err != nil {
		return err
	}
	ctx := context.Background()
	for _, doc := range docs {
		_ = cs.InvalidateMethod(ctx, "FindByID", doc.ID)
		if doc.ParentID != nil {
			_ = cs.InvalidateMethod(ctx, "FindByParentID", *doc.ParentID)
		}
	}
	return nil
}

func (cs *CachedStorage) FindByID(id string) (*Document, error) {
	return cache.CallCached[*Document](cs.CachedBackend, context.Background(), "FindByID", []interface{}{id}, func() (*Document, error) {
		return cs.base.FindByID(id)
	})
}

func (cs *CachedStorage) FindByParentID(parentID string) ([]Document, error) {
	return cache.CallCached[[]Document](cs.CachedBackend, context.Background(), "FindByParentID", []interface{}{parentID}, func() ([]Document, error) {
		return cs.base.FindByParentID(parentID)
	})
}

func (cs *CachedStorage) Delete(id string) error {
	if err := cs.base.Delete(id); err != nil {
		return err
	}
	_ = cs.InvalidateMethod(context.Background(), "FindByID", id)
	return nil
}

func (cs *CachedStorage) DeleteByParentID(parentID string) error {
	if err := cs.base.DeleteByParentID(parentID); err != nil {
		return err
	}
	_ = cs.InvalidateMethod(context.Background(), "FindByParentID", parentID)
	return nil
}

// Search is never cached — results are contextual and stale results are dangerous for LLM injection.
func (cs *CachedStorage) Search(emb []float32, filter SearchFilter, opts SearchOptions) ([]SearchResult, error) {
	return cs.base.Search(emb, filter, opts)
}
