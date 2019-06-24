package main

// Tengo que usar este fork hasta que arreglen los problemas con el parseo OSPF
// Ver:
// https://github.com/google/gopacket/pull/671
// https://github.com/google/gopacket/pull/672

type bucket struct {
	second int64
	hits   int
}

// SlidingCount es una ventana deslizante de paquetes
// Cuenta el número de paquetes recibidos durante los últimos N segundos
// que cumplen una cierta característica, y mantiene una lista circular
// con esos últimos paquetes.
type slidingCount struct {
	Ring
	buckets  []bucket // counting buckets
	interval int
	head     int // Memoize head to be able to increment the value
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
		Ring:     Ring{Size: bucketSize},
		buckets:  make([]bucket, bucketSize),
		interval: interval,
	}
	// Initialize first bucket
	sc.head = sc.HeadNext()
	sc.buckets[sc.head] = bucket{second: 0, hits: 0}
	return sc
}

// Increment the sliding window count at a given second.
// Also store the packet number and packet content in the circular buffer.
// This function must be called with monotonically increasing "atSecond" number.
func (s *slidingCount) Inc(atSecond int64) int {
	lastSecond := s.buckets[s.head].second
	switch {
	case atSecond < lastSecond:
		panic("Time cannot go backwards!")
	case atSecond == lastSecond:
		// Accumulate in the current second
		s.buckets[s.head].hits++
		s.accum++
		return s.accum
	}
	// Before adding a new entry, pop old ones
	deadline := atSecond - int64(s.interval)
	for iter := s.Each(); iter.Next(); {
		tail := s.buckets[iter.At]
		// If we reached the deadline, stop
		if tail.second > deadline {
			break
		}
		// Otherwise, decrement accumulator and pop oldest item
		s.accum -= tail.hits
		s.TailNext()
	}
	// If still full, pop oldest item. We are only interested in bursts up to
	// 'burst' size, anyway.
	if s.Ring.Full() {
		tail := s.buckets[s.TailNext()]
		s.accum -= tail.hits
	}
	// Add a new bucket for current second
	s.head = s.HeadNext()
	s.buckets[s.head] = bucket{second: atSecond, hits: 1}
	s.accum++
	return s.accum
}
