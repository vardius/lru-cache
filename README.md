Vardius - lru-cache
================
[![Build Status](https://travis-ci.org/vardius/lru-cache.svg?branch=master)](https://travis-ci.org/vardius/lru-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/lru-cache)](https://goreportcard.com/report/github.com/vardius/lru-cache)
[![codecov](https://codecov.io/gh/vardius/lru-cache/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/lru-cache)
[![](https://godoc.org/github.com/vardius/lru-cache?status.svg)](http://godoc.org/github.com/vardius/lru-cache)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/lru-cache/blob/master/LICENSE.md)

Go simple LRU in memory cache

A Least Recently Used (LRU) Cache organizes items in order of use, allowing you to quickly identify which item hasn't been used for the longest amount of time.

* Strengths:
	- Super fast accesses. LRU caches store items in order from most-recently used to least-recently used. That means both can be accessed in O(1)O(1) time.
	- Super fast updates. Each time an item is accessed, updating the cache takes O(1) time.

ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/lru-cache/issues) to manage them.

HOW TO USE
==================================================

1. [GoDoc](http://godoc.org/github.com/vardius/lru-cache)

## Benchmark
**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  gocache go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/vardius/lru-cache
Benchmark-4      5000000               358 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/vardius/lru-cache    2.159s
```

## Basic example
```go
package main

import (
	"fmt"

    lrucache "github.com/vardius/lru-cache"
)

func main() {
	c := lrucache.New(2)

	item := c.Get("test")

	if item == nil {
		c.Set("test", 10)
	}

	fmt.Println(c.Get("test"))

	// Output:
	// 10
}
```

License
-------

This package is released under the MIT license. See the complete license in the package:

[LICENSE](LICENSE.md)
