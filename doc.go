// Package lrucache is an in-memory cache that holds a fixed number of entries and,
// when it is full, evicts the one used least recently to make room for a new key.
//
//	pages := lrucache.New[string, []byte](1000)
//	pages.Set("/about", body)
//	body, ok := pages.Get("/about")
//
// A Cache is generic: K is any comparable type and V any type, so values come back typed, with no assertions.
// Get reports with a bool whether the key was there, so a stored zero value is told apart from a missing key.
//
// # Eviction
//
// The capacity passed to New is a number of entries. Get and Set both make their key the most recently used;
// a Get that misses changes nothing. When Set adds a new key to a full cache, the least recently used entry goes first.
// Setting a key that is already cached replaces its value and evicts nothing. Delete frees an entry's place at once.
//
// The cache counts entries, not bytes: it can't know what a value of any type costs to keep. To bound memory,
// pick a capacity from the size of a typical value. The cache doesn't copy values: a stored slice, map or pointer
// shares its contents with the caller.
//
// # Concurrency
//
// A Cache is safe for concurrent use. Every method takes one lock, Get included, since a hit changes the order of use.
// Goroutines that do little besides calling the cache wait for each other; spreading keys over several caches,
// as the repository's sharded example does, gives each its own lock, at the price of evicting by recency within each
// cache only.
//
// # Cost
//
// New makes three allocations: the Cache, room for every entry, and the index that finds them. After that the cache
// allocates nothing: Get, Set and Delete work in that room, and Set on a full cache puts the new entry where the
// evicted one was. Entries live in one slice, linked in order of use by int32 indices, and the index is a hash table
// of those indices, so the garbage collector has one slice to scan rather than an object per entry.
//
// One cost can land on the caller: hashing makes a key escape to the heap when its type holds pointers other than
// strings. A key that is a pointer, an interface or a struct holding either must then live on the heap, so on a
// Cache[any, V], Get(n) allocates to box an int n of 256 or more.
package lrucache
