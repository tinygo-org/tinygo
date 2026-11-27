//go:build tinygo.wasm && tinygo.shadowstack

package runtime

import "unsafe"

// The call site tables of the shadow stack, defined by the compiler. Site id
// has the frames from shadowSites[id-1] up to shadowSites[id].

//go:extern tinygo_shadow_nsites
var shadowNumSites uintptr

//go:extern tinygo_shadow_sites
var shadowSites [0]uint32

//go:extern tinygo_shadow_frames
var shadowFrames [0]shadowFrame

type shadowFrame struct {
	name, file string
	line       uintptr
}

// pprof_siteFrame returns frame i of a call site, innermost first.
//
//go:linkname pprof_siteFrame runtime/pprof.siteFrame
func pprof_siteFrame(site uintptr, i int) (name, file string, line int, ok bool) {
	if site == 0 || site > shadowNumSites || i < 0 {
		return
	}
	sites := unsafe.Slice((*uint32)(unsafe.Pointer(&shadowSites)), shadowNumSites+1)
	start, end := uintptr(sites[site-1]), uintptr(sites[site])
	if uintptr(i) >= end-start {
		return
	}
	f := (*shadowFrame)(unsafe.Add(unsafe.Pointer(&shadowFrames), (start+uintptr(i))*unsafe.Sizeof(shadowFrame{})))
	return f.name, f.file, int(f.line), true
}
