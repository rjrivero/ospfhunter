package main

import "github.com/rjrivero/ring"

type bucket struct {
	second int64
	hits   int
}

// SlidingCount es una ventana deslizante de paquetes.
// Cuenta el número de paquetes recibidos durante los últimos N segundos
// que cumplen una cierta característica, y mantiene una lista circular
// con esos últimos paquetes.
type slidingCount struct {
	ring.Ring
	buckets  []bucket // counting buckets
	interval int
	tail     int // Memoize tail to be able to increment the value
	accum    int // sum of items in the time window
}

func makeSlidingCount(interval, burst int) slidingCount {
	// We are only interested in bursts up to 'burst' packets long
	// There is no point in keeping more buckets than 'burst' or
	// 'interval', whichever is lower.
	bucketSize := interval
	if burst < interval {
		bucketSize = burst
	}
	sc := slidingCount{
		Ring:     ring.New(bucketSize),
		buckets:  make([]bucket, bucketSize),
		interval: interval,
	}
	// Initialize first bucket
	sc.tail = sc.Push()
	sc.buckets[sc.tail] = bucket{second: 0, hits: 0}
	return sc
}

// Increment the sliding window count at a given second.
// Also store the packet number and packet content in the circular buffer.
// This function must be called with monotonically increasing "atSecond" number.
func (s *slidingCount) Inc(atSecond int64) int {
	lastSecond := s.buckets[s.tail].second
	switch {
	case atSecond < lastSecond:
		panic("Time cannot go backwards!")
	case atSecond == lastSecond:
		// Accumulate in the current second
		s.buckets[s.tail].hits++
		s.accum++
		return s.accum
	}
	// Before adding a new entry, pop old ones
	deadline := atSecond - int64(s.interval)
	for iter := s.Ring; iter.Some(); {
		head := s.buckets[iter.PopFront()]
		// If we reached the deadline, stop
		if head.second > deadline {
			break
		}
		// Otherwise, decrement accumulator and pop oldest item
		s.accum -= head.hits
		s.PopFront()
	}
	// If still full, pop oldest item. We are only interested in bursts up to
	// 'burst' size, anyway.
	if s.Full() {
		head := s.buckets[s.PopFront()]
		s.accum -= head.hits
	}
	// Add a new bucket for current second
	s.tail = s.Push()
	s.buckets[s.tail] = bucket{second: atSecond, hits: 1}
	s.accum++
	return s.accum
}
