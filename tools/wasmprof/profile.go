package main

import (
	"io"
	"math"
	"slices"

	"github.com/google/pprof/profile"
)

type pprofProfile struct{ *profile.Profile }

func (p pprofProfile) write(w io.Writer) error { return p.Profile.Write(w) }

// Stats counts samples by how their leaf position was resolved, to show how
// much of a profile rests on heuristics.
type Stats map[string]int64

func buildProfile(gp *guestProfile, am *addrMap, wf *wasmFile, wasmPath string, stats Stats) pprofProfile {
	intervalNs := int64(math.Round(gp.Meta.Interval * 1e6))
	p := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
		PeriodType: &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:     intervalNs,
		TimeNanos:  int64(gp.Meta.StartTime * 1e6),
	}
	hasDWARF := slices.ContainsFunc(wf.modules, func(m *wasmModule) bool { return m.dw != nil })
	mapping := &profile.Mapping{
		ID:              1,
		Start:           0,
		Limit:           math.MaxUint32,
		File:            wasmPath,
		HasFunctions:    true,
		HasFilenames:    hasDWARF,
		HasLineNumbers:  hasDWARF,
		HasInlineFrames: hasDWARF,
	}
	p.Mapping = []*profile.Mapping{mapping}

	type funcKey struct {
		name, file string
		line       int64
	}
	funcs := map[funcKey]*profile.Function{}
	function := func(name, file string, startLine int64) *profile.Function {
		k := funcKey{name, file, startLine}
		if f := funcs[k]; f != nil {
			return f
		}
		f := &profile.Function{ID: uint64(len(p.Function) + 1), Name: name, SystemName: name, Filename: file, StartLine: startLine}
		funcs[k] = f
		p.Function = append(p.Function, f)
		return f
	}

	// The same native address always maps to the same wasm position, so key
	// locations by it.
	locs := map[int64]*profile.Location{}
	leafKind := map[*profile.Location]string{}
	guessedLoc := map[*profile.Location]bool{}
	location := func(t *guestThread, frameIdx int) *profile.Location {
		addr := t.FrameTable.Address[frameIdx]
		profName := t.str(t.FuncTable.Name[t.FrameTable.Func[frameIdx]])
		if l := locs[addr]; l != nil && addr >= 0 {
			return l
		}
		l := &profile.Location{ID: uint64(len(p.Location) + 1), Mapping: mapping}
		if addr >= 0 {
			l.Address = uint64(addr)
		}
		var frames []frame
		kind := "no wasm position (prologue or outside addrmap)"
		wasmOff := int64(-1)
		if addr >= 0 {
			wasmOff = am.lookup(addr)
		}
		var mod *wasmModule
		if wasmOff >= 0 {
			mod = wf.moduleFor(wasmOff)
		}
		if mod != nil && mod.dw != nil {
			dw := mod.dw
			kind = "direct line row"
			pc := uint64(mod.sampleOffset(wasmOff) - mod.codeStart)
			if mod.sampleOffset(wasmOff) != wasmOff {
				kind = "loop header moved into body"
				pc = dw.loopBodyPC(pc)
			}
			frames = dw.frames(pc)
			switch {
			case len(frames) == 0:
				kind = "no DWARF function or line"
			case frames[0].guessed:
				kind = "function guessed (inline scope missing)"
			case frames[0].prologue:
				kind = "no line row, used declaration line"
			case frames[0].function == "":
				kind = "line row without DWARF function (asm)"
			}
			for _, f := range frames {
				if f.guessed {
					guessedLoc[l] = true
				}
			}
		} else if mod != nil {
			kind = "no DWARF in module"
		}
		leafKind[l] = kind
		if len(frames) == 0 {
			l.Line = []profile.Line{{Function: function(profName, "", 0)}}
		}
		for i, f := range frames {
			name := f.function
			if i == len(frames)-1 && profName != "" {
				// The outermost function is the wasm function, so use its name
				// from the name section, as the profile does.
				name = profName
			}
			if name == "" {
				name = "?"
			}
			file := f.file
			if file == "" {
				file = f.srcFile
			}
			l.Line = append(l.Line, profile.Line{Function: function(name, file, f.declLine), Line: f.line})
		}
		p.Location = append(p.Location, l)
		if addr >= 0 {
			locs[addr] = l
		}
		return l
	}

	var durationMs float64
	for ti := range gp.Threads {
		t := &gp.Threads[ti]
		stackLocs := make([][]*profile.Location, len(t.StackTable.Frame))
		var stack func(i int) []*profile.Location
		stack = func(i int) []*profile.Location {
			if stackLocs[i] != nil {
				return stackLocs[i]
			}
			// pprof lists the leaf first.
			ls := []*profile.Location{location(t, t.StackTable.Frame[i])}
			if pre := t.StackTable.Prefix[i]; pre != nil {
				ls = append(ls, stack(*pre)...)
			}
			stackLocs[i] = ls
			return ls
		}
		for si, st := range t.Samples.Stack {
			if si < len(t.Samples.TimeDeltas) {
				durationMs += t.Samples.TimeDeltas[si]
			}
			if st == nil {
				continue
			}
			w := int64(1)
			if si < len(t.Samples.Weight) {
				w = int64(t.Samples.Weight[si])
			}
			ls := stack(*st)
			if stats != nil {
				stats[leafKind[ls[0]]] += w
				stats["total"] += w
				for _, l := range ls {
					if guessedLoc[l] {
						stats["any frame guessed"] += w
						break
					}
				}
			}
			p.Sample = append(p.Sample, &profile.Sample{
				Location: ls,
				Value:    []int64{w, w * intervalNs},
			})
		}
	}
	p.DurationNanos = int64(durationMs * 1e6)
	p = p.Compact()
	return pprofProfile{p}
}
