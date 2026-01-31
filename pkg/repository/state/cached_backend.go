package state

import (
	"context"
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/cache"
	"gorm.io/gorm"
)

// CachedBackendStorage provides a caching layer over the state BackendStorage.
// It intercepts all read operations and caches their results to reduce database load.
// Write operations automatically invalidate relevant cache entries to maintain consistency.
// This implementation uses the generic cache package, making it easy to switch cache backends.
type CachedBackendStorage struct {
	*cache.CachedBackend                 // Embedded generic caching functionality
	base                 *BackendStorage // The underlying state backend
}

// DefaultConfig returns the default TTL configuration for state backend.
func DefaultConfig(baseTTL time.Duration) *cache.MethodTTLConfig {
	config := cache.NewMethodTTLConfig(baseTTL)

	// State data changes less frequently
	config.SetMethodTTL("FindState", 5*time.Minute)
	config.SetMethodTTL("FindStateFull", 5*time.Minute)

	// Column definitions are relatively static
	config.SetMethodTTL("FindDataColumnDefinitionsByStateID", 10*time.Minute)

	// Column data is accessed frequently but changes occasionally
	config.SetMethodTTL("FindDataRowColumnDataByColumnID", baseTTL)

	// Config attributes and key definitions change less frequently
	config.SetMethodTTL("FindConfigAttributes", 5*time.Minute)
	config.SetMethodTTL("FindStateConfigKeyDefinitionsGroupByDefinitionType", 5*time.Minute)

	return config
}

// NewCachedBackend creates a new state backend with caching enabled.
// Uses the default state configuration with provided base TTL.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation
//   - baseTTL: Base TTL for cache entries
//
// Returns:
//   - A new CachedBackendStorage instance with caching enabled
func NewCachedBackend(dsn string, c cache.Cache, baseTTL time.Duration) *CachedBackendStorage {
	config := DefaultConfig(baseTTL)
	return NewCachedBackendWithConfig(dsn, c, config)
}

// NewCachedBackendWithConfig creates a state backend with custom TTL configuration.
// Use this for fine-grained control over method-specific cache TTLs.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation
//   - config: Method-specific TTL configuration
//
// Returns:
//   - A new CachedBackendStorage instance with configured TTLs
//
// Example:
//
//	config := cache.DefaultStateConfig(30*time.Second)
//	backend := NewCachedBackendWithConfig(dsn, cache, config)
func NewCachedBackendWithConfig(dsn string, c cache.Cache, config *cache.MethodTTLConfig) *CachedBackendStorage {
	base := NewBackend(dsn)
	cachedBackend := cache.NewCachedBackend(base, c, config.DefaultTTL)

	// Apply method-specific TTL configuration
	config.ApplyToBackend(cachedBackend)

	return &CachedBackendStorage{
		CachedBackend: cachedBackend,
		base:          base,
	}
}

// FindState retrieves a state by ID with caching.
// This is one of the most frequently called methods, so caching provides significant benefits.
// Cache key format: "FindState:hash(id)"
//
// Parameters:
//   - id: The ID of the state to find
//
// Returns:
//   - *State: The state if found
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindState(id string) (*State, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindState", []interface{}{id},
		func() (*State, error) {
			return cb.base.FindState(id)
		})
}

// FindStateFull retrieves a state with all associated data with caching.
// Results are cached to avoid repeated database queries for complete state data.
// Cache key format: "FindStateFull:hash(id, flags)"
//
// Parameters:
//   - id: The state ID to find
//   - flags: Flags controlling what data to load
//
// Returns:
//   - *State: The state with requested data loaded
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindStateFull(id string, flags StateLoadFlags) (*State, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindStateFull", []interface{}{id, flags},
		func() (*State, error) {
			return cb.base.FindStateFull(id, flags)
		})
}

// FindDataRowColumnDataByColumnID retrieves column data by column ID with caching.
// Column data doesn't change frequently, making it ideal for caching.
// Cache key format: "FindDataRowColumnDataByColumnID:hash(columnID)"
//
// Parameters:
//   - id: The column ID to find data for
//
// Returns:
//   - *DataRowColumnData: The column data if found
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindDataRowColumnDataByColumnID(id *int64) (*DataRowColumnData, error) {
	ctx := context.Background()

	// Convert pointer to value for consistent cache keys
	var idVal int64
	if id != nil {
		idVal = *id
	}

	return cache.CallCached(cb.CachedBackend, ctx, "FindDataRowColumnDataByColumnID", []interface{}{idVal},
		func() (*DataRowColumnData, error) {
			return cb.base.FindDataRowColumnDataByColumnID(id)
		})
}

