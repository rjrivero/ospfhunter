package main

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ringTestCase struct {
	label    string
	ringSize int
	input    []int
	contents []int
	results  []int
}

func push(index int) int {
	return index
}

func pop() int {
	return math.MinInt32
}

func TestRing(t *testing.T) {

	testCases := []ringTestCase{
		{
			label:    "Empty ring should return no data",
			ringSize: 5,
			input:    []int{},
			contents: []int{},
			results:  []int{},
		},
		{
			label:    "Half ring should return less than <size> values",
			ringSize: 5,
			input:    []int{push(3), push(2), push(1)},
			contents: []int{3, 2, 1},
			results:  []int{0, 0, 0},
		},
		{
			label:    "Full size ring should skip no items",
			ringSize: 5,
			input:    []int{push(3), push(2), push(1), push(4), push(5)},
			contents: []int{3, 2, 1, 4, 5},
			results:  []int{0, 0, 0, 0, 0},
		},
		{
			label:    "Overflown ring should return latest items",
			ringSize: 5,
			input:    []int{push(3), push(2), push(1), push(4), push(5), push(6)},
			contents: []int{2, 1, 4, 5, 6},
			results:  []int{0, 0, 0, 0, 0, 3},
		},
		{
			label:    "Overflown ring (twice) should return latest items",
			ringSize: 2,
			input:    []int{push(3), push(2), push(1), push(4), push(5), push(6)},
			contents: []int{5, 6},
			results:  []int{0, 0, 3, 2, 1, 4},
		},
		{
			label:    "Iterates properly when wrapped (tail > head)",
			ringSize: 3,
			input:    []int{push(3), push(2), pop(), push(4)},
			contents: []int{2, 4},
			results:  []int{0, 0, 3, 0},
		},
		{
			label:    "Pop should remove items",
			ringSize: 5,
			input:    []int{push(5), push(2), push(1), push(4), pop(), pop()},
			contents: []int{1, 4},
			results:  []int{0, 0, 0, 0, 5, 2},
		},
		{
			label:    "Can not pop more items than pushed",
			ringSize: 5,
			input:    []int{push(5), pop(), pop()},
			contents: []int{},
			results:  []int{0, 5, 0},
		},
		{
			label:    "Can pop after wrapping",
			ringSize: 3,
			input:    []int{push(3), push(4), push(5), push(6), pop(), push(8)},
			contents: []int{5, 6, 8},
			results:  []int{0, 0, 0, 3, 4, 0},
		},
		{
			label:    "Can pop past empty, and push again",
			ringSize: 3,
			input:    []int{push(3), push(4), push(5), push(6), pop(), pop(), pop(), pop(), push(8)},
			contents: []int{8},
			results:  []int{0, 0, 0, 3, 4, 5, 6, 0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, tc.run)
	}
}

func (tc *ringTestCase) run(t *testing.T) {
	result := make([]int, 0, len(tc.contents))
	remain := make([]int, 0, len(tc.results))
	tested := makeIntRing(tc.ringSize)
	for _, item := range tc.input {
		var evicted int
		if item > math.MinInt32 {
			evicted = tested.Push(item)
		} else {
			evicted = tested.Pop()
		}
		remain = append(remain, evicted)
	}
	for iter := tested.Each(); iter.Next(); {
		result = append(result, tested.Items[iter.At])
	}
	assert.Equal(t, tc.contents, result, fmt.Sprint("Result: ", tc.label))
	assert.Equal(t, tc.results, remain, fmt.Sprint("Remain: ", tc.label))
}
