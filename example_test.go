package lrucache_test

import (
	"fmt"
	"log"

	lrucache "github.com/vardius/lru-cache"
)

func Example() {
	c, err := lrucache.New("example-cache", lrucache.CacheSizeMB)
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
