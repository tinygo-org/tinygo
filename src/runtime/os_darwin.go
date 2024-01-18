//go:build darwin

package runtime

import "unsafe"

import "C" // dummy import so that os_darwin.c works

const GOOS = "darwin"

const (
	// See https://github.com/golang/go/blob/master/src/syscall/zerrors_darwin_amd64.go
	flag_PROT_READ     = 0x1
	flag_PROT_WRITE    = 0x2
	flag_MAP_PRIVATE   = 0x2
	flag_MAP_ANONYMOUS = 0x1000 // MAP_ANON
)

// Source: https://opensource.apple.com/source/Libc/Libc-1439.100.3/include/time.h.auto.html
const (
	clock_REALTIME      = 0
	clock_MONOTONIC_RAW = 4
)

// https://opensource.apple.com/source/xnu/xnu-7195.141.2/EXTERNAL_HEADERS/mach-o/loader.h.auto.html
type machHeader struct {
	magic      uint32
	cputype    uint32
	cpusubtype uint32
	filetype   uint32
	ncmds      uint32
	sizeofcmds uint32
	flags      uint32
	reserved   uint32
}

// Struct for the LC_SEGMENT_64 load command.
type segmentLoadCommand struct {
	cmd      uint32 // LC_SEGMENT_64
	cmdsize  uint32
	segname  [16]byte
	vmaddr   uintptr
	vmsize   uintptr
	fileoff  uintptr
	filesize uintptr
	maxprot  uint32
	initprot uint32
	nsects   uint32
	flags    uint32
}

// MachO header of the currently running process.
//
//go:extern _mh_execute_header
var libc_mh_execute_header machHeader

// Find global variables in .data/.bss sections.
// The MachO linker doesn't seem to provide symbols for the start and end of the
// data section. There is get_etext, get_edata, and get_end, but these are
// undocumented and don't work with ASLR (which is enabled by default).
// Therefore, read the MachO header directly.
func findGlobals(found func(start, end uintptr)) {
	// Here is a useful blog post to understand the MachO file format:
	// https://h3adsh0tzz.com/2020/01/macho-file-format/

	const (
		MH_MAGIC_64   = 0xfeedfacf
		LC_SEGMENT_64 = 0x19
		VM_PROT_WRITE = 0x02
	)

	// Sanity check that we're actually looking at a MachO header.
	if gcAsserts && libc_mh_execute_header.magic != MH_MAGIC_64 {
		runtimePanic("gc: unexpected MachO header")
	}

	// Iterate through the load commands.
	// Because we're only interested in LC_SEGMENT_64 load commands, cast the
	// pointer to that struct in advance.
	var offset uintptr
	var hasOffset bool
	cmd := (*segmentLoadCommand)(unsafe.Pointer(uintptr(unsafe.Pointer(&libc_mh_execute_header)) + unsafe.Sizeof(machHeader{})))
	for i := libc_mh_execute_header.ncmds; i != 0; i-- {
		if cmd.cmd == LC_SEGMENT_64 {
			if cmd.fileoff == 0 && cmd.nsects != 0 {
				// Detect ASLR offset by checking fileoff and nsects. This
				// locates the __TEXT segment. This matches getsectiondata:
				// https://opensource.apple.com/source/cctools/cctools-973.0.1/libmacho/getsecbyname.c.auto.html
				offset = uintptr(unsafe.Pointer(&libc_mh_execute_header)) - cmd.vmaddr
				hasOffset = true
			}
			if cmd.maxprot&VM_PROT_WRITE != 0 {
				// Found a writable segment, which may contain Go globals.
				if gcAsserts && !hasOffset {
					// No ASLR offset detected. Did the __TEXT segment come
					// after the __DATA segment?
					// Note that when ASLR is disabled (for example, when
					// running inside lldb), the offset is zero. That's why we
					// need a separate hasOffset for this assert.
					runtimePanic("gc: did not detect ASLR offset")
				}
				// Scan this segment for GC roots.
				// This could be improved by only reading the memory areas
				// covered by sections. That would reduce the amount of memory
				// scanned a little bit (up to a single VM page).
				found(offset+cmd.vmaddr, offset+cmd.vmaddr+cmd.vmsize)
			}
		}

		// Move on to the next load command (wich may or may not be a
		// LC_SEGMENT_64).
		cmd = (*segmentLoadCommand)(unsafe.Add(unsafe.Pointer(cmd), cmd.cmdsize))
	}
}

func hardwareRand() (n uint64, ok bool) {
	n |= uint64(libc_arc4random())
	n |= uint64(libc_arc4random()) << 32
	return n, true
}

// uint32_t arc4random(void);
//
//export arc4random
func libc_arc4random() uint32
