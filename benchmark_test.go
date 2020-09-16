package lrucache

import (
	"testing"
)

func BenchmarkRead(b *testing.B) {
	value := []byte(`test`)

	c, err := New("bench-read", 1*KB)
	if err != nil {
		b.Fatalf("failed creating new cache: %v", err)
	}

	_ = c.Set("value", value)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			item, _ := c.Get("value")
			if item == nil {
				b.Fail()
			}
		}
	})
}

func BenchmarkWrite(b *testing.B) {
	value := []byte(`test`)

	c, err := New("bench-write", 1*KB)
	if err != nil {
		b.Fatalf("failed creating new cache: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.Set("value", value)
		}
	})
}
