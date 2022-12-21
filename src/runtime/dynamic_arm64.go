package runtime

import (
	"unsafe"
)

const debugLoader = false

const (
	rAARCH64_RELATIVE = 1027
	dtNULL            = 0 /* Terminating entry. */
	dtRELA            = 7 /* Address of ElfNN_Rela relocations. */
	dtRELASZ          = 8 /* Total size of ElfNN_Rela relocations. */
)

/* ELF64 relocations that need an addend field. */
type rela64 struct {
	Off    uint64 /* Location to be relocated. */
	Info   uint64 /* Relocation type and symbol index. */
	Addend int64  /* Addend. */
}

// ELF64 Dynamic structure. The ".dynamic" section contains an array of them.
type dyn64 struct {
	Tag int64  /* Entry type. */
	Val uint64 /* Integer/address value */
}

//export __dynamic_loader
func dynamicLoader(base uintptr, dyn *dyn64) {
	var rela *rela64
	relasz := uint64(0)

	if debugLoader {
		println("ASLR Base: ", base)
	}

	for dyn.Tag != dtNULL {
		switch dyn.Tag {
		case dtRELA:
			rela = (*rela64)(unsafe.Pointer(base + uintptr(dyn.Val)))
		case dtRELASZ:
			relasz = uint64(dyn.Val) / uint64(unsafe.Sizeof(rela64{}))
		}

		ptr := unsafe.Pointer(dyn)
		ptr = unsafe.Add(ptr, unsafe.Sizeof(dyn64{}))
		dyn = (*dyn64)(ptr)
	}

	if rela == nil {
		runtimePanic("bad reloc")
	}

	if debugLoader {
		println("Sections to load: ", relasz)
	}

	for relasz > 0 && rela != nil {
		switch rela.Info {
		case rAARCH64_RELATIVE:
			if debugLoader {
				println("relocating ", uintptr(rela.Addend), " to ", base+uintptr(rela.Addend))
			}
			ptr := (*uint64)(unsafe.Pointer(base + uintptr(rela.Off)))
			*ptr = uint64(base + uintptr(rela.Addend))
		default:
			if debugLoader {
				println("unknown section to load:", rela.Info)
			}
		}

		rptr := unsafe.Pointer(rela)
		rptr = unsafe.Add(rptr, unsafe.Sizeof(rela64{}))
		rela = (*rela64)(rptr)
		relasz--
	}
}