// FindDataColumnDefinitionsByStateID retrieves column definitions for a state with caching.
// Column definitions are relatively static configuration.
// Cache key format: "FindDataColumnDefinitionsByStateID:hash(stateID)"
//
// Parameters:
//   - id: The state ID to find column definitions for
//
// Returns:
//   - Columns: Map of column name to definition
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindDataColumnDefinitionsByStateID(id string) (Columns, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindDataColumnDefinitionsByStateID", []interface{}{id},
		func() (Columns, error) {
			return cb.base.FindDataColumnDefinitionsByStateID(id)
		})
}

// FindConfigAttributes retrieves config attributes for a state with caching.
// Config attributes don't change frequently, making them ideal for caching.
// Cache key format: "FindConfigAttributes:hash(stateID)"
//
// Parameters:
//   - stateID: The state ID to find config attributes for
//
// Returns:
//   - ConfigAttributes: The config attributes for the state
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindConfigAttributes(stateID string) (ConfigAttributes, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindConfigAttributes", []interface{}{stateID},
		func() (ConfigAttributes, error) {
			return cb.base.FindConfigAttributes(stateID)
		})
}

// FindStateConfigKeyDefinitionsGroupByDefinitionType retrieves key definitions grouped by type with caching.
// Key definitions are static configuration that rarely changes.
// Cache key format: "FindStateConfigKeyDefinitionsGroupByDefinitionType:hash(stateID)"
//
// Parameters:
//   - stateID: The state ID to find key definitions for
//
// Returns:
//   - TypedColumnKeyDefinitions: Key definitions grouped by type
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindStateConfigKeyDefinitionsGroupByDefinitionType(stateID string) (TypedColumnKeyDefinitions, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindStateConfigKeyDefinitionsGroupByDefinitionType", []interface{}{stateID},
		func() (TypedColumnKeyDefinitions, error) {
			return cb.base.FindStateConfigKeyDefinitionsGroupByDefinitionType(stateID)
		})
}

// UpsertState inserts or updates a state and invalidates related cache entries.
// This ensures that subsequent reads will get the updated data from the database.
// Cache invalidation is precise - only affected entries are removed.
//
// Parameters:
//   - state: The state to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindState cache for this state
//   - Invalidates FindStateFull cache for this state
func (cb *CachedBackendStorage) UpsertState(state *State) error {
	// Perform the database operation first
	err := cb.base.UpsertState(state)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Invalidate cache entries that would be affected by this change
	_ = cb.InvalidateMethod(ctx, "FindState", state.ID)
	// Invalidate all FindStateFull entries for this state (different flag combinations)
	_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", state.ID)

	return nil
}

// UpsertStateComplete inserts or updates a state with all related data and invalidates cache.
// This method handles state, attributes, and other related data in a transaction.
//
// Parameters:
//   - state: The complete state to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindState cache
//   - Invalidates FindStateFull cache
//   - Invalidates FindConfigAttributes cache
//   - Invalidates FindStateConfigKeyDefinitionsGroupByDefinitionType cache
func (cb *CachedBackendStorage) UpsertStateComplete(state *State) error {
	// Perform the database operation first
	err := cb.base.UpsertStateComplete(state)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Invalidate all cache entries that could be affected by this change
	_ = cb.InvalidateMethod(ctx, "FindState", state.ID)
	_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", state.ID)
	_ = cb.InvalidateMethod(ctx, "FindConfigAttributes", state.ID)
	_ = cb.InvalidateMethod(ctx, "FindStateConfigKeyDefinitionsGroupByDefinitionType", state.ID)

	return nil
}

