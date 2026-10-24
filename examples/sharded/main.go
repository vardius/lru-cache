// Command sharded spreads keys over several caches, so goroutines that use different keys rarely wait for the same
// lock. It times the same lookups on one cache and on one split into 16 shards.
//
//	go run ./examples/sharded
//
// Each shard evicts on its own: what goes first is the least recently used key of its shard, not of the whole cache,
// and one shard can be full while another has room. It pays off when many goroutines spend most of their time
// in the cache.
package main

import (
	"fmt"
	"hash/maphash"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	lrucache "github.com/vardius/lru-cache"
)

type cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
}

// sharded is a cache split into shards by the hash of the key.
type sharded[K comparable, V any] struct {
	seed   maphash.Seed
	shards []*lrucache.Cache[K, V]
}

// newSharded returns n shards that hold at least capacity entries between them.
func newSharded[K comparable, V any](n, capacity int) *sharded[K, V] {
	s := &sharded[K, V]{seed: maphash.MakeSeed(), shards: make([]*lrucache.Cache[K, V], n)}
	for i := range s.shards {
		s.shards[i] = lrucache.New[K, V]((capacity + n - 1) / n)
	}
	return s
}

func (s *sharded[K, V]) shard(key K) *lrucache.Cache[K, V] {
	return s.shards[maphash.Comparable(s.seed, key)%uint64(len(s.shards))]
}

func (s *sharded[K, V]) Get(key K) (V, bool) { return s.shard(key).Get(key) }
func (s *sharded[K, V]) Set(key K, value V)  { s.shard(key).Set(key, value) }

// readThrough makes lookups from several goroutines at once. Each gets a random key out of keys and sets the
// keys it misses. It returns how many gets hit.
func readThrough(c cache[int, int], goroutines, lookups, keys int) int {
	var (
		mu   sync.Mutex
		hits int
		wg   sync.WaitGroup
	)
	for g := range goroutines {
		wg.Go(func() {
			r := rand.New(rand.NewPCG(uint64(g), 0))
			n := 0 // counted per goroutine: one shared counter would be a contended lock of its own
			for range lookups {
				k := r.IntN(keys)
				if _, ok := c.Get(k); ok {
					n++
					continue
				}
				c.Set(k, k*k)
			}
			mu.Lock()
			hits += n
			mu.Unlock()
		})
	}
	wg.Wait()

	return hits
}

func main() {
	const capacity, keys, lookups = 10_000, 20_000, 1_000_000
	goroutines := runtime.GOMAXPROCS(0)

	for _, layout := range []struct {
		name  string
		cache cache[int, int]
	}{
		{"one cache", lrucache.New[int, int](capacity)},
		{"16 shards", newSharded[int, int](16, capacity)},
	} {
		start := time.Now()
		hits := readThrough(layout.cache, goroutines, lookups, keys)
		total := goroutines * lookups
		fmt.Printf("%s: %d lookups from %d goroutines in %v, %d%% hits\n",
			layout.name, total, goroutines, time.Since(start).Round(time.Millisecond), 100*hits/total)
	}
}
