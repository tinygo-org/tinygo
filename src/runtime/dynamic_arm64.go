package runtime

import (
	"unsafe"
)

const debugLoader = false

const (
	R_AARCH64_RELATIVE = 1027
	DT_NULL            = 0 /* Terminating entry. */
	DT_RELA            = 7 /* Address of ElfNN_Rela relocations. */
	DT_RELASZ          = 8 /* Total size of ElfNN_Rela relocations. */
)

/* ELF64 relocations that need an addend field. */
type Rela64 struct {
	Off    uint64 /* Location to be relocated. */
	Info   uint64 /* Relocation type and symbol index. */
	Addend int64  /* Addend. */
}

// ELF64 Dynamic structure. The ".dynamic" section contains an array of them.
type Dyn64 struct {
	Tag int64  /* Entry type. */
	Val uint64 /* Integer/address value */
}

//export __dynamic_loader
func dynamicLoader(base uintptr, dyn *Dyn64) {
	var rela *Rela64
	relasz := uint64(0)

	if debugLoader {
		println("ASLR Base: ", base)
	}

	for dyn.Tag != DT_NULL {
		switch dyn.Tag {
		case DT_RELA:
			rela = (*Rela64)(unsafe.Pointer(base + uintptr(dyn.Val)))
		case DT_RELASZ:
			relasz = uint64(dyn.Val) / uint64(unsafe.Sizeof(Rela64{}))
		}

		ptr := uintptr(unsafe.Pointer(dyn))
		ptr += unsafe.Sizeof(Dyn64{})
		dyn = (*Dyn64)(unsafe.Pointer(ptr))
	}

	if rela == nil {
		runtimePanic("bad reloc")
	}

	if debugLoader {
		println("Sections to load: ", relasz)
	}

	for relasz > 0 && rela != nil {
		switch rela.Info {
		case R_AARCH64_RELATIVE:
			ptr := (*uint64)(unsafe.Pointer(base + uintptr(rela.Off)))
			*ptr = uint64(base + uintptr(rela.Addend))
		}

		rptr := uintptr(unsafe.Pointer(rela))
		rptr += unsafe.Sizeof(Rela64{})
		rela = (*Rela64)(unsafe.Pointer(rptr))
		relasz--
	}
}