// UpsertStateColumns inserts or updates column definitions and invalidates cache.
// Column changes affect multiple cache entries related to state structure.
//
// Parameters:
//   - columns: Map of column definitions to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindDataColumnDefinitionsByStateID cache for affected states
//   - Invalidates FindStateFull cache for affected states
func (cb *CachedBackendStorage) UpsertStateColumns(columns Columns) error {
	// Perform the database operation first
	err := cb.base.UpsertStateColumns(columns)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Collect unique state IDs from the columns
	stateIDs := make(map[string]bool)
	for _, column := range columns {
		stateIDs[column.StateID] = true
	}

	// Invalidate cache for each affected state
	for stateID := range stateIDs {
		_ = cb.InvalidateMethod(ctx, "FindDataColumnDefinitionsByStateID", stateID)
		_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", stateID)
	}

	return nil
}

// DeleteStateColumns deletes all column definitions for a state and invalidates cache.
// This operation affects the structure of the state data.
//
// Parameters:
//   - stateID: The state ID whose columns should be deleted
//
// Returns:
//   - int: Number of rows affected
//
// Cache invalidation:
//   - Invalidates FindDataColumnDefinitionsByStateID cache
//   - Invalidates FindStateFull cache
//   - Invalidates FindDataRowColumnDataByColumnID cache for all deleted columns
func (cb *CachedBackendStorage) DeleteStateColumns(stateID string) int {
	// Perform the database operation
	affected := cb.base.DeleteStateColumns(stateID)

	ctx := context.Background()

	// Invalidate cache entries that would be affected by this change
	_ = cb.InvalidateMethod(ctx, "FindDataColumnDefinitionsByStateID", stateID)
	_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", stateID)
	// Note: We can't easily invalidate FindDataRowColumnDataByColumnID for specific columns
	// without first querying them, so we might want to consider a broader invalidation

	return affected
}

// DeleteStateColumn deletes a specific column definition and invalidates cache.
// This operation affects the structure of the state data.
//
// Parameters:
//   - id: The column ID to delete
//
// Returns:
//   - bool: True if a row was deleted
//
// Cache invalidation:
//   - Invalidates FindDataRowColumnDataByColumnID cache for this column
//   - Note: We don't have the state ID here, so we can't invalidate state-specific caches
func (cb *CachedBackendStorage) DeleteStateColumn(id int64) bool {
	// Perform the database operation
	deleted := cb.base.DeleteStateColumn(id)

	if deleted {
		ctx := context.Background()

		// Invalidate cache for this specific column
		_ = cb.InvalidateMethod(ctx, "FindDataRowColumnDataByColumnID", id)
		// Note: Ideally we'd also invalidate state-related caches, but we don't have the state ID
	}

	return deleted
}

// UpsertConfigAttribute inserts or updates a single config attribute and invalidates cache.
// Config attributes are part of the state configuration.
//
// Parameters:
//   - attribute: The config attribute to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindConfigAttributes cache for the state
//   - Invalidates FindStateFull cache for the state
func (cb *CachedBackendStorage) UpsertConfigAttribute(attribute *ConfigAttribute) error {
	// Perform the database operation first
	err := cb.base.UpsertConfigAttribute(attribute)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Invalidate cache entries that would be affected by this change
	_ = cb.InvalidateMethod(ctx, "FindConfigAttributes", attribute.StateID)
	_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", attribute.StateID)

	return nil
}

// UpsertConfigAttributes inserts or updates multiple config attributes and invalidates cache.
// This is more efficient than calling UpsertConfigAttribute multiple times.
//
// Parameters:
//   - attributes: The config attributes to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindConfigAttributes cache for affected states
//   - Invalidates FindStateFull cache for affected states
func (cb *CachedBackendStorage) UpsertConfigAttributes(attributes ConfigAttributes) error {
	// Perform the database operation first
	err := cb.base.UpsertConfigAttributes(attributes)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Collect unique state IDs from the attributes
	stateIDs := make(map[string]bool)
	for _, attr := range attributes {
		stateIDs[attr.StateID] = true
	}

	// Invalidate cache for each affected state
	for stateID := range stateIDs {
		_ = cb.InvalidateMethod(ctx, "FindConfigAttributes", stateID)
		_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", stateID)
	}

	return nil
}

// DeleteConfigAttributes deletes config attributes for a state and invalidates cache.
// This removes all configuration attributes for the specified state.
//
// Parameters:
//   - stateID: The state ID whose attributes should be deleted
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindConfigAttributes cache
//   - Invalidates FindStateFull cache
func (cb *CachedBackendStorage) DeleteConfigAttributes(stateID string) error {
	// Perform the database operation first
	err := cb.base.DeleteConfigAttributes(stateID)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Invalidate cache entries that would be affected by this change
	_ = cb.InvalidateMethod(ctx, "FindConfigAttributes", stateID)
	_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", stateID)

	return nil
}

