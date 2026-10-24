# Changelog

## v1.0.0 (2026-10-24)

The first tagged release: a generic cache with a capacity in entries, and the untagged version's bugs fixed.
See [MIGRATION.md](MIGRATION.md) for the before and after of every call.

### Added

- Type parameters: `Cache[K comparable, V any]` takes any comparable key and any value, and `Get` returns a `V`.
- `Delete` and `Len`.
- Runnable examples in `examples/`, each with a test: caching compiled regular expressions in a server,
  and spreading keys over several caches for many goroutines.

### Changed

- `New[K, V](capacity int)` takes a number of entries and returns a `*Cache[K, V]`; a capacity below 1 panics.
- `Get` returns `(V, bool)`, so a stored `nil` is told apart from a missing key. Nothing returns an error,
  and an empty key is a valid key.
- `Cache` is a struct instead of an interface.
- Go 1.26 or newer is required.
- Faster, with fewer allocations. `New` makes room for every entry, and with string or number keys the cache
  allocates nothing after it: a `Set` that evicts went from 70 ns and 2 allocations to 15 ns and none, and
  filling 1,024 entries from 2,072 allocations to 3. Allocation counts are guarded by tests.

### Fixed

- `Set` of an existing key stores the new value. It used to change a copy, keeping the old value while counting the new one's size.
- A cache filled exactly to its limit keeps every entry. It used to evict once the total reached the limit,
  so a value as large as the limit evicted itself.
- Updating a key to a larger value can't take the cache over its limit.
- A limit of 0 no longer panics with a nil pointer dereference on the first `Set`: `New` rejects it.

### Removed

- The byte budget: `ByteSize` and the `KB` to `YB` constants. It counted only the values' bytes, not the cache's own memory.
- The `name` argument of `New`, which was never used.
