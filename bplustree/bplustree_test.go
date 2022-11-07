package bplustree

import (
	"fmt"
	"strconv"
	"testing"
)

func TestSortSearch(t *testing.T) {
	cases := []struct {
		a      []int
		search int
		expect int
	}{
		{a: []int{1, 2, 3, 4}, search: 2, expect: 1},
		{a: []int{1, 2, 4}, search: 3, expect: 2},
		{a: []int{2}, search: 2, expect: 0},
		{a: []int{2}, search: 3, expect: 0},
		{a: []int{2}, search: 1, expect: 0},
	}
	less := func(a, b int) bool { return a < b }

	for i, c := range cases {
		n := strconv.Itoa(i)
		t.Run(n, func(t *testing.T) {
			input := items[int](c.a)
			got, found := input.find(c.search, less)
			fmt.Println(got, found, c.expect)
		})
	}
}
