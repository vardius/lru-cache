package lrucache_test

import (
	"fmt"
	"strconv"

	lrucache "github.com/vardius/lru-cache"
)

func Example() {
	c := lrucache.New(10)

	for i := 1; i <= 20; i++ {
		c.Set(strconv.Itoa(i), i)
	}

	fmt.Println(c.Get("1"))
	fmt.Println(c.Get("11"))
	fmt.Println(c.Get("20"))

	// Output:
	// <nil>
	// 11
	// 20
}

func Example_second() {
	c := lrucache.New(2)

	item := c.Get("test")

	if item == nil {
		c.Set("test", 10)
	}

	fmt.Println(c.Get("test"))

	// Output:
	// 10
}
