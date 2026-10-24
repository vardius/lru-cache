package main

import "testing"

func TestShardedHoldsWhatFits(t *testing.T) {
	s := newSharded[int, string](4, 1000)
	for k := range 100 {
		s.Set(k, "v")
	}
	for k := range 100 {
		if _, ok := s.Get(k); !ok {
			t.Errorf("Get(%d) missed in shards with room for 250 entries each", k)
		}
	}
}

func TestShardedEvictsWithinEachShard(t *testing.T) {
	s := newSharded[int, int](4, 4) // one entry per shard
	for k := range 100 {
		s.Set(k, k)
	}

	held := 0
	for k := range 100 {
		if v, ok := s.Get(k); ok {
			held++
			if v != k {
				t.Errorf("Get(%d) = %d", k, v)
			}
		}
	}
	if held == 0 || held > 4 {
		t.Errorf("4 shards of 1 hold %d entries, want 1 to 4", held)
	}
}

func TestReadThrough(t *testing.T) {
	// One key: each goroutine misses it at most once, the first time, before some goroutine has set it.
	const goroutines, lookups = 4, 1000
	hits := readThrough(newSharded[int, int](2, 2), goroutines, lookups, 1)
	if hits < goroutines*(lookups-1) || hits >= goroutines*lookups {
		t.Errorf("%d hits, want between %d and %d", hits, goroutines*(lookups-1), goroutines*lookups-1)
	}
}
