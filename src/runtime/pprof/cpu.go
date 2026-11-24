//go:build darwin || (linux && scheduler.threads && (amd64 || arm64))

package pprof

import (
	"errors"
	"io"
	"sync"
	"time"
)

// The same sampling rate as Go.
const cpuHz = 100

func cpuProfileStart(hz int32) bool
func cpuProfileStop() (samples []uintptr, dropped int)

var cpu struct {
	sync.Mutex
	w     io.Writer
	start time.Time
}

// StartCPUProfile enables CPU profiling for the current process.
// StopCPUProfile writes the profile to w.
func StartCPUProfile(w io.Writer) error {
	cpu.Lock()
	defer cpu.Unlock()
	if cpu.w != nil {
		return errors.New("cpu profiling already in use")
	}
	if !cpuProfileStart(cpuHz) {
		return errors.New("runtime/pprof: could not start CPU profiling")
	}
	cpu.w = w
	cpu.start = time.Now()
	return nil
}

// StopCPUProfile stops the current CPU profile, if any, and writes it out.
func StopCPUProfile() {
	cpu.Lock()
	defer cpu.Unlock()
	if cpu.w == nil {
		return
	}
	samples, _ := cpuProfileStop()
	writeCPUProfile(cpu.w, samples, cpu.start, time.Since(cpu.start))
	cpu.w = nil
}

func writeCPUProfile(w io.Writer, samples []uintptr, start time.Time, duration time.Duration) error {
	p := newProfileBuilder()
	p.valueType(tagProfileSampleType, "samples", "count")
	p.valueType(tagProfileSampleType, "cpu", "nanoseconds")

	// Add up the samples with the same stack, in the order they first appear.
	type stackKey string
	counts := map[stackKey]int64{}
	cpuNs := map[stackKey]int64{}
	var order []stackKey
	var ids []uint64
	for i := 0; i < len(samples); {
		n := int(samples[i])
		weight := int64(samples[i+1])
		pcs := samples[i+2 : i+2+n]
		i += 2 + n
		ids = p.locationIDs(ids[:0], pcs, true)
		key := make([]byte, 0, n*8)
		for _, id := range ids {
			for k := 0; k < 8; k++ {
				key = append(key, byte(id>>(8*k)))
			}
		}
		k := stackKey(key)
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
		cpuNs[k] += weight
	}
	for _, k := range order {
		ids = ids[:0]
		for j := 0; j < len(k); j += 8 {
			var id uint64
			for x := 0; x < 8; x++ {
				id |= uint64(k[j+x]) << (8 * x)
			}
			ids = append(ids, id)
		}
		p.sample(ids, []int64{counts[k], cpuNs[k]}, "", 0)
	}

	p.valueType(tagProfilePeriodType, "cpu", "nanoseconds")
	p.b.int64(tagProfilePeriod, int64(time.Second/cpuHz))
	return p.finish(w, start, duration)
}
