package windowing

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/state"
	"strings"
	"time"
)

// FormatKeyDefinitions formats key definitions for logging
func FormatKeyDefinitions(keyDefs state.ColumnKeyDefinitions) string {
	parts := make([]string, len(keyDefs))
	for i, keyDef := range keyDefs {
		parts[i] = keyDef.Name
	}
	return strings.Join(parts, ", ")
}

// FormatKeyDefinitionsWithValues formats key definitions with their values for logging
func FormatKeyDefinitionsWithValues(keyDefs state.ColumnKeyDefinitions, data map[string]interface{}) string {
	parts := make([]string, len(keyDefs))
	for i, keyDef := range keyDefs {
		if val, ok := data[keyDef.Name]; ok {
			parts[i] = fmt.Sprintf("%s=%v", keyDef.Name, val)
		} else {
			parts[i] = fmt.Sprintf("%s=<missing>", keyDef.Name)
		}
	}
	return strings.Join(parts, ", ")
}

// FormatStatistics formats statistics for logging
func FormatStatistics(stats *Statistics) string {
	if stats == nil {
		return "no stats"
	}

	avgTime := time.Duration(stats.Avg())
	return fmt.Sprintf("count=%d, avg=%v, total=%v",
		stats.Count(), avgTime, time.Duration(stats.Sum()))
}

// LogBlockEviction logs detailed information when a block is evicted
func LogBlockEviction(prefix string, blk *Block, keyDefs state.ColumnKeyDefinitions,
	currentBlocks int, softLimit int) string {

	totalParts := 0
	sourceCount := len(blk.partsBySource)
	for _, parts := range blk.partsBySource {
		totalParts += len(parts)
	}

	keyDefStr := FormatKeyDefinitions(keyDefs)

	return fmt.Sprintf("[%s] Soft evicting block - Key: %s | KeyDefs: [%s] | Sources: %d | Parts: %d | EvictionTime: %v | BlockCount: %d/%d",
		prefix, blk.key, keyDefStr, sourceCount, totalParts,
		blk.evictionTime.Format(time.RFC3339), currentBlocks-1, softLimit)
}

// LogCombineOperation logs detailed information about a combine operation (join or merge)
func LogCombineOperation(combineKey string, keyDefs state.ColumnKeyDefinitions, combineResult map[string]interface{},
	storedSourceID, inboundSourceID string, storedPart, inboundPart *BlockPart,
	maxJoinCount int, avgTime time.Duration) string {

	keyDefStr := FormatKeyDefinitionsWithValues(keyDefs, combineResult)

	return fmt.Sprintf("[BlockStore] Combine completed - Key: %s | KeyValues: [%s] | Sources: %s+%s | StoredPartJoinCount: %d/%d | InboundPartJoinCount: %d/%d | AvgTime: %.6fs",
		combineKey, keyDefStr, storedSourceID, inboundSourceID,
		storedPart.JoinCount, maxJoinCount,
		inboundPart.JoinCount, maxJoinCount,
		avgTime.Seconds())
}

// LogPartSkipped logs information about skipped parts during combine
func LogPartSkipped(combineKey, sourceID string, skippedExpired, skippedMaxJoins, kept int,
	maxAge time.Duration, maxJoinCount int) string {

	return fmt.Sprintf("[BlockStore] Parts skipped - Key: %s | Source: %s | Expired: %d (maxAge: %v) | MaxJoins: %d (limit: %d) | Kept: %d",
		combineKey, sourceID, skippedExpired, maxAge, skippedMaxJoins, maxJoinCount, kept)
}

// LogNewPartAdded logs when a new part is added to a block
func LogNewPartAdded(combineKey, sourceID string, existingParts, totalSources int,
	expireAt time.Time, maxAge time.Duration) string {

	return fmt.Sprintf("[BlockStore] New part added - Key: %s | Source: %s | ExistingParts: %d | TotalSources: %d | PartExpiry: %v | MaxAge: %v",
		combineKey, sourceID, existingParts, totalSources,
		expireAt.Format(time.RFC3339), maxAge)
}

// LogNewBlockCreated logs when a new block is created
func LogNewBlockCreated(combineKey string, evictionTime time.Time, windowTTL time.Duration,
	totalBlocks, softLimit int) string {

	return fmt.Sprintf("[BlockStore] New block created - Key: %s | EvictionTime: %v | WindowTTL: %v | TotalBlocks: %d/%d",
		combineKey, evictionTime.Format(time.RFC3339), windowTTL, totalBlocks, softLimit)
}

// LogBlockStoreCreated logs when a new BlockStore is created
func LogBlockStoreCreated(keyDefs state.ColumnKeyDefinitions, blockCountSoftLimit, blockPartMaxJoinCount int,
	blockWindowTTL, blockPartMaxAge time.Duration) string {

	keyDefStr := FormatKeyDefinitions(keyDefs)

	return fmt.Sprintf("[BlockStore] Created new store - Keys: [%s] | BlockCountSoftLimit: %d | BlockWindowTTL: %v | PartMaxJoinCount: %d | PartMaxAge: %v",
		keyDefStr, blockCountSoftLimit, blockWindowTTL, blockPartMaxJoinCount, blockPartMaxAge)
}

// LogBlockStoreShutdown logs when a BlockStore is shutting down
func LogBlockStoreShutdown(keyDefs state.ColumnKeyDefinitions, blockCount, totalParts, totalSources int,
	stats *Statistics) string {

	keyDefStr := FormatKeyDefinitions(keyDefs)
	statsStr := FormatStatistics(stats)

	return fmt.Sprintf("[BlockStore] Shutting down - Keys: [%s] | ActiveBlocks: %d | TotalParts: %d | UniqueSources: %d | Stats: %s",
		keyDefStr, blockCount, totalParts, totalSources, statsStr)
}
