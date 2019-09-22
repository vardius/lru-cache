package lrucache

import (
	"testing"
)

func Benchmark(b *testing.B) {
	c := New(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set("1", 1)

			if c.Get("1") == nil {
				b.Fail()
			}
		}
	})
}
