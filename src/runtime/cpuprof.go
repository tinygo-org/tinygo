//go:build darwin

package runtime

import "unsafe"

// Number of samples the CPU profiler can hold, about 5 minutes of one busy
// thread at 100Hz.
const cpuProfileSlots = 1 << 15

//export tinygo_cpuprof_start
func tinygo_cpuprof_start(hz int32, slots uint32) int32

//export tinygo_cpuprof_stop
func tinygo_cpuprof_stop()

//export tinygo_cpuprof_samples
func tinygo_cpuprof_samples(used, dropped, slotWords *uint32) unsafe.Pointer

//export tinygo_cpuprof_unsampled
func tinygo_cpuprof_unsampled(unattributed, sampler *uint64)

//go:linkname pprof_cpuProfileStart runtime/pprof.cpuProfileStart
func pprof_cpuProfileStart(hz int32) bool {
	return tinygo_cpuprof_start(hz, cpuProfileSlots) != 0
}

// pprof_cpuProfileStop stops profiling and returns the samples, each as the
// number of PCs, the CPU time in nanoseconds and the PCs, leaf first.
//
//go:linkname pprof_cpuProfileStop runtime/pprof.cpuProfileStop
func pprof_cpuProfileStop() (samples []uintptr, dropped int) {
	tinygo_cpuprof_stop()
	samples = appendUnsampled(samples)
	var used, drop, slotWords uint32
	buf := tinygo_cpuprof_samples(&used, &drop, &slotWords)
	if buf == nil {
		return nil, int(drop)
	}
	slots := unsafe.Slice((*uintptr)(buf), uintptr(used)*uintptr(slotWords))
	for i := uintptr(0); i < uintptr(used); i++ {
		slot := slots[i*uintptr(slotWords) : (i+1)*uintptr(slotWords)]
		n := slot[0]
		if n == 0 || n+2 > uintptr(slotWords) {
			drop++
			continue
		}
		samples = append(samples, slot[:n+2]...)
	}
	return samples, int(drop)
}

// Fake PCs for CPU time no sample covers. runtime/pprof gives them names.
const (
	cpuProfileUnattributedPC = 1
	cpuProfileProfilerPC     = 2
)

// appendUnsampled adds samples for CPU time used while profiling that the
// sampler did not attribute to a stack.
func appendUnsampled(samples []uintptr) []uintptr {
	var unattributed, sampler uint64
	tinygo_cpuprof_unsampled(&unattributed, &sampler)
	if unattributed > 0 {
		samples = append(samples, 1, uintptr(unattributed), cpuProfileUnattributedPC)
	}
	if sampler > 0 {
		samples = append(samples, 1, uintptr(sampler), cpuProfileProfilerPC)
	}
	return samples
}
