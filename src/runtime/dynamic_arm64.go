package runtime

import (
	"debug/elf"
	"unsafe"
)

const debugLoader = false

//export __dynamic_loader
func dynamicLoader(base uintptr, dyn *elf.Dyn64) {
	var rela *elf.Rela64
	relasz := uint64(0)

	if debugLoader {
		println("ASLR Base: ", base)
	}

	for dyn.Tag != int64(elf.DT_NULL) {
		switch elf.DynTag(dyn.Tag) {
		case elf.DT_RELA:
			rela = (*elf.Rela64)(unsafe.Pointer(base + uintptr(dyn.Val)))
		case elf.DT_RELASZ:
			relasz = uint64(dyn.Val) / uint64(unsafe.Sizeof(elf.Rela64{}))
		}

		ptr := uintptr(unsafe.Pointer(dyn))
		ptr += unsafe.Sizeof(elf.Dyn64{})
		dyn = (*elf.Dyn64)(unsafe.Pointer(ptr))
	}

	if rela == nil {
		runtimePanic("bad reloc")
	}
	if rela == nil {
		runtimePanic("bad reloc")
	}

	if debugLoader {
		println("Sections to load: ", relasz)
	}

	for relasz > 0 && rela != nil {
		switch elf.R_AARCH64(rela.Info) {
		case elf.R_AARCH64_RELATIVE:
			ptr := (*uint64)(unsafe.Pointer(base + uintptr(rela.Off)))
			*ptr = uint64(base + uintptr(rela.Addend))
		}

		rptr := uintptr(unsafe.Pointer(rela))
		rptr += unsafe.Sizeof(elf.Rela64{})
		rela = (*elf.Rela64)(unsafe.Pointer(rptr))
		relasz--
	}
}
