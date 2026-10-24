# Migrating from the untagged version

v1 is the first tagged release. It keeps the module path and what the cache is for, but the API is new:
the cache is generic, its capacity is a number of entries instead of bytes, and nothing returns an error.
The untagged version's bugs are fixed, so a few behaviours change: they are listed below.

## 1. Update the module

```sh
go get github.com/vardius/lru-cache@v1.0.0
```

The import path stays the same:

```go
import lrucache "github.com/vardius/lru-cache"
```

v1 needs Go 1.26 or newer.

## 2. Before and after

```go
// untagged
c, err := lrucache.New("pages", 10*lrucache.MB)
if err != nil {
	return err
}
if err := c.Set("/about", body); err != nil {
	return err
}
body, err := c.Get("/about")
if err != nil {
	return err
}
if body == nil {
	// missing, or stored as nil
}
```

```go
// v1
c := lrucache.New[string, []byte](1000)
c.Set("/about", body)
body, ok := c.Get("/about")
if !ok {
	// missing
}
```

## 3. Every call

| untagged | v1 |
| --- | --- |
| `New(name string, maxSize ByteSize) (Cache, error)` | `New[K comparable, V any](capacity int) *Cache[K, V]` |
| `type Cache interface` | `type Cache[K comparable, V any] struct`, used as a `*Cache` |
| `c.Set(key string, value []byte) error` | `c.Set(key K, value V)` |
| `c.Get(key string) ([]byte, error)` | `c.Get(key K) (V, bool)` |
| none | `c.Delete(key K)` |
| none | `c.Len() int` |
| `type ByteSize float64` and `KB`, `MB`, … `YB` | removed: the capacity is a number of entries |

The name argument is gone: it was required but never used. `New` takes the key and value types as type parameters,
`lrucache.New[string, []byte](1000)` for the untagged version's types.

`Cache` was an interface and is now a struct. If you passed the interface around, declare the methods you use where
you use them, for example `type pages interface { Get(string) ([]byte, bool); Set(string, []byte) }`.
A `*lrucache.Cache[string, []byte]` satisfies it.

## 4. Choosing a capacity

The untagged version took a budget in bytes and counted only `len(value)` against it. v1 counts entries, because it
can't know what a value of any type costs to keep. To move a byte budget over, divide it by the size of a typical value:
a 10 MB budget of values around 10 KB becomes `New[string, []byte](1000)`.

The old budget never bounded memory anyway: keys, list elements and map entries weren't counted. A 1 MB budget of
8-byte values held 131,071 entries in 18.7 MB of heap, keys and values aside. v1 holds as many in 9 MB, and
`New` makes room for all of them at once, so the cache's own memory doesn't grow after `New`.

## 5. Behaviour changes

- **Setting an existing key stores the new value.** The untagged version changed a copy of the entry, so `Get` kept
  returning the first value while the budget counted the new one.
- **A full cache keeps every entry.** The untagged version evicted as soon as the total reached the budget, so a cache
  filled exactly to it lost its oldest entry, and a value as large as the budget evicted everything, itself included.
- **Updating a key never evicts.** The untagged version didn't check the budget on an update, so making a value larger
  could leave the cache over its budget. In v1 an update doesn't change the number of entries.
- **A capacity below 1 panics in `New`**, with `lrucache: capacity 0 out of range [1, 2147483646]`. The untagged
  version accepted a budget of 0 and panicked with a nil pointer dereference on the first `Set`.
- **`Get` tells a missing key from a stored `nil`.** The untagged version returned `nil, nil` for both;
  v1 returns `nil, true` for a stored `nil` and `nil, false` for a missing key.
- **An empty key is a key like any other.** The untagged version rejected `""` with an error.
- **Values are stored as given, as before.** Neither version copies a slice: change its contents after `Set`,
  or after `Get`, and the cached value changes too.
