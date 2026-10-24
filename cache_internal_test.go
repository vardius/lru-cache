package lrucache

import (
	"strconv"
	"testing"
)

// The index's size is what the documented 16 to 32 bytes per entry rest on: the smallest power of two at least
// twice the capacity.
func TestIndexSize(t *testing.T) {
	tests := []struct{ capacity, slots int }{
		{1, 2},
		{2, 4},
		{3, 8},
		{4, 8},
		{5, 16},
		{1024, 2048},
		{1025, 4096},
	}
	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.capacity), func(t *testing.T) {
			if got := len(New[string, int](tt.capacity).slots); got != tt.slots {
				t.Errorf("New(%d) made %d slots, want %d", tt.capacity, got, tt.slots)
			}
		})
	}
}
