package windowing

import (
	"container/heap"
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/state"
	"log"
	"sync"
	"time"
)

// BlockStore is the upper-level structure that defines key fields,
// holds blocks, and performs combine and eviction of entire blocks and individual parts
type BlockStore struct {
	KeyDefinitions state.ColumnKeyDefinitions // fields defining the correlation key

	// combineFunc defines how two BlockParts from different sources are combined
	combineFunc CombineFunc

	// stop watch for measuring performance of the store
	Statistics *Statistics

	// blocks provides fast lookup by key.
	blocks map[string]*Block
	// heap orders the blocks by evictionTime.
	heap blockHeap

	mu sync.Mutex

	// Block management configuration
	blockCountSoftLimit int           // if total blocks exceed this, eviction window will apply
	blockWindowTTL      time.Duration // sliding window TTL for a block, resets on each new event

	// BlockPart management configuration
	blockPartMaxJoinCount int           // hard limit on how many times a part can be joined
	blockPartMaxAge       time.Duration // absolute lifetime of a part from creation

	// Lifecycle management
	shutdownCh   chan struct{}
	lastAccessed time.Time
}

type KeyedBlock map[string]*Block

// NewBlockStore creates a new BlockStore with a pluggable CombineFunc.
func NewBlockStore(
	keyDefinitions state.ColumnKeyDefinitions,
	combineFunc CombineFunc,
	blockCountSoftLimit, blockPartMaxJoinCount int,
	blockWindowTTL, blockPartMaxAge time.Duration,
) *BlockStore {
	store := &BlockStore{
		KeyDefinitions:        keyDefinitions,
		combineFunc:           combineFunc,
		blocks:                make(KeyedBlock),
		heap:                  blockHeap{},
		blockCountSoftLimit:   blockCountSoftLimit,
		blockPartMaxJoinCount: blockPartMaxJoinCount,
		blockWindowTTL:        blockWindowTTL,
		blockPartMaxAge:       blockPartMaxAge,
		Statistics:            NewStopWatch().Start(),
		shutdownCh:            make(chan struct{}),
		lastAccessed:          time.Now(),
	}

	log.Print(LogBlockStoreCreated(keyDefinitions, blockCountSoftLimit, blockPartMaxJoinCount, blockWindowTTL, blockPartMaxAge))

	heap.Init(&store.heap)
	go store.evictionLoop()
	return store
}

// GetKeyValue builds a unique key for an event based on the store's KeyDefinitions.
func (store *BlockStore) GetKeyValue(event models.Data) (string, error) {
	key := ""
	for _, field := range store.KeyDefinitions {
		value, ok := event[field.Name]
		if !ok {
			return "", fmt.Errorf("field `%s` not present in event", field.Name)
		}
		key += fmt.Sprintf("%v|", value)
	}
	return key, nil
}

func (store *BlockStore) EvictExpiredBlocks() {
	now := time.Now()
	for store.heap.Len() > 0 {
		if store.heap[0].evictionTime.After(now) {
			break
		}
		heap.Pop(&store.heap)
	}
}

func (store *BlockStore) GetOrAddBlock(keyValue string) (*Block, error) {
	if block, ok := store.blocks[keyValue]; ok {
		return block, nil
	}

	now := time.Now()
	block := &Block{
		key:           keyValue,
		partsBySource: make(PartsBySource),
		evictionTime:  now.Add(store.blockWindowTTL),
		heapIndex:     -1,
	}

	store.blocks[keyValue] = block
	heap.Push(&store.heap, block)

	log.Print(LogNewBlockCreated(keyValue, block.evictionTime, store.blockWindowTTL,
		len(store.blocks), store.blockCountSoftLimit))

	return block, nil
}

