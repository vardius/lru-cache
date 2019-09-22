package lrucache

import (
	"fmt"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	c := New(10)

	if c == nil {
		t.Fail()
	}
}

func TestLimit(t *testing.T) {
	c := New(10)

	for i := 1; i <= 20; i++ {
		c.Set(strconv.Itoa(i), i)

		if i > 10 && c.Get(strconv.Itoa(i-10)) != nil {
			t.Fail()
		}
	}

	if len(c.(*cache).items) > 10 {
		t.Fail()
	}
}

func TestSequence(t *testing.T) {
	c := New(4)

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

	getKeys := func(m map[string]*item) []string {
		keys := make([]string, len(m))
		i := 0
		for k := range m {
			keys[i] = k
			i++
		}

		return keys
	}

	for index, value := range []int{7, 0, 1, 2, 0, 3, 0, 4, 2, 3, 0, 3, 2} {
		c.Set(strconv.Itoa(value), value)

		if len(c.(*cache).items) != len(expected[index]) {
			t.Fail()
		}

		fmt.Printf("%v\n", getKeys(c.(*cache).items))

		for _, want := range expected[index] {
			got := c.Get(strconv.Itoa(want))

			if want != got {
				t.Fatalf("want (%v), got (%v), not in %v", want, got, getKeys(c.(*cache).items))
			}
		}
	}
}
