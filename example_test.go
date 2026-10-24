package lrucache_test

import (
	"fmt"

	lrucache "github.com/vardius/lru-cache"
)

func Example() {
	// Room for two entries: a third evicts the one used least recently.
	c := lrucache.New[string, int](2)
	c.Set("tea", 450)
	c.Set("cake", 300)
	c.Get("tea") // tea is now the most recently used
	c.Set("coffee", 500)

	_, ok := c.Get("cake")
	fmt.Println("cake still cached:", ok)
	price, ok := c.Get("tea")
	fmt.Println("tea:", price, ok)
	// Output:
	// cake still cached: false
	// tea: 450 true
}

func ExampleCache_Get() {
	c := lrucache.New[string, int](10)
	c.Set("stock", 0)

	// ok tells a stored zero value from a key that isn't there.
	stock, ok := c.Get("stock")
	fmt.Println(stock, ok)
	missing, ok := c.Get("missing")
	fmt.Println(missing, ok)
	// Output:
	// 0 true
	// 0 false
}

func ExampleCache_Set() {
	c := lrucache.New[string, string](2)
	c.Set("a", "first")
	c.Set("b", "second")
	c.Set("a", "replaced") // an existing key: its value changes and nothing is evicted

	v, _ := c.Get("a")
	fmt.Println(v, c.Len())
	// Output: replaced 2
}

func ExampleCache_Delete() {
	c := lrucache.New[string, string](2)
	c.Set("session", "alice")
	c.Delete("session")

	_, ok := c.Get("session")
	fmt.Println(ok, c.Len())
	// Output: false 0
}
