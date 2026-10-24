package lrucache_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"weak"

	lrucache "github.com/vardius/lru-cache"
)

func TestNewPanicsOnCapacityOutOfRange(t *testing.T) {
	for _, capacity := range []int{0, -1, math.MaxInt32} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) {
			defer func() {
				want := fmt.Sprintf("lrucache: capacity %d out of range [1, 2147483646]", capacity)
				if r := recover(); r != want {
					t.Errorf("New(%d) panicked with %v, want %q", capacity, r, want)
				}
			}()
			lrucache.New[string, int](capacity)
		})
	}
}

func TestGet(t *testing.T) {
	c := lrucache.New[string, []byte](4)
	c.Set("nil", nil)
	c.Set("", []byte("empty key"))

	tests := []struct {
		name   string
		key    string
		want   []byte
		wantOK bool
	}{
		{name: "missing key", key: "missing", want: nil, wantOK: false},
		// The untagged version returned nil, nil for both a stored nil and a missing key.
		{name: "stored nil", key: "nil", want: nil, wantOK: true},
		// The untagged version rejected an empty key with an error.
		{name: "empty key", key: "", want: []byte("empty key"), wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := c.Get(tt.key)
			if string(got) != string(tt.want) || (got == nil) != (tt.want == nil) || ok != tt.wantOK {
				t.Errorf("Get(%q) = %q, %v, want %q, %v", tt.key, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// The untagged version changed a copy of the entry, so Get kept returning the first value.
func TestSetReplacesTheValueOfAnExistingKey(t *testing.T) {
	c := lrucache.New[string, string](2)
	c.Set("a", "old")
	c.Set("a", "new")

	if got, ok := c.Get("a"); got != "new" || !ok {
		t.Errorf(`Get("a") = %q, %v, want "new", true`, got, ok)
	}
	if n := c.Len(); n != 1 {
		t.Errorf("Len() = %d, want 1", n)
	}
}

// The untagged version evicted as soon as its contents reached the limit, so a cache filled exactly to it lost an entry,
// and an entry as large as the limit evicted itself.
func TestFullCacheKeepsEveryEntry(t *testing.T) {
	for _, capacity := range []int{1, 3} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) {
			c := lrucache.New[int, int](capacity)
			for k := range capacity {
				c.Set(k, k)
			}

			if n := c.Len(); n != capacity {
				t.Errorf("Len() = %d, want %d", n, capacity)
			}
			for k := range capacity {
				if v, ok := c.Get(k); v != k || !ok {
					t.Errorf("Get(%d) = %d, %v, want %d, true", k, v, ok, k)
				}
			}
		})
	}
}

// run applies ops to c. Each op is "set k", "get k" or "del k"; set stores the key as its own value.
func run(t *testing.T, c *lrucache.Cache[string, string], ops ...string) {
	t.Helper()
	for _, op := range ops {
		verb, key, _ := strings.Cut(op, " ")
		switch verb {
		case "set":
			c.Set(key, key)
		case "get":
			c.Get(key)
		case "del":
			c.Delete(key)
		default:
			t.Fatalf("unknown op %q", op)
		}
	}
}

// assertOrder checks that c, of the given capacity, holds exactly the keys in want, least recently used first.
// It gets every key in want, then fills c up, sets new keys one at a time and checks that each evicts the next key
// in want.
func assertOrder(t *testing.T, c *lrucache.Cache[string, string], capacity int, want ...string) {
	t.Helper()
	if n := c.Len(); n != len(want) {
		t.Fatalf("Len() = %d, want %d", n, len(want))
	}
	// Getting the keys least recently used first leaves them in the order they were in.
	for _, key := range want {
		if _, ok := c.Get(key); !ok {
			t.Fatalf("Get(%q) found nothing: want the cache to hold %q", key, want)
		}
	}
	for i := len(want); i < capacity; i++ {
		c.Set(fmt.Sprintf("filler-%d", i), "")
	}
	for i, key := range want {
		c.Set(fmt.Sprintf("new-%d", i), "")
		// A miss doesn't count as a use, so it leaves the order alone.
		if _, ok := c.Get(key); ok {
			t.Fatalf("setting new key %d evicted something other than %q: want order %q", i, key, want)
		}
	}
}

func TestEviction(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		ops      []string
		want     []string // least recently used first
	}{
		{"the least recently set key goes first", 2, []string{"set a", "set b", "set c"}, []string{"b", "c"}},
		{"get counts as a use", 2, []string{"set a", "set b", "get a", "set c"}, []string{"a", "c"}},
		{"setting an existing key counts as a use", 2, []string{"set a", "set b", "set a", "set c"}, []string{"a", "c"}},
		{"setting an existing key evicts nothing", 2, []string{"set a", "set b", "set b", "set a"}, []string{"b", "a"}},
		{"a miss doesn't count as a use", 2, []string{"set a", "set b", "get x", "set c"}, []string{"b", "c"}},
		{"one entry: the newest stays", 1, []string{"set a", "set b"}, []string{"b"}},
		{"a deleted key frees its place", 2, []string{"set a", "set b", "del a", "set c"}, []string{"b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := lrucache.New[string, string](tt.capacity)
			run(t, c, tt.ops...)
			assertOrder(t, c, tt.capacity, tt.want...)
		})
	}
}

// The references are the start of the classic page replacement example, here with four frames.
// want is what the cache holds after each reference.
func TestEvictionSequence(t *testing.T) {
	refs := []string{"7", "0", "1", "2", "0", "3", "0", "4", "2", "3", "0", "3", "2"}
	want := [][]string{
		{"7"},
		{"7", "0"},
		{"7", "0", "1"},
		{"7", "0", "1", "2"},
		{"7", "1", "2", "0"},
		{"1", "2", "0", "3"},
		{"1", "2", "3", "0"},
		{"2", "3", "0", "4"},
		{"3", "0", "4", "2"},
		{"0", "4", "2", "3"},
		{"4", "2", "3", "0"},
		{"4", "2", "0", "3"},
		{"4", "0", "3", "2"},
	}
	for i := range refs {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := lrucache.New[string, string](4)
			for _, ref := range refs[:i+1] {
				if _, ok := c.Get(ref); !ok {
					c.Set(ref, ref)
				}
			}
			assertOrder(t, c, 4, want[i]...)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name string
		ops  []string
		want []string // least recently used first
	}{
		{"a missing key", []string{"set a", "set b", "del x"}, []string{"a", "b"}},
		{"the only key", []string{"set a", "del a"}, nil},
		{"the least recently used key", []string{"set a", "set b", "set c", "del a"}, []string{"b", "c"}},
		{"the most recently used key", []string{"set a", "set b", "set c", "del c"}, []string{"a", "b"}},
		{"a key in the middle", []string{"set a", "set b", "set c", "del b"}, []string{"a", "c"}},
		{"twice", []string{"set a", "set b", "del a", "del a"}, []string{"b"}},
		{"then refill", []string{"set a", "set b", "set c", "del a", "del c", "set d", "set e"}, []string{"b", "d", "e"}},
		{"every key, then refill", []string{"set a", "set b", "set c", "del b", "del a", "del c", "set d", "set e", "set f", "set g"}, []string{"e", "f", "g"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := lrucache.New[string, string](3)
			run(t, c, tt.ops...)
			for _, op := range tt.ops {
				if key, ok := strings.CutPrefix(op, "del "); ok && !slices.Contains(tt.want, key) {
					if _, found := c.Get(key); found {
						t.Errorf("Get(%q) found a deleted key", key)
					}
				}
			}
			assertOrder(t, c, 3, tt.want...)
		})
	}
}

// storeNew stores a new value for key and returns a weak pointer to it, so the caller holds no strong reference.
func storeNew(c *lrucache.Cache[string, *[64]byte], key string) weak.Pointer[[64]byte] {
	v := new([64]byte)
	c.Set(key, v)
	return weak.Make(v)
}

func TestCacheReleasesValuesItNoLongerHolds(t *testing.T) {
	tests := []struct {
		name   string
		remove func(c *lrucache.Cache[string, *[64]byte])
		gone   []string // keys whose first value the cache no longer holds
	}{
		{"deleted", func(c *lrucache.Cache[string, *[64]byte]) { c.Delete("b") }, []string{"b"}},
		// b moves into a's place, then goes too: the place b left must not keep it either.
		{"deleted, after moving", func(c *lrucache.Cache[string, *[64]byte]) { c.Delete("a"); c.Delete("b") }, []string{"a", "b"}},
		{"evicted", func(c *lrucache.Cache[string, *[64]byte]) { c.Set("c", nil) }, []string{"a"}},
		{"replaced", func(c *lrucache.Cache[string, *[64]byte]) { c.Set("a", nil) }, []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := lrucache.New[string, *[64]byte](2)
			values := map[string]weak.Pointer[[64]byte]{"a": storeNew(c, "a"), "b": storeNew(c, "b")}

			tt.remove(c)
			runtime.GC()

			for key, value := range values {
				gone := slices.Contains(tt.gone, key)
				if alive := value.Value() != nil; alive == gone {
					t.Errorf("%s's first value alive = %v, want %v", key, alive, !gone)
				}
			}
			runtime.KeepAlive(c) // so the values it holds stay reachable through it
		})
	}
}

