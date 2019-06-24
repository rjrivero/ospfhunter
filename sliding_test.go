package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type slidingTestCase struct {
	label string
	size  int
	burst int
	time  []int
	memo  []int
}

func TestSliding(t *testing.T) {

	testCases := []slidingTestCase{
		{
			label: "1-item Empty sequence should count 1",
			size:  5,
			burst: 5,
			time:  []int{3},
			memo:  []int{1},
		},
		{
			label: "Repeated values should accumulate",
			size:  5,
			burst: 5,
			time:  []int{4, 4, 4, 5, 5, 5, 6, 6, 6},
			memo:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			label: "Items should fall off after the window passes",
			size:  5,
			burst: 5,
			time:  []int{1, 2, 3, 4, 5, 6, 7, 8},
			memo:  []int{1, 2, 3, 4, 5, 5, 5, 5},
		},
		{
			label: "Accumulated items should fall off",
			size:  5,
			burst: 5,
			time:  []int{1, 1, 2, 3, 4, 5, 6, 7, 8},
			memo:  []int{1, 2, 3, 4, 5, 6, 5, 5, 5},
		},
		{
			label: "Accumulated should fall off in gaps",
			size:  5,
			burst: 5,
			time:  []int{1, 1, 2, 3, 4, 5, 5, 5, 8, 11},
			memo:  []int{1, 2, 3, 4, 5, 6, 7, 8, 5, 2},
		},
		{
			label: "Large gaps should reset count",
			size:  5,
			burst: 5,
			time:  []int{4, 5, 20, 22},
			memo:  []int{1, 2, 1, 2},
		},
		{
			label: "Small burst - 1-item Empty sequence should count 1",
			size:  5,
			burst: 2,
			time:  []int{3},
			memo:  []int{1},
		},
		{
			label: "Small burst - Repeated values should accumulate",
			size:  5,
			burst: 2,
			time:  []int{4, 4, 4, 5, 5, 5, 6, 6, 6},
			memo:  []int{1, 2, 3, 4, 5, 6, 4, 5, 6},
		},
		{
			label: "Small burst - Items should fall off after the window passes",
			size:  5,
			burst: 2,
			time:  []int{1, 2, 3, 4, 5, 6, 7, 8},
			memo:  []int{1, 2, 2, 2, 2, 2, 2, 2},
		},
		{
			label: "Small burst - Accumulated items should fall off",
			size:  5,
			burst: 2,
			time:  []int{1, 1, 2, 3, 4, 5, 6, 7, 8},
			memo:  []int{1, 2, 3, 2, 2, 2, 2, 2, 2},
		},
		{
			label: "Small burst - Accumulated should fall off in gaps",
			size:  5,
			burst: 2,
			time:  []int{1, 1, 2, 3, 4, 5, 5, 5, 8, 11},
			memo:  []int{1, 2, 3, 2, 2, 2, 3, 4, 4, 2},
		},
		{
			label: "Small burst - Large gaps should reset count",
			size:  5,
			burst: 5,
			time:  []int{4, 5, 20, 22},
			memo:  []int{1, 2, 1, 2},
		},
		{
			label: "Small interval - 1-item Empty sequence should count 1",
			size:  2,
			burst: 5,
			time:  []int{3},
			memo:  []int{1},
		},
		{
			label: "Small interval - Repeated values should accumulate",
			size:  2,
			burst: 5,
			time:  []int{4, 4, 4, 5, 5, 5, 6, 6, 6},
			memo:  []int{1, 2, 3, 4, 5, 6, 4, 5, 6},
		},
		{
			label: "Small interval - Items should fall off after the window passes",
			size:  2,
			burst: 5,
			time:  []int{1, 2, 3, 4, 5, 6, 7, 8},
			memo:  []int{1, 2, 2, 2, 2, 2, 2, 2},
		},
		{
			label: "Small interval - Accumulated items should fall off",
			size:  2,
			burst: 5,
			time:  []int{1, 1, 2, 3, 4, 5, 6, 7, 8},
			memo:  []int{1, 2, 3, 2, 2, 2, 2, 2, 2},
		},
		{
			label: "Small interval - Accumulated should fall off in gaps",
			size:  2,
			burst: 5,
			time:  []int{1, 1, 2, 3, 4, 5, 5, 5, 8, 9},
			memo:  []int{1, 2, 3, 2, 2, 2, 3, 4, 1, 2},
		},
		{
			label: "Small interval - Large gaps should reset count",
			size:  2,
			burst: 5,
			time:  []int{4, 5, 20, 22},
			memo:  []int{1, 2, 1, 1},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, tc.run)
	}
}

func (tc *slidingTestCase) run(t *testing.T) {
	result := make([]int, 0, len(tc.memo))
	tested := makeSlidingCount(tc.size, tc.burst)
	for _, atSecond := range tc.time {
		accum := tested.Inc(int64(atSecond))
		result = append(result, accum)
	}
	assert.Equal(t, tc.memo, result, tc.label)
}
