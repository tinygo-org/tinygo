package pprof

import (
	"errors"
	"io"
	"math"
	"runtime"
	"time"
)

// WriteHeapProfile is shorthand for Lookup("heap").WriteTo(w, 0).
func WriteHeapProfile(w io.Writer) error {
	return writeHeap(w, 0, "")
}

func writeHeap(w io.Writer, debug int, defaultSampleType string) error {
	if debug != 0 {
		return errors.New("runtime/pprof: only debug=0 is supported for heap profiles")
	}
	// The number of records can grow between the calls, so leave some room.
	n, _ := runtime.MemProfile(nil, true)
	var p []runtime.MemProfileRecord
	for {
		p = make([]runtime.MemProfileRecord, n+50)
		var ok bool
		n, ok = runtime.MemProfile(p, true)
		if ok {
			p = p[:n]
			break
		}
	}
	return writeHeapProto(w, p, int64(runtime.MemProfileRate), defaultSampleType)
}

// writeHeapProto writes the records as a heap profile, like upstream.
func writeHeapProto(w io.Writer, records []runtime.MemProfileRecord, rate int64, defaultSampleType string) error {
	p := newProfileBuilder()
	p.valueType(tagProfilePeriodType, "space", "bytes")
	p.b.int64(tagProfilePeriod, rate)
	p.valueType(tagProfileSampleType, "alloc_objects", "count")
	p.valueType(tagProfileSampleType, "alloc_space", "bytes")
	p.valueType(tagProfileSampleType, "inuse_objects", "count")
	p.valueType(tagProfileSampleType, "inuse_space", "bytes")
	if defaultSampleType != "" {
		p.b.int64(tagProfileDefaultSampleType, p.str(defaultSampleType))
	}

	values := make([]int64, 4)
	var ids []uint64
	for i := range records {
		r := &records[i]
		ids = p.locationIDs(ids[:0], r.Stack(), false)
		values[0], values[1] = scaleHeapSample(r.AllocObjects, r.AllocBytes, rate)
		values[2], values[3] = scaleHeapSample(r.InUseObjects(), r.InUseBytes(), rate)
		var blockSize int64
		if r.AllocObjects > 0 {
			blockSize = r.AllocBytes / r.AllocObjects
		}
		label := ""
		if blockSize != 0 {
			label = "bytes"
		}
		p.sample(ids, values, label, blockSize)
	}
	return p.finish(w, time.Now(), 0)
}

// scaleHeapSample estimates the counts of all allocations from the samples.
// An allocation of size bytes is sampled with probability 1-exp(-size/rate).
func scaleHeapSample(count, size, rate int64) (int64, int64) {
	if count == 0 || size == 0 {
		return 0, 0
	}
	if rate <= 1 {
		return count, size
	}
	avgSize := float64(size) / float64(count)
	scale := 1 / (1 - math.Exp(-avgSize/float64(rate)))
	return int64(float64(count) * scale), int64(float64(size) * scale)
}
