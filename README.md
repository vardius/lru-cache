🗃️ lru-cache
================
[![Build Status](https://travis-ci.org/vardius/lru-cache.svg?branch=master)](https://travis-ci.org/vardius/lru-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/lru-cache)](https://goreportcard.com/report/github.com/vardius/lru-cache)
[![codecov](https://codecov.io/gh/vardius/lru-cache/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/lru-cache)
[![](https://godoc.org/github.com/vardius/lru-cache?status.svg)](https://pkg.go.dev/github.com/vardius/lru-cache)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/lru-cache/blob/master/LICENSE.md)

<img align="right" height="180px" src="https://github.com/vardius/gorouter/blob/master/website/src/static/img/logo.png?raw=true" alt="logo" />

Go simple LRU in memory cache

A Least Recently Used (LRU) Cache organizes items in order of use, allowing you to quickly identify which item hasn't been used for the longest amount of time.

* Strengths:
	- Super fast accesses. LRU caches store items in order from most-recently used to least-recently used. That means both can be accessed in O(1)O(1) time.
	- Super fast updates. Each time an item is accessed, updating the cache takes O(1) time.

📖 ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/lru-cache/issues) to manage them.

## 📚 Documentation

For __examples__ **visit [godoc#pkg-examples](http://godoc.org/github.com/vardius/lru-cache#pkg-examples)**

For **GoDoc** reference, **visit [pkg.go.dev](https://pkg.go.dev/github.com/vardius/lru-cache)**

🚏 HOW TO USE
==================================================

## 🚅 Benchmark
**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  lru-cache git:(master) go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/vardius/lru-cache
BenchmarkRead-4         23051137                50.1 ns/op             0 B/op          0 allocs/op
BenchmarkWrite-4        23097510                51.8 ns/op             0 B/op          0 allocs/op
PASS
ok  	github.com/vardius/lru-cache	2.713s
```

## 🏫 Basic example
```go
package main

import (
	"fmt"
	"log"

    lrucache "github.com/vardius/lru-cache"
)

func main() {
	c, err := lrucache.New("example-cache", 10*lrucache.MB)
	if err != nil {
		log.Fatal(err)
		return
	}

	item, err := c.Get("test")
	if err != nil {
		log.Fatal(err)
		return
	}

	if item == nil {
		if err = c.Set("test", []byte("value")); err != nil {
			log.Fatal(err)
			return
		}
	}

	got, err := c.Get("test")
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(string(got))

	// Output:
	// value
}
```

📜 [License](LICENSE.md)
-------

This package is released under the MIT license. See the complete license in the package
