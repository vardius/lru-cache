package lrucache_test

import (
	"testing"

	lrucache "github.com/vardius/lru-cache"
)

// TestAllocations guards the allocations per call. Lower a number when an optimisation lands;
// never raise one without a reason in the commit.
func TestAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector changes allocation counts")
	}

	const capacity = 1024
	keys := makeKeys("key", 4*capacity)
	value := []byte("value")

	fill := func() *lrucache.Cache[string, []byte] {
		c := lrucache.New[string, []byte](capacity)
		for _, k := range keys[:capacity] {
			c.Set(k, value)
		}
		return c
	}
	hits, existing, evicting, deleting := fill(), fill(), fill(), fill()
	roomy := lrucache.New[string, []byte](len(keys)) // never full: AllocsPerRun makes 1001 calls
	boxed := lrucache.New[any, int](capacity)

	n := 0
	tests := []struct {
		name string
		call func()
		want float64
	}{
		{"Get, hit", func() { hits.Get(keys[n%capacity]); n++ }, 0},
		{"Get, miss", func() { hits.Get("missing") }, 0},
		{"Set, existing key", func() { existing.Set(keys[n%capacity], value); n++ }, 0},
		// Cycling through three times as many keys as fit, each Set is of a key evicted long before.
		{"Set, new key, full cache", func() { evicting.Set(keys[capacity+n%(3*capacity)], value); n++ }, 0},
		{"Set, new key, room left", func() { roomy.Set(keys[n%len(keys)], value); n++ }, 0},
		{"Delete and Set again", func() { k := keys[n%capacity]; deleting.Delete(k); deleting.Set(k, value); n++ }, 0},
		{"Len", func() { hits.Len() }, 0},
		// The Cache, its entries and its index, all for the full capacity: nothing after this allocates.
		{"New", func() { lrucache.New[string, []byte](capacity) }, 3},
		// Hashing makes a key escape when its type holds pointers other than strings, so an int boxed into an
		// interface key is moved to the heap. Ints below 256 and constants are boxed without allocating.
		{"Get, int boxed into an any key", func() { boxed.Get(1000 + n); n++ }, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testing.AllocsPerRun(1000, tt.call); got != tt.want {
				t.Errorf("allocates %v times per call, want %v", got, tt.want)
			}
		})
	}
}
