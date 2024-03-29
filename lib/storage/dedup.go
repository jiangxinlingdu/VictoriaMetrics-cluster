package storage

import (
	"fmt"
	"time"
)

// SetMinScrapeIntervalForDeduplication sets the minimum interval for data points during de-duplication.
//
// De-duplication is disabled if interval is 0.
//
// This function must be called before initializing the storage.
func SetMinScrapeIntervalForDeduplication(interval time.Duration) {
	minScrapeInterval = interval.Milliseconds()
}

var minScrapeInterval = int64(0)

// DeduplicateSamples removes samples from src* if they are closer to each other than minScrapeInterval.
func DeduplicateSamples(srcTimestamps []int64, srcValues []float64) ([]int64, []float64) {
	if minScrapeInterval <= 0 {
		return srcTimestamps, srcValues
	}
	if !needsDedup(srcTimestamps, minScrapeInterval) {
		// Fast path - nothing to deduplicate
		return srcTimestamps, srcValues
	}

	// Slow path - dedup data points.
	tsNext := (srcTimestamps[0] - srcTimestamps[0]%minScrapeInterval) + minScrapeInterval
	dstTimestamps := srcTimestamps[:1]
	dstValues := srcValues[:1]
	for i := 1; i < len(srcTimestamps); i++ {
		ts := srcTimestamps[i]
		if ts < tsNext {
			continue
		}
		dstTimestamps = append(dstTimestamps, ts)
		dstValues = append(dstValues, srcValues[i])

		// Update tsNext
		tsNext += minScrapeInterval
		if ts >= tsNext {
			// Slow path for updating ts.
			tsNext = (ts - ts%minScrapeInterval) + minScrapeInterval
		}
	}
	return dstTimestamps, dstValues
}

func deduplicateSamplesDuringMerge(srcTimestamps, srcValues []int64) ([]int64, []int64) {
	//if minScrapeInterval <= 0 {
	//	return srcTimestamps, srcValues
	//}
	nowTime := time.Now().UnixNano() / 1e6
	fmt.Printf("now time: %d\n", nowTime)

	if len(srcTimestamps) < 2 {
		return srcTimestamps, srcValues
	}

	minScrapeInterval = int64(15000)
	if nowTime-srcTimestamps[0] > 7200000 {
		minScrapeInterval = 15000 * 6
	} else if nowTime-srcTimestamps[0] > 5400000 {
		minScrapeInterval = 15000 * 5
	} else if nowTime-srcTimestamps[0] > 3600000 {
		minScrapeInterval = 15000 * 4
	} else if nowTime-srcTimestamps[0] > 1800000 {
		minScrapeInterval = 15000 * 3
	} else if nowTime-srcTimestamps[0] > 600000 {
		minScrapeInterval = 15000 * 2
	}
	fmt.Printf("deduplicateSamplesDuringMerge: %d\n", minScrapeInterval)

	if !needsDedup(srcTimestamps, minScrapeInterval) {
		// Fast path - nothing to deduplicate
		return srcTimestamps, srcValues
	}

	// Slow path - dedup data points.
	tsNext := (srcTimestamps[0] - srcTimestamps[0]%minScrapeInterval) + minScrapeInterval
	dstTimestamps := srcTimestamps[:1]
	dstValues := srcValues[:1]
	for i := 1; i < len(srcTimestamps); i++ {
		ts := srcTimestamps[i]
		if ts < tsNext {
			continue
		}
		dstTimestamps = append(dstTimestamps, ts)
		dstValues = append(dstValues, srcValues[i])

		// Update tsNext
		tsNext += minScrapeInterval
		if ts >= tsNext {
			// Slow path for updating ts.
			tsNext = (ts - ts%minScrapeInterval) + minScrapeInterval
		}
	}
	return dstTimestamps, dstValues
}

func needsDedup(timestamps []int64, minDelta int64) bool {
	if len(timestamps) == 0 {
		return false
	}
	prevTimestamp := timestamps[0]
	for _, ts := range timestamps[1:] {
		if ts-prevTimestamp < minDelta {
			return true
		}
		prevTimestamp = ts
	}
	return false
}
