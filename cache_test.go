package lrucache

import (
	"bytes"
	"container/list"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	c, err := New("test-new", 10)
	if err != nil {
		t.Fatalf("failed creating new cache: %v", err)
	}

	if c == nil {
		t.Fatalf("cache is nil")
	}
}

func TestMaxSize(t *testing.T) {
	c, err := New("test-limit", 10)
	if err != nil {
		t.Fatalf("failed creating new cache: %v", err)
	}

	for i := 1; i <= 20; i++ {
		key := strconv.Itoa(i)
		value := []byte(key)

		if err := c.Set(key, value); err != nil {
			t.Fatal(err)
			return
		}

		item, err := c.Get(key)
		if err != nil {
			t.Fatal(err)
			return
		}

		if i > 10 && item == nil {
			t.Fatalf("item not found for i:%d", i)
		}
	}

	if len(c.(*cache).elements) > 10 {
		t.Fatalf("cache has more then 10 elements")
	}
}

func TestSequence(t *testing.T) {
	c, err := New("test-sequence", CacheSizeMB)
	if err != nil {
		t.Fatalf("failed creating new cache: %v", err)
	}

	expected := map[int][]int{
		0:  {7},
		1:  {7, 0},
		2:  {7, 0, 1},
		3:  {7, 0, 1, 2},
		4:  {7, 1, 2, 0},
		5:  {1, 2, 0, 3},
		6:  {1, 2, 3, 0},
		7:  {2, 3, 0, 4},
		8:  {3, 0, 4, 2},
		9:  {0, 4, 2, 3},
		10: {4, 2, 3, 0},
		11: {4, 2, 0, 3},
		12: {4, 0, 3, 2},
	}

	getKeys := func(m map[string]*list.Element) []string {
		keys := make([]string, len(m))
		i := 0
		for k := range m {
			keys[i] = k
			i++
		}

		return keys
	}

	for index, v := range []int{7, 0, 1, 2, 0, 3, 0, 4, 2, 3, 0, 3, 2} {
		key := strconv.Itoa(v)
		value := []byte(key)

		if err := c.Set(key, value); err != nil {
			t.Fatal(err)
			return
		}

		for _, want := range expected[index] {
			wantKey := strconv.Itoa(want)
			wantValue := []byte(wantKey)

			got, _ := c.Get(wantKey)

			if got == nil || bytes.Compare(wantValue, got) != 0 {
				t.Fatalf("want (%v), got (%v), not in %v", wantValue, got, getKeys(c.(*cache).elements))
			}
		}
	}
}