// A NaN is never equal to itself: as with a map, every Set adds an entry that no Get finds. Eviction must still work.
func TestNaNKeys(t *testing.T) {
	nan := math.NaN()
	c := lrucache.New[float64, string](2)
	c.Set(nan, "a")
	c.Set(nan, "b")

	if n := c.Len(); n != 2 {
		t.Errorf("Len() = %d, want 2", n)
	}
	if v, ok := c.Get(nan); ok {
		t.Errorf("Get(NaN) = %q, true, want a miss", v)
	}
	c.Delete(nan)

	c.Set(1, "c")
	c.Set(2, "d")
	if n := c.Len(); n != 2 {
		t.Errorf("Len() = %d after evicting both NaNs, want 2", n)
	}
	for k, want := range map[float64]string{1: "c", 2: "d"} {
		if v, ok := c.Get(k); v != want || !ok {
			t.Errorf("Get(%v) = %q, %v, want %q, true", k, v, ok, want)
		}
	}
}

// Comparing interface keys can panic, as with a map. The cache must stay usable after the panic is recovered.
func TestUnhashableKeyPanicsWithoutLockingTheCache(t *testing.T) {
	c := lrucache.New[any, int](2)
	for name, call := range map[string]func(){
		"Get":    func() { c.Get([]int{1}) },
		"Set":    func() { c.Set([]int{1}, 1) },
		"Delete": func() { c.Delete([]int{1}) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("%s with a slice key didn't panic", name)
				}
			}()
			call()
		})
	}

	// These would block forever if a panic had left the cache locked.
	c.Set("a", 1)
	if v, ok := c.Get("a"); v != 1 || !ok {
		t.Errorf(`Get("a") = %d, %v, want 1, true`, v, ok)
	}
}