// FindStateConfigKeyDefinitions retrieves all key definitions for a state with caching.
// Key definitions are relatively static configuration.
// Cache key format: "FindStateConfigKeyDefinitions:hash(stateID)"
//
// Parameters:
//   - stateID: The state ID to find key definitions for
//
// Returns:
//   - ColumnKeyDefinitions: List of key definitions
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindStateConfigKeyDefinitions(stateID string) (ColumnKeyDefinitions, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindStateConfigKeyDefinitions", []interface{}{stateID},
		func() (ColumnKeyDefinitions, error) {
			return cb.base.FindStateConfigKeyDefinitions(stateID)
		})
}

// FindStateConfigKeyDefinitionsByType retrieves key definitions by type with caching.
// This is an optimized query for specific definition types.
// Cache key format: "FindStateConfigKeyDefinitionsByType:hash(stateID, definitionType)"
//
// Parameters:
//   - stateID: The state ID to find key definitions for
//   - definitionType: The specific type of definitions to retrieve
//
// Returns:
//   - ColumnKeyDefinitions: List of key definitions of the specified type
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindStateConfigKeyDefinitionsByType(stateID string, definitionType DefinitionType) (ColumnKeyDefinitions, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindStateConfigKeyDefinitionsByType",
		[]interface{}{stateID, definitionType},
		func() (ColumnKeyDefinitions, error) {
			return cb.base.FindStateConfigKeyDefinitionsByType(stateID, definitionType)
		})
}

// UpsertStateConfigKeyDefinitions inserts or updates key definitions and invalidates cache.
// Key definitions define the structure and constraints of state data.
//
// Parameters:
//   - definitions: The key definitions to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindStateConfigKeyDefinitions cache
//   - Invalidates FindStateConfigKeyDefinitionsGroupByDefinitionType cache
//   - Invalidates FindStateConfigKeyDefinitionsByType cache
//   - Invalidates FindStateFull cache
func (cb *CachedBackendStorage) UpsertStateConfigKeyDefinitions(definitions []*ColumnKeyDefinition) error {
	// Perform the database operation first
	err := cb.base.UpsertStateConfigKeyDefinitions(definitions)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Collect unique state IDs and definition types
	stateIDs := make(map[string]bool)
	typesByState := make(map[string]map[DefinitionType]bool)

	for _, def := range definitions {
		stateIDs[def.StateID] = true
		if typesByState[def.StateID] == nil {
			typesByState[def.StateID] = make(map[DefinitionType]bool)
		}
		typesByState[def.StateID][def.DefinitionType] = true
	}

	// Invalidate cache for each affected state
	for stateID := range stateIDs {
		_ = cb.InvalidateMethod(ctx, "FindStateConfigKeyDefinitions", stateID)
		_ = cb.InvalidateMethod(ctx, "FindStateConfigKeyDefinitionsGroupByDefinitionType", stateID)
		_ = cb.InvalidateMethodPrefix(ctx, "FindStateFull", stateID)

		// Invalidate specific type queries
		if types, ok := typesByState[stateID]; ok {
			for defType := range types {
				_ = cb.InvalidateMethod(ctx, "FindStateConfigKeyDefinitionsByType", stateID, defType)
			}
		}
	}

	return nil
}

// RunTransactionIsolation runs a function within a database transaction.
// This method bypasses caching and operates directly on the database.
// Use this for complex operations that require transactional consistency.
//
// Parameters:
//   - fn: Function to execute within the transaction
//
// Returns:
//   - error: Error from the function or transaction handling
//
// Note: Cache invalidation must be handled manually for operations within the transaction
func (cb *CachedBackendStorage) RunTransactionIsolation(fn func(db *gorm.DB) error) error {
	return cb.base.RunTransactionIsolation(fn)
}

// Access returns the underlying BackendStorage for direct database access.
// Use this when you need to bypass the cache layer, for example:
//   - During data migrations
//   - For administrative operations
//   - When debugging cache issues
//
// Returns:
//   - *BackendStorage: The underlying database backend
func (cb *CachedBackendStorage) Access() *BackendStorage {
	return cb.base
}