// AddData processes an incoming event from a given source. It only combines events from different sources.
func (store *BlockStore) AddData(inboundSourceID string, inboundSourceData models.Data, callback func(data models.Data) error) error {
	stopWatch := NewStopWatch().Start()
	defer func() {
		elapsed := stopWatch.Stop().Elapsed()
		store.Statistics.LapWith(elapsed)
	}()

	store.mu.Lock()
	defer store.mu.Unlock()

	// Update last accessed time
	store.lastAccessed = time.Now()

	// get the key value for the inbound data
	keyValue, err := store.GetKeyValue(inboundSourceData)
	if err != nil {
		return fmt.Errorf("could not get key value for source data %v: %v", inboundSourceData, err)
	}

	now := time.Now()

	// Within the store, we maintain a map of blocks - the key is derived from the key definition.
	block, _ := store.GetOrAddBlock(keyValue)

	// we track the inbound data by wrapping it in a block part
	inboundSourcePart := &BlockPart{
		Data:      inboundSourceData,
		ExpireAt:  now.Add(store.blockPartMaxAge),
		JoinCount: 0,
	}

	// store the new inbound part
	block.partsBySource[inboundSourceID] = append(block.partsBySource[inboundSourceID], inboundSourcePart)

	// Log the new part addition
	existingParts := len(block.partsBySource[inboundSourceID]) - 1
	totalSourcesInBlock := len(block.partsBySource)
	log.Print(LogNewPartAdded(keyValue, inboundSourceID, existingParts, totalSourcesInBlock,
		inboundSourcePart.ExpireAt, store.blockPartMaxAge))

	// under the block, we separate out the arrival data by source
	// this allows us to combine the received data (on source) against all other sources
	for storedSourceID, storedParts := range block.partsBySource {
		if inboundSourceID == storedSourceID {
			continue // do not combine events from the same source
		}

		write := 0 // in-place compaction position
		skippedExpired := 0
		skippedMaxJoins := 0
		for _, storedPart := range storedParts {
			expired := storedPart.ExpireAt.Before(now)
			maxJoinsReached := storedPart.JoinCount >= store.blockPartMaxJoinCount

			if expired || maxJoinsReached {
				if expired {
					skippedExpired++
				}
				if maxJoinsReached {
					skippedMaxJoins++
				}
				continue
			}

			storedParts[write] = storedPart

			// call the pluggable combine function
			combineResult, combineErr := store.combineFunc(
				storedSourceID,
				storedPart,
				inboundSourceID,
				inboundSourcePart,
				store.KeyDefinitions)
			if combineErr != nil {
				return fmt.Errorf("combine error: %v", combineErr)
			}

			log.Print(LogCombineOperation(keyValue, store.KeyDefinitions, combineResult,
				storedSourceID, inboundSourceID, storedPart, inboundSourcePart,
				store.blockPartMaxJoinCount, time.Duration(store.Statistics.Avg())))
			if err = callback(combineResult); err != nil {
				return fmt.Errorf("could not process part: %v", err)
			}

			write++
		}

		// Log if parts were skipped
		if skippedExpired > 0 || skippedMaxJoins > 0 {
			log.Print(LogPartSkipped(keyValue, storedSourceID, skippedExpired, skippedMaxJoins,
				write, store.blockPartMaxAge, store.blockPartMaxJoinCount))
		}

		// compact the list
		for index := write; index < len(storedParts); index++ {
			storedParts[index] = nil
		}
		block.partsBySource[storedSourceID] = storedParts[:write]
	}

	// Always reset the block eviction time on each new event (sliding window)
	block.evictionTime = now.Add(store.blockWindowTTL)
	heap.Fix(&store.heap, block.heapIndex)
	return nil
}

// Shutdown stops the eviction loop and cleans up resources
func (store *BlockStore) Shutdown() {
	store.mu.Lock()
	blockCount := len(store.blocks)
	totalParts := 0
	totalSources := 0
	sourceMap := make(map[string]int)

	for _, block := range store.blocks {
		for sourceID, parts := range block.partsBySource {
			totalParts += len(parts)
			sourceMap[sourceID] += len(parts)
		}
	}
	totalSources = len(sourceMap)
	store.mu.Unlock()

	log.Print(LogBlockStoreShutdown(store.KeyDefinitions, blockCount, totalParts, totalSources, store.Statistics))
	close(store.shutdownCh)
}

// IsIdle returns true if the store hasn't been accessed for longer than the idle duration
func (store *BlockStore) IsIdle(idleDuration time.Duration) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	return time.Since(store.lastAccessed) > idleDuration
}

// evictionLoop runs periodically to remove stale blocks.
// If the total number of blocks exceeds the soft threshold, blocks whose evictionTime has passed are evicted.
func (store *BlockStore) evictionLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	evictFn := func() {
		now := time.Now()

		store.mu.Lock()
		defer store.mu.Unlock()

		for store.heap.Len() > 0 {
			if len(store.blocks) <= store.blockCountSoftLimit {
				return
			}

			blk := store.heap[0]
			if blk.evictionTime.Before(now) {
				heap.Pop(&store.heap)
				delete(store.blocks, blk.key)

				log.Print(LogBlockEviction("BlockStore", blk, store.KeyDefinitions,
					len(store.blocks), store.blockCountSoftLimit))
			} else {
				break
			}
		}
	}

	for {
		select {
		case <-ticker.C:
			evictFn()
		case <-store.shutdownCh:
			return
		}
	}
}
