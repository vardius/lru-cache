package lrucache

import (
	"fmt"
	"hash/maphash"
	"math"
	"math/bits"
	"sync"
)

// maxCapacity is the most entries a Cache can hold: entries are indexed by int32, and index 0 is the list's sentinel.
const maxCapacity = math.MaxInt32 - 1

// Cache holds up to a fixed number of entries and evicts the least recently used one to make room for a new key.
// Get and Set both count as a use.
//
// A Cache is safe for concurrent use by multiple goroutines. Create one with New; the zero value is not usable.
type Cache[K comparable, V any] struct {
	// Set by New and never changed, so read without the lock.
	seed maphash.Seed
	mask uint32 // len(slots) - 1

	// Each method hashes its key before it locks, so the panic of an unhashable key never leaves mu locked.
	// Nothing between Lock and Unlock can panic, so they unlock without defer, which measured a third of a miss.
	mu sync.Mutex
	// entries[1:] are the cached entries, linked in a circular list through prev and next in order of use.
	// entries[0] is the list's sentinel: its next is the most recently used entry and its prev the least.
	// cap(entries) is the capacity plus the sentinel, allocated by New, so adding an entry never allocates.
	entries []entry[K, V]
	// slots index entries by key hash with linear probing. There are at least twice as many slots as the capacity,
	// so a probe always ends at an empty slot after a few steps.
	slots []slot
}

type entry[K comparable, V any] struct {
	key        K
	value      V
	prev, next int32
	hash       uint32 // finds the entry's slot when it is evicted or moved
}

type slot struct {
	hash  uint32 // the low 32 bits of the key's hash, compared before the key
	entry int32  // index in entries; 0 marks an empty slot
}

// New returns an empty Cache that holds up to capacity entries.
// It allocates room for all of them up front, so the cache allocates nothing after New.
// New panics if capacity is below 1 or above math.MaxInt32-1.
func New[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity < 1 || capacity > maxCapacity {
		panic(fmt.Sprintf("lrucache: capacity %d out of range [1, %d]", capacity, maxCapacity))
	}

	// The smallest power of two at least twice the capacity: the table is never more than half full.
	size := 1 << bits.Len(uint(2*capacity-1))

	return &Cache[K, V]{
		seed:    maphash.MakeSeed(),
		mask:    uint32(size - 1),
		entries: make([]entry[K, V], 1, capacity+1),
		slots:   make([]slot, size),
	}
}

// Get returns the value stored for key and true, or the zero value and false if key isn't in the cache.
// Finding key makes it the most recently used entry.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	h := c.hash(key)

	c.mu.Lock()
	_, i := c.find(key, h)
	if i == 0 {
		c.mu.Unlock()
		var zero V
		return zero, false
	}
	c.moveToFront(i)
	value := c.entries[i].value
	c.mu.Unlock()

	return value, true
}

// Set stores value for key and makes key the most recently used entry.
// If key is new and the cache is full, Set first evicts the least recently used entry.
func (c *Cache[K, V]) Set(key K, value V) {
	h := c.hash(key)

	c.mu.Lock()
	s, i := c.find(key, h)
	if i != 0 {
		c.entries[i].value = value
		c.moveToFront(i)
		c.mu.Unlock()
		return
	}

	if len(c.entries) < cap(c.entries) {
		i = int32(len(c.entries))
		c.entries = append(c.entries, entry[K, V]{key: key, value: value, hash: h})
	} else {
		// Full: the least recently used entry makes way, and the new one takes its place.
		i = c.entries[0].prev
		c.unlink(i)
		c.removeSlot(c.slotOf(i))
		c.entries[i] = entry[K, V]{key: key, value: value, hash: h}
		// Removing a slot can empty one earlier in key's probe sequence, where a lookup would stop.
		s = c.freeSlot(h)
	}
	c.slots[s] = slot{hash: h, entry: i}
	c.pushFront(i)
	c.mu.Unlock()
}

// Delete removes key and its value from the cache. It does nothing if key isn't there.
func (c *Cache[K, V]) Delete(key K) {
	h := c.hash(key)

	c.mu.Lock()
	s, i := c.find(key, h)
	if i == 0 {
		c.mu.Unlock()
		return
	}
	c.removeSlot(s)
	c.unlink(i)

	// Keep entries[1:] dense: the last entry moves into the gap.
	last := int32(len(c.entries) - 1)
	if i != last {
		c.slots[c.slotOf(last)].entry = i
		c.entries[i] = c.entries[last]
		c.entries[c.entries[i].prev].next = i
		c.entries[c.entries[i].next].prev = i
	}
	c.entries[last] = entry[K, V]{} // so the cache no longer keeps the key and value alive
	c.entries = c.entries[:last]
	c.mu.Unlock()
}

// Len returns the number of entries in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	n := len(c.entries) - 1
	c.mu.Unlock()

	return n
}

func (c *Cache[K, V]) hash(key K) uint32 {
	return uint32(maphash.Comparable(c.seed, key))
}

// find returns the slot holding key and the index of its entry or, if key is absent,
// the empty slot that ends its probe sequence and 0.
func (c *Cache[K, V]) find(key K, h uint32) (uint32, int32) {
	for s := h & c.mask; ; s = (s + 1) & c.mask {
		sl := c.slots[s]
		if sl.entry == 0 || sl.hash == h && c.entries[sl.entry].key == key {
			return s, sl.entry
		}
	}
}

// freeSlot returns the first empty slot in the probe sequence of hash h.
func (c *Cache[K, V]) freeSlot(h uint32) uint32 {
	s := h & c.mask
	for c.slots[s].entry != 0 {
		s = (s + 1) & c.mask
	}
	return s
}

// slotOf returns the slot of entry i. It matches the index, not the key, so it finds a key that isn't equal to
// itself, such as a NaN.
func (c *Cache[K, V]) slotOf(i int32) uint32 {
	s := c.entries[i].hash & c.mask
	for c.slots[s].entry != i {
		s = (s + 1) & c.mask
	}
	return s
}

// removeSlot empties slot s. A lookup stops at the first empty slot, so a hole would hide the keys probed past it:
// each later slot in the run whose probe sequence passes through the hole moves back into it.
func (c *Cache[K, V]) removeSlot(s uint32) {
	for next := (s + 1) & c.mask; c.slots[next].entry != 0; next = (next + 1) & c.mask {
		home := c.slots[next].hash & c.mask
		// The hole is between home and next, so the key's probe sequence still reaches it there.
		if (next-home)&c.mask >= (next-s)&c.mask {
			c.slots[s] = c.slots[next]
			s = next
		}
	}
	c.slots[s] = slot{}
}

func (c *Cache[K, V]) moveToFront(i int32) {
	if c.entries[0].next == i {
		return
	}
	c.unlink(i)
	c.pushFront(i)
}

func (c *Cache[K, V]) unlink(i int32) {
	e := &c.entries[i]
	c.entries[e.prev].next = e.next
	c.entries[e.next].prev = e.prev
}

func (c *Cache[K, V]) pushFront(i int32) {
	first := c.entries[0].next
	c.entries[i].prev, c.entries[i].next = 0, first
	c.entries[first].prev = i
	c.entries[0].next = i
}