// model is a slow, obviously correct LRU cache to check Cache against.
type model[K comparable] struct {
	capacity int
	keys     []K // least recently used first
	values   map[K]int
}

func (m *model[K]) touch(key K) {
	m.keys = append(slices.DeleteFunc(m.keys, func(k K) bool { return k == key }), key)
}

func (m *model[K]) get(key K) (int, bool) {
	v, ok := m.values[key]
	if ok {
		m.touch(key)
	}
	return v, ok
}

func (m *model[K]) set(key K, value int) {
	if _, ok := m.values[key]; !ok && len(m.keys) == m.capacity {
		delete(m.values, m.keys[0])
		m.keys = m.keys[1:]
	}
	m.values[key] = value
	m.touch(key)
}

func (m *model[K]) delete(key K) {
	delete(m.values, key)
	m.keys = slices.DeleteFunc(m.keys, func(k K) bool { return k == key })
}

// checkAgainstModel runs random operations on a Cache and on the model, and compares every result.
// Small key spaces make keys share slots of the index and get evicted, deleted and set again often.
func checkAgainstModel[K comparable](t *testing.T, capacity int, key func(r *rand.Rand) K) {
	c := lrucache.New[K, int](capacity)
	m := &model[K]{capacity: capacity, values: map[K]int{}}
	r := rand.New(rand.NewPCG(uint64(capacity), 1))

	for op := range 20_000 {
		k := key(r)
		switch n := r.IntN(10); {
		case n < 4:
			c.Set(k, op)
			m.set(k, op)
		case n < 5:
			c.Delete(k)
			m.delete(k)
		default:
			got, ok := c.Get(k)
			want, wantOK := m.get(k)
			if got != want || ok != wantOK {
				t.Fatalf("op %d: Get(%T(%v)) = %d, %v, want %d, %v", op, k, k, got, ok, want, wantOK)
			}
		}
		if got, want := c.Len(), len(m.keys); got != want {
			t.Fatalf("op %d: Len() = %d, want %d", op, got, want)
		}
	}
}

func TestAgainstModel(t *testing.T) {
	for _, capacity := range []int{1, 2, 3, 5, 8, 13, 64, 100} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) {
			checkAgainstModel(t, capacity, func(r *rand.Rand) int { return r.IntN(3 * capacity) })
		})
	}
}

// Interface keys holding equal numbers of different integer types are different keys with the same hash,
// so these collide in full: the cache must tell them apart by the key, and find their slots by the entry.
func TestAgainstModelWithCollidingKeys(t *testing.T) {
	variants := []func(n int) any{
		func(n int) any { return n },
		func(n int) any { return int64(n) },
		func(n int) any { return uint64(n) },
		func(n int) any { return int32(n) },
	}
	for _, capacity := range []int{1, 2, 3, 5, 8, 13, 64} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) {
			checkAgainstModel(t, capacity, func(r *rand.Rand) any {
				return variants[r.IntN(len(variants))](r.IntN(capacity))
			})
		})
	}
}

func TestConcurrentUse(t *testing.T) {
	const capacity = 64
	c := lrucache.New[int, int](capacity)

	var wg sync.WaitGroup
	for g := range 8 {
		wg.Go(func() {
			r := rand.New(rand.NewPCG(uint64(g), 2))
			for range 5_000 {
				key := r.IntN(4 * capacity)
				switch r.IntN(4) {
				case 0:
					c.Delete(key)
				case 1:
					c.Set(key, -key)
				default:
					if v, ok := c.Get(key); ok && v != -key {
						t.Errorf("Get(%d) = %d, want %d", key, v, -key)
					}
				}
				if n := c.Len(); n > capacity {
					t.Errorf("Len() = %d, over the capacity of %d", n, capacity)
				}
			}
		})
	}
	wg.Wait()
}
