//go:build darwin || (linux && scheduler.threads && (amd64 || arm64))

package pprof

import (
	"compress/gzip"
	"errors"
	"io"
	"sync"
	"time"
)

// The same sampling rate as Go.
const cpuHz = 100

// Names for the fake PCs the runtime uses for CPU time it could not attribute
// to a stack.
var fakeFunctions = map[uintptr]string{
	1: "(unattributed: threads that ran between samples)",
	2: "(cpu profiler)",
}

func cpuProfileStart(hz int32) bool
func cpuProfileStop() (samples []uintptr, dropped int)

// A fileMapping is code that pprof can symbolize from a file.
type fileMapping struct {
	start, end, offset uintptr
	file               string
}

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
	files := fileMappings()
	period := int64(time.Second / cpuHz)

	var b protobuf
	strings := map[string]int64{"": 0}
	stringList := []string{""}
	str := func(s string) int64 {
		if i, ok := strings[s]; ok {
			return i
		}
		i := int64(len(stringList))
		strings[s] = i
		stringList = append(stringList, s)
		return i
	}
	valueType := func(field int, typ, unit string) {
		b.message(field, func(b *protobuf) {
			b.int64(1, str(typ))
			b.int64(2, str(unit))
		})
	}
	valueType(1, "samples", "count")
	valueType(1, "cpu", "nanoseconds")

	// All frames but the leaf hold return addresses, so use the call
	// instruction before them.
	type stackKey string
	counts := map[stackKey]int64{}
	cpuNs := map[stackKey]int64{}
	var order []stackKey
	locations := map[uintptr]uint64{}
	var locationAddrs []uintptr
	var ids []uint64
	for i := 0; i < len(samples); {
		n := int(samples[i])
		weight := int64(samples[i+1])
		pcs := samples[i+2 : i+2+n]
		i += 2 + n
		key := make([]byte, 0, n*8)
		for j, pc := range pcs {
			if j > 0 {
				pc--
			}
			id, ok := locations[pc]
			if !ok {
				id = uint64(len(locationAddrs) + 1)
				locations[pc] = id
				locationAddrs = append(locationAddrs, pc)
			}
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
		count, ns := counts[k], cpuNs[k]
		b.message(2, func(b *protobuf) {
			b.packedUint64(1, ids)
			b.packedInt64(2, []int64{count, ns})
		})
	}

	// pprof symbolizes code in files. Other code is named here if possible.
	for i, m := range files {
		b.message(3, func(b *protobuf) {
			b.uint64(1, uint64(i+1))
			b.uint64(2, uint64(m.start))
			b.uint64(3, uint64(m.end))
			b.uint64(4, uint64(m.offset))
			b.int64(5, str(m.file))
		})
	}
	type sysMapping struct {
		id         uint64
		lib        string
		start, end uintptr
	}
	type function struct {
		name, file string
	}
	var (
		sysMappings = map[string]*sysMapping{}
		sysOrder    []*sysMapping
		funcIDs     = map[function]uint64{}
		funcs       []function
	)
	functionID := func(name, file string) uint64 {
		f := function{name, file}
		id := funcIDs[f]
		if id == 0 {
			funcs = append(funcs, f)
			id = uint64(len(funcs))
			funcIDs[f] = id
		}
		return id
	}
	for i, addr := range locationAddrs {
		mappingID, funcID := uint64(0), uint64(0)
		for j, m := range files {
			if addr >= m.start && addr < m.end {
				mappingID = uint64(j + 1)
				break
			}
		}
		if name, ok := fakeFunctions[addr]; ok {
			mappingID, funcID = 0, functionID(name, "")
		} else if mappingID == 0 {
			lib, base, sym, ok := lookupSystem(addr)
			if !ok {
				lib, base = "[unknown]", addr
			}
			m := sysMappings[lib]
			if m == nil {
				m = &sysMapping{id: uint64(len(files) + 1 + len(sysOrder)), lib: lib, start: base, end: addr + 1}
				sysMappings[lib] = m
				sysOrder = append(sysOrder, m)
			}
			m.start = min(m.start, base)
			m.end = max(m.end, addr+1)
			mappingID = m.id
			if sym != "" {
				funcID = functionID(sym, lib)
			}
		}
		b.message(4, func(b *protobuf) {
			b.uint64(1, uint64(i+1))
			b.uint64(2, mappingID)
			b.uint64(3, uint64(addr))
			if funcID != 0 {
				b.message(4, func(b *protobuf) {
					b.uint64(1, funcID)
				})
			}
		})
	}
	for _, m := range sysOrder {
		b.message(3, func(b *protobuf) {
			b.uint64(1, m.id)
			b.uint64(2, uint64(m.start))
			b.uint64(3, uint64(m.end))
			b.int64(5, str(m.lib))
			b.uint64(7, 1) // has_functions
		})
	}
	for i, f := range funcs {
		b.message(5, func(b *protobuf) {
			b.uint64(1, uint64(i+1))
			b.int64(2, str(f.name))
			b.int64(3, str(f.name))
			b.int64(4, str(f.file))
		})
	}
	b.int64(9, start.UnixNano())
	b.int64(10, int64(duration))
	valueType(11, "cpu", "nanoseconds")
	b.int64(12, period)
	for _, s := range stringList {
		b.string(6, s)
	}

	zw := gzip.NewWriter(w)
	if _, err := zw.Write(b.data); err != nil {
		return err
	}
	return zw.Close()
}
