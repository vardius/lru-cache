# 🗃️ lru-cache

[![Test](https://github.com/vardius/lru-cache/actions/workflows/test.yml/badge.svg)](https://github.com/vardius/lru-cache/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vardius/lru-cache.svg)](https://pkg.go.dev/github.com/vardius/lru-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/lru-cache)](https://goreportcard.com/report/github.com/vardius/lru-cache)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

A least recently used (LRU) cache for Go: it holds a fixed number of entries and, when it is full,
evicts the one used longest ago to make room for a new key.

- **Typed.** `Cache[K, V]` takes any comparable key and any value, and gives the value back as a `V`.
- **Exact LRU order.** `Get` and `Set` both count as a use; the entry that goes is always the least recently used.
- **No allocations after `New`**, for keys of strings, numbers and structs of them. Room for every entry is made
  up front, and a new key takes the evicted one's place. Tests guard every allocation count.
- **Safe for concurrent use**, behind one lock.
- **Small API.** `New`, `Get`, `Set`, `Delete` and `Len`, and no dependencies beyond the standard library.

📖 **[Documentation](https://vardius.github.io/lru-cache/)** · [API reference](https://pkg.go.dev/github.com/vardius/lru-cache) · [Examples](examples) · [Migrating from the untagged version](MIGRATION.md)

## Install

```sh
go get github.com/vardius/lru-cache
```

Requires Go 1.26 or newer.

## Usage

```go
import lrucache "github.com/vardius/lru-cache"

pages := lrucache.New[string, []byte](1000) // room for 1,000 entries

pages.Set("/about", body)

if body, ok := pages.Get("/about"); ok {
	w.Write(body)
}

pages.Delete("/about")
```

`Get` returns the value and whether the key was there, so a stored zero value or `nil` isn't mistaken for a miss.
Setting a key that is already cached replaces its value and evicts nothing.

## Eviction

The capacity is a number of entries. When `Set` adds a new key to a full cache, the least recently used entry goes first:

```go
c := lrucache.New[string, int](2)
c.Set("tea", 450)
c.Set("cake", 300)
c.Get("tea")         // tea is now the most recently used
c.Set("coffee", 500) // evicts cake
```

The cache counts entries, not bytes, because it can't know what a value of any type costs to keep.
To bound memory, pick a capacity from the size of a typical value. The untagged version's byte budget counted only
the values' bytes: a 1 MB budget of 8-byte values held 131,071 entries in 18.7 MB of heap, keys and values aside.
v1 holds as many in 9 MB.

Values aren't copied: a stored slice, map or pointer shares its contents with the caller.

## Concurrency

Every method takes one lock, `Get` included, since a hit changes the order of use. When many goroutines do little
besides calling the cache, they wait on that lock; [examples/sharded](examples/sharded/main.go) spreads keys over
several caches, each with its own lock, at the price of evicting by recency within each cache only.

## Benchmarks

Time and allocations per call, on an Apple M4 Max with Go 1.27.0, median of 5 runs. Both versions ran on the same harness:
a cache of 1,024 entries, string keys and byte slice values. In the read-through row, 16 goroutines get random keys out
of twice as many as fit and set the ones they miss; its time is wall time per call across all 16, so it measures
throughput. Run v1's own suite with `go test -bench .`.

| Scenario | untagged | v1 |
| --- | ---: | ---: |
| Get, hit | 15.0 ns · 0 allocs | **15.5 ns · 0 allocs** |
| Get, miss | 8.5 ns · 0 allocs | **9.3 ns · 0 allocs** |
| Set, existing key | 17.7 ns · 0 allocs | **15.0 ns · 0 allocs** |
| Set, new key, evicting | 69.9 ns · 2 allocs | **15.0 ns · 0 allocs** |
| New and fill 1,024 entries | 68.6 µs · 2,072 allocs | **18.8 µs · 3 allocs** |
| Read-through, 16 goroutines at once | 391 ns · 0.9 allocs | **253 ns · 0 allocs** |

A miss costs a little more than the untagged version's map lookup: the cache hashes keys into an index of its own,
so a new key can take the evicted entry's place without allocating. That makes an evicting `Set` 4.7 times faster, with no allocations,
and the index takes 16 to 32 bytes per entry, where a `map[string]int32` of the same keys takes about 55.

## Upgrading from the untagged version

`New` takes a number of entries and type parameters, and nothing returns an error:

```go
// untagged
c, err := lrucache.New("pages", 10*lrucache.MB)
if err != nil {
	return err
}
if err := c.Set("/about", body); err != nil {
	return err
}
body, err := c.Get("/about") // nil, nil when missing

// v1
c := lrucache.New[string, []byte](1000)
c.Set("/about", body)
body, ok := c.Get("/about")
```

[MIGRATION.md](MIGRATION.md) maps every old call to its v1 form and lists the behaviour changes.

## Contributing

Bug reports and pull requests are welcome: see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

MIT, see [LICENSE.md](LICENSE.md).
