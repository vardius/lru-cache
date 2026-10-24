package lrucache_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	lrucache "github.com/vardius/lru-cache"
)

// Every benchmark runs on a cache of 1,024 string keys and byte slice values. Keys are made up front,
// so what is measured is the cache's own cost.
const benchCapacity = 1024

var (
	benchKeys    = makeKeys("key", 2*benchCapacity) // the first benchCapacity fill the cache
	benchMissing = makeKeys("missing", benchCapacity)
	benchOrder   = rand.New(rand.NewPCG(1, 2)).Perm(benchCapacity) // hits in no particular order
	benchValue   = []byte("value")
)

func makeKeys(prefix string, n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("%s-%06d", prefix, i)
	}
	return keys
}

func benchFull() *lrucache.Cache[string, []byte] {
	c := lrucache.New[string, []byte](benchCapacity)
	for _, k := range benchKeys[:benchCapacity] {
		c.Set(k, benchValue)
	}
	return c
}

func BenchmarkGet(b *testing.B) {
	c := benchFull()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		c.Get(benchKeys[benchOrder[i%benchCapacity]])
		i++
	}
}

func BenchmarkGetMiss(b *testing.B) {
	c := benchFull()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		c.Get(benchMissing[i%benchCapacity])
		i++
	}
}

func BenchmarkSetExisting(b *testing.B) {
	c := benchFull()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		c.Set(benchKeys[benchOrder[i%benchCapacity]], benchValue)
		i++
	}
}

// Cycling through twice as many keys as fit, in order, every Set is of a key evicted 1,024 Sets before.
func BenchmarkSetNew(b *testing.B) {
	c := benchFull()
	b.ReportAllocs()
	i := benchCapacity
	for b.Loop() {
		c.Set(benchKeys[i%len(benchKeys)], benchValue)
		i++
	}
}

func BenchmarkFill(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchFull()
	}
}

// A read-through cache under load: GOMAXPROCS goroutines get random keys out of twice as many as fit,
// and set each one they miss. About half the gets hit.
func BenchmarkParallel(b *testing.B) {
	r := rand.New(rand.NewPCG(3, 4))
	picks := make([]int, 1<<16)
	for i := range picks {
		picks[i] = r.IntN(len(benchKeys))
	}
	c := benchFull()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := rand.IntN(len(picks))
		for pb.Next() {
			k := benchKeys[picks[i%len(picks)]]
			if _, ok := c.Get(k); !ok {
				c.Set(k, benchValue)
			}
			i++
		}
	})
}
