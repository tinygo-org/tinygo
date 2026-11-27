//go:build darwin

package runtime

import "unsafe"

//export tinygo_cpuprof_text
func tinygo_cpuprof_text(start, end *uintptr)

//export tinygo_cpuprof_executable
func tinygo_cpuprof_executable(buf *byte, size uint32) uint32

//export tinygo_cpuprof_dladdr
func tinygo_cpuprof_dladdr(pc uintptr, lib **byte, base *uintptr, sym **byte) int32

// pprof_executable returns the path of the executable and the address range
// of its text segment.
//
//go:linkname pprof_executable runtime/pprof.executable
func pprof_executable() (path string, start, end uintptr) {
	tinygo_cpuprof_text(&start, &end)
	buf := make([]byte, 1024)
	n := tinygo_cpuprof_executable(&buf[0], uint32(len(buf)))
	return string(buf[:n]), start, end
}

// pprof_lookupSystem returns the library and symbol containing pc, for code
// outside the executable.
//
//go:linkname pprof_lookupSystem runtime/pprof.lookupSystem
func pprof_lookupSystem(pc uintptr) (lib string, base uintptr, sym string, ok bool) {
	var libPtr, symPtr *byte
	if tinygo_cpuprof_dladdr(pc, &libPtr, &base, &symPtr) == 0 {
		return "", 0, "", false
	}
	return gostringFromC(libPtr), base, gostringFromC(symPtr), true
}

// gostringFromC copies a NUL-terminated C string.
func gostringFromC(p *byte) string {
	if p == nil {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(p, n))
}
