package cache

import "time"

// cacheEntry represents a single cached item with its expiration time.
// This internal structure tracks both the cached value and when it should expire.
type cacheEntry struct {
	key     string    // The cache key
	value   any       // The actual cached value
	evictAt time.Time // When this entry expires and should be evicted
	index   int
}

// cacheItemsHeap implements a min-heap sorted by evictionTime.
type cacheItemsHeap []*cacheEntry

func (h cacheItemsHeap) Len() int { return len(h) }
func (h cacheItemsHeap) Less(i, j int) bool {
	return h[i].evictAt.Before(h[j].evictAt)
}
func (h cacheItemsHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *cacheItemsHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*cacheEntry)
	item.index = n
	*h = append(*h, item)
}

func (h *cacheItemsHeap) Pop() interface{} {
	old := *h         // save the old heap on the stack
	n := len(old)     // get the length of the old heap from the stack
	x := old[n-1]     // get the last element from the old heap
	x.index = -1      // for safety
	*h = old[0 : n-1] // remove the last element from the old heap
	return x          // return the last element
}
