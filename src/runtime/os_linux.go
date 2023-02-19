//go:build linux && !baremetal && !nintendoswitch && !wasi

package runtime

// This file is for systems that are _actually_ Linux (not systems that pretend
// to be Linux, like baremetal systems).

import "unsafe"

const GOOS = "linux"

const (
	// See https://github.com/torvalds/linux/blob/master/include/uapi/asm-generic/mman-common.h
	flag_PROT_READ     = 0x1
	flag_PROT_WRITE    = 0x2
	flag_MAP_PRIVATE   = 0x2
	flag_MAP_ANONYMOUS = flag_linux_MAP_ANONYMOUS
)

// Source: https://github.com/torvalds/linux/blob/master/include/uapi/linux/time.h
const (
	clock_REALTIME      = 0
	clock_MONOTONIC_RAW = 4
)

// For the definition of the various header structs, see:
// https://refspecs.linuxfoundation.org/elf/elf.pdf
// Also useful:
// https://en.wikipedia.org/wiki/Executable_and_Linkable_Format
type elfHeader struct {
	ident_magic      uint32
	ident_class      uint8
	ident_data       uint8
	ident_version    uint8
	ident_osabi      uint8
	ident_abiversion uint8
	_                [7]byte // reserved
	filetype         uint16
	machine          uint16
	version          uint32
	entry            uintptr
	phoff            uintptr
	shoff            uintptr
	flags            uint32
	ehsize           uint16
	phentsize        uint16
	phnum            uint16
	shentsize        uint16
	shnum            uint16
	shstrndx         uint16
}

type elfProgramHeader64 struct {
	_type  uint32
	flags  uint32
	offset uintptr
	vaddr  uintptr
	paddr  uintptr
	filesz uintptr
	memsz  uintptr
	align  uintptr
}

type elfProgramHeader32 struct {
	_type  uint32
	offset uintptr
	vaddr  uintptr
	paddr  uintptr
	filesz uintptr
	memsz  uintptr
	flags  uint32
	align  uintptr
}

// ELF header of the currently running process.
//
//go:extern __ehdr_start
var ehdr_start elfHeader

// markGlobals marks all globals, which are reachable by definition.
// It parses the ELF program header to find writable segments.
func markGlobals() {
	// Relevant constants from the ELF specification.
	// See: https://refspecs.linuxfoundation.org/elf/elf.pdf
	const (
		PT_LOAD = 1
		PF_W    = 0x2 // program flag: write access
	)

	headerPtr := unsafe.Pointer(uintptr(unsafe.Pointer(&ehdr_start)) + ehdr_start.phoff)
	for i := 0; i < int(ehdr_start.phnum); i++ {
		// Look for a writable segment and scan its contents.
		// There is a little bit of duplication here, which is unfortunate. But
		// the alternative would be to put elfProgramHeader in separate files
		// which is IMHO a lot uglier. If only the ELF spec was consistent
		// between 32-bit and 64-bit...
		if TargetBits == 64 {
			header := (*elfProgramHeader64)(headerPtr)
			if header._type == PT_LOAD && header.flags&PF_W != 0 {
				start := header.vaddr
				end := start + header.memsz
				markRoots(start, end)
			}
		} else {
			header := (*elfProgramHeader32)(headerPtr)
			if header._type == PT_LOAD && header.flags&PF_W != 0 {
				start := header.vaddr
				end := start + header.memsz
				markRoots(start, end)
			}
		}
		headerPtr = unsafe.Pointer(uintptr(headerPtr) + uintptr(ehdr_start.phentsize))
	}
}

//export getpagesize
func libc_getpagesize() int

//go:linkname syscall_Getpagesize syscall.Getpagesize
func syscall_Getpagesize() int {
	return libc_getpagesize()
}
