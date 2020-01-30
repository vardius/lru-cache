package lrucache

import (
	"testing"
)

func BenchmarkRead(b *testing.B) {
	c := New(10)

	c.Set("value", 1)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if c.Get("value") == nil {
				b.Fail()
			}
		}
	})
}

func BenchmarkWrite(b *testing.B) {
	c := New(b.N)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set("value", 1)
		}
	})
}
