package pprof

import (
	"compress/gzip"
	"io"
	"time"
)

// A fileMapping is code that pprof can symbolize from a file.
type fileMapping struct {
	start, end, offset uintptr
	file               string
}

// Names for the fake PCs the runtime uses for CPU time it could not attribute
// to a stack.
var fakeFunctions = map[uintptr]string{
	1: "(unattributed: threads that ran between samples)",
	2: "(cpu profiler)",
}

// Field numbers in profile.proto.
const (
	tagProfileSampleType        = 1
	tagProfileSample            = 2
	tagProfileMapping           = 3
	tagProfileLocation          = 4
	tagProfileFunction          = 5
	tagProfileStringTable       = 6
	tagProfileTimeNanos         = 9
	tagProfileDurationNanos     = 10
	tagProfilePeriodType        = 11
	tagProfilePeriod            = 12
	tagProfileDefaultSampleType = 14
)

// A profileBuilder writes a profile.proto. Add samples with sample, then call
// finish.
type profileBuilder struct {
	b          protobuf
	strings    map[string]int64
	stringList []string

	locations     map[uintptr]uint64
	locationAddrs []uintptr
}

func newProfileBuilder() *profileBuilder {
	return &profileBuilder{
		strings:    map[string]int64{"": 0},
		stringList: []string{""},
		locations:  map[uintptr]uint64{},
	}
}

func (p *profileBuilder) str(s string) int64 {
	if i, ok := p.strings[s]; ok {
		return i
	}
	i := int64(len(p.stringList))
	p.strings[s] = i
	p.stringList = append(p.stringList, s)
	return i
}

func (p *profileBuilder) valueType(field int, typ, unit string) {
	p.b.message(field, func(b *protobuf) {
		b.int64(1, p.str(typ))
		b.int64(2, p.str(unit))
	})
}

// locationIDs appends the location IDs of a stack of return addresses to ids.
// If leafIsPC is set, the leaf is an instruction address, as in CPU profiles.
func (p *profileBuilder) locationIDs(ids []uint64, pcs []uintptr, leafIsPC bool) []uint64 {
	for j, pc := range pcs {
		if j > 0 || !leafIsPC {
			pc--
		}
		id, ok := p.locations[pc]
		if !ok {
			id = uint64(len(p.locationAddrs) + 1)
			p.locations[pc] = id
			p.locationAddrs = append(p.locationAddrs, pc)
		}
		ids = append(ids, id)
	}
	return ids
}

// sample adds a sample. If label is set, the sample gets a numeric label with
// that key and value labelValue.
func (p *profileBuilder) sample(ids []uint64, values []int64, label string, labelValue int64) {
	p.b.message(tagProfileSample, func(b *protobuf) {
		b.packedUint64(1, ids)
		b.packedInt64(2, values)
		if label != "" {
			b.message(3, func(b *protobuf) {
				b.int64(1, p.str(label))
				b.int64(3, labelValue)
			})
		}
	})
}

// finish adds the locations and the remaining profile fields and writes the
// gzipped profile to w.
func (p *profileBuilder) finish(w io.Writer, start time.Time, duration time.Duration) error {
	p.writeLocations()
	p.b.int64(tagProfileTimeNanos, start.UnixNano())
	p.b.int64(tagProfileDurationNanos, int64(duration))
	for _, s := range p.stringList {
		p.b.string(tagProfileStringTable, s)
	}

	zw := gzip.NewWriter(w)
	if _, err := zw.Write(p.b.data); err != nil {
		return err
	}
	return zw.Close()
}

// writeLocations adds the locations and their mappings. pprof symbolizes code
// in files. Other code is named here if possible.
func (p *profileBuilder) writeLocations() {
	if len(p.locationAddrs) == 0 {
		return
	}
	files := fileMappings()
	for i, m := range files {
		p.b.message(tagProfileMapping, func(b *protobuf) {
			b.uint64(1, uint64(i+1))
			b.uint64(2, uint64(m.start))
			b.uint64(3, uint64(m.end))
			b.uint64(4, uint64(m.offset))
			b.int64(5, p.str(m.file))
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
	for i, addr := range p.locationAddrs {
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
		p.b.message(tagProfileLocation, func(b *protobuf) {
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
		p.b.message(tagProfileMapping, func(b *protobuf) {
			b.uint64(1, m.id)
			b.uint64(2, uint64(m.start))
			b.uint64(3, uint64(m.end))
			b.int64(5, p.str(m.lib))
			b.uint64(7, 1) // has_functions
		})
	}
	for i, f := range funcs {
		p.b.message(tagProfileFunction, func(b *protobuf) {
			b.uint64(1, uint64(i+1))
			b.int64(2, p.str(f.name))
			b.int64(3, p.str(f.name))
			b.int64(4, p.str(f.file))
		})
	}
}
