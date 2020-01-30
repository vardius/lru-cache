package lrucache

import (
	"strconv"
	"testing"
)

func BenchmarkRead(b *testing.B) {
	c := New(10)

	c.Set("1", 1)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if c.Get("1") == nil {
				b.Fail()
			}
		}
	})
}

func BenchmarkWrite(b *testing.B) {
	c := New(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Set(strconv.Itoa(i), i)
	}
}
