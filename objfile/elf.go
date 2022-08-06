package objfile

// This file provides an ELF file writer. It can produce both ELF32 and ELF64
// files using generics.

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"sort"
	"unsafe"
)

// ELF represents the information in an (unsaved) ELF file.
type ELF struct {
	machine elf.Machine
	symbols []*elfSymbol
}

// NewELF creates a new ELF file with the given ELF machine. The ELF file type
// (32-bit or 64-bit) is based on the machine type.
func NewELF(machine elf.Machine) *ELF {
	return &ELF{
		machine: machine,
	}
}

// AddSymbol adds a new symbol to the ELF file. The section can be either "text"
// or "rodata". The symbol type (func or object) is determined from the section
// type. It returns a symbol index that can be used for AddReloc.
func (f *ELF) AddSymbol(name, section string, linkage Linkage, data []byte) int {
	index := len(f.symbols)
	f.symbols = append(f.symbols, &elfSymbol{
		name:    name,
		section: section,
		linkage: linkage,
		data:    data,
	})
	return index
}

// AddReloc adds a new relocation to the symbol with the given index.
func (f *ELF) AddReloc(index int, offset uint64, reloc Reloc, symbol string, addend int64) {
	sym := f.symbols[index]
	sym.relocs = append(sym.relocs, relocInfo{
		offset: offset,
		reloc:  reloc,
		symbol: symbol,
		addend: addend,
	})
}

// Bytes returns the ELF file as a byte slice. It is either ELF32 or ELF64
// depending on the bitness of the architecture.
func (f *ELF) Bytes() []byte {
	switch f.machine {
	case elf.EM_AARCH64, elf.EM_X86_64:
		return makeELF[uint64](f.machine, f.symbols)
	default:
		return makeELF[uint32](f.machine, f.symbols)
	}
}

type elfSymbol struct {
	name    string
	section string
	linkage Linkage
	data    []byte
	relocs  []relocInfo
}

type relocInfo struct {
	offset uint64
	reloc  Reloc
	symbol string
	addend int64
}

// A generic address type, representing either 32-bit (for ELF32) or 64-bit (for
// ELF64) addresses.
type address interface {
	uint32 | uint64
}

// This is ELF32_Ehdr or ELF64_Ehdr. The struct layout matches between the two
// versions.
type elfHeader[addr address] struct {
	ident_magic      [4]byte
	ident_class      elf.Class
	ident_data       elf.Data
	ident_version    elf.Version
	ident_osabi      elf.OSABI
	ident_abiversion byte
	_                [7]byte // padding
	e_type           elf.Type
	e_machine        elf.Machine
	e_version        uint32
	e_entry          addr
	e_phoff          addr
	e_shoff          addr
	e_flags          uint32
	e_ehsize         uint16
	e_phentsize      uint16
	e_phnum          uint16
	e_shentsize      uint16
	e_shnum          uint16
	e_shstrndx       uint16
}

// This is ELF32_Shdr or ELF64_Shdr.
type elfSectionHeader[addr address] struct {
	sh_name      uint32
	sh_type      elf.SectionType
	sh_flags     addr
	sh_addr      addr
	sh_offset    addr
	sh_size      addr
	sh_link      uint32
	sh_info      uint32
	sh_addralign addr
	sh_entsize   addr
}

type elfSection[addr address] struct {
	header elfSectionHeader[addr]
	data   []byte
}

// This is ELF32_Rela or ELF64_Rela.
type elfReloc[addr address] struct {
	r_offset addr
	r_info   addr
	r_addend addr // technically a signed integer but it doesn't matter
}

// This is ELF64_Sym. Unfortunately, it doesn't match the layout of ELF32_Sym.
type elfSymtabSymbol[addr address] struct {
	// This struct follows the ELF64 format.
	st_name  uint32
	st_info  uint8
	st_other uint8
	st_shndx uint16
	st_value addr
	st_size  addr
}

// This is ELF32_Sym.
type elf32SymtabSymbol struct {
	st_name  uint32
	st_value uint32
	st_size  uint32
	st_info  uint8
	st_other uint8
	st_shndx uint16
}

type elfStringTable struct {
	data    []byte
	indices map[string]uint32
	names   map[uint32]string
}

func newELFStringTable() *elfStringTable {
	return &elfStringTable{
		data:    []byte{0},
		indices: make(map[string]uint32),
		names:   make(map[uint32]string),
	}
}

func (tab *elfStringTable) add(s string) uint32 {
	if idx, ok := tab.indices[s]; ok {
		return uint32(idx)
	}
	idx := uint32(len(tab.data))
	tab.data = append(tab.data, []byte(s)...)
	tab.data = append(tab.data, 0)
	tab.indices[s] = idx
	tab.names[idx] = s
	return idx
}

func (tab *elfStringTable) get(idx uint32) string {
	return tab.names[idx]
}

func makeStringTableSection[addr address](tab *elfStringTable, name uint32) elfSection[addr] {
	// This function has to be a separate function (not a method) because of a
	// limitation of Go generics (as of Go 1.18).
	return elfSection[addr]{
		header: elfSectionHeader[addr]{
			sh_name:      name,
			sh_type:      elf.SHT_STRTAB,
			sh_flags:     0,
			sh_addr:      0,
			sh_offset:    0, // to be filled in
			sh_size:      addr(len(tab.data)),
			sh_link:      0,
			sh_info:      0,
			sh_addralign: 1,
			sh_entsize:   0,
		},
		data: tab.data,
	}
}

// Return an ELF file as a byte slice. It returns either an ELF32 or ELF64 file,
// depending on the addr type.
func makeELF[addr address](machine elf.Machine, symbols []*elfSymbol) []byte {
	isELF64 := unsafe.Sizeof(addr(0)) == 8
	elfSectionHeaderSize := int(unsafe.Sizeof(elfSectionHeader[addr]{}))

	shstrtab := newELFStringTable()
	strtab := newELFStringTable()

	symtab := []elfSymtabSymbol[addr]{
		{st_info: elf.ST_INFO(elf.STB_LOCAL, 0)}, // zero entry
	}

	// Create one section for each symbol.
	var sections []elfSection[addr]
	sections = append(sections, elfSection[addr]{}) // zero index section
	symbolInSymtab := make(map[string]struct{})
	symbolSectionIndices := make(map[string]uint16)
	for _, symbol := range symbols {
		sectionIndex := uint16(len(sections))
		symbolSectionIndices[symbol.name] = sectionIndex
		typ := elf.STT_OBJECT
		var flags elf.SectionFlag
		switch symbol.section {
		case "text":
			typ = elf.STT_FUNC
			flags = elf.SHF_ALLOC | elf.SHF_EXECINSTR
		case "rodata":
			flags = elf.SHF_ALLOC
		}
		sections = append(sections, elfSection[addr]{
			header: elfSectionHeader[addr]{
				sh_name:      shstrtab.add("." + symbol.section + "." + symbol.name),
				sh_type:      elf.SHT_PROGBITS,
				sh_flags:     addr(flags),
				sh_addr:      0,
				sh_offset:    0, // to be filled in
				sh_size:      addr(len(symbol.data)),
				sh_link:      0,
				sh_info:      0,
				sh_addralign: 1,
				sh_entsize:   0,
			},
			data: symbol.data,
		})
		symbolInSymtab[symbol.name] = struct{}{}
		symtab = append(symtab, elfSymtabSymbol[addr]{
			st_name:  strtab.add(symbol.name),
			st_info:  elf.ST_INFO(symbol.linkage.elf(), typ),
			st_other: uint8(elf.STV_DEFAULT),
			st_shndx: sectionIndex,
			st_value: 0,
			st_size:  addr(len(symbol.data)),
		})
	}

	// Add the referenced but undefined symbol to the symbol table.
	for _, symbol := range symbols {
		for _, reloc := range symbol.relocs {
			if _, ok := symbolInSymtab[reloc.symbol]; !ok {
				symbolInSymtab[reloc.symbol] = struct{}{}
				symtab = append(symtab, elfSymtabSymbol[addr]{
					st_name:  strtab.add(reloc.symbol),
					st_info:  elf.ST_INFO(elf.STB_GLOBAL, elf.STT_NOTYPE),
					st_other: uint8(elf.STV_DEFAULT),
					st_shndx: 0,
					st_value: 0,
					st_size:  0,
				})
			}
		}
	}

	// Sort local symbols before global and weak symbols. This is a requirement
	// of ELF files (see documentation on the sh_info field).
	sort.SliceStable(symtab, func(i, j int) bool {
		local_i := elf.ST_BIND(symtab[i].st_info) == elf.STB_LOCAL
		local_j := elf.ST_BIND(symtab[j].st_info) == elf.STB_LOCAL
		return local_i && !local_j
	})

	// Create the contents of the symtab section.
	symtabData := &bytes.Buffer{}
	if isELF64 {
		binary.Write(symtabData, binary.LittleEndian, symtab)
	} else {
		// ELF32
		// We have to do this because the structs are slightly different, and Go
		// doesn't yet support accessing struct fields of generic structs (where
		// a given field is present in all possible structs). It might get fixed
		// in Go 1.20.
		// https://github.com/golang/go/issues/48522
		for _, sym := range symtab {
			binary.Write(symtabData, binary.LittleEndian, elf32SymtabSymbol{
				st_name:  sym.st_name,
				st_value: uint32(sym.st_value),
				st_size:  uint32(sym.st_size),
				st_info:  sym.st_info,
				st_other: sym.st_other,
				st_shndx: sym.st_shndx,
			})
		}
	}

	// Create the symtab section header.
	symtabSection := len(sections)
	firstNonLocal := len(symtab)
	for i, sym := range symtab {
		if elf.ST_BIND(sym.st_info) != elf.STB_LOCAL {
			firstNonLocal = i
			break
		}
	}
	sections = append(sections, elfSection[addr]{
		header: elfSectionHeader[addr]{
			sh_name:      shstrtab.add(".symtab"),
			sh_type:      elf.SHT_SYMTAB,
			sh_flags:     0,
			sh_addr:      0,
			sh_offset:    0, // to be filled in
			sh_size:      addr(symtabData.Len()),
			sh_link:      0, // .strtab, to be filled in later
			sh_info:      uint32(firstNonLocal),
			sh_addralign: 1,
			sh_entsize:   addr(unsafe.Sizeof(elfSymtabSymbol[addr]{})),
		},
		data: symtabData.Bytes(),
	})

	// Create relocation sections. There has to be a relocation section for each
	// symbol section that contains relocations.
	symbolSymtabIndices := make(map[string]int)
	for i, sym := range symtab {
		symbolSymtabIndices[strtab.get(sym.st_name)] = i
	}
	for _, symbol := range symbols {
		var relocs []elfReloc[addr]
		for _, reloc := range symbol.relocs {
			symtabIndex := symbolSymtabIndices[reloc.symbol]
			addReloc := func(offset uint64, reloc uint32, addend int64) {
				var r_info addr
				if isELF64 {
					r_info = addr(elf.R_INFO(uint32(symtabIndex), reloc))
				} else {
					r_info = addr(elf.R_INFO32(uint32(symtabIndex), reloc))
				}
				relocs = append(relocs, elfReloc[addr]{
					r_offset: addr(offset),
					r_info:   r_info,
					r_addend: addr(addend),
				})
			}
			switch machine {
			case elf.EM_386:
				switch reloc.reloc {
				case RelocCALL:
					addReloc(reloc.offset, uint32(elf.R_386_PC32), reloc.addend-4)
				default:
					panic("unknown relocation " + reloc.reloc.String() + " for 386")
				}
			case elf.EM_X86_64:
				switch reloc.reloc {
				case RelocCALL:
					addReloc(reloc.offset, uint32(elf.R_X86_64_PLT32), reloc.addend-4)
				case RelocPCREL:
					addReloc(reloc.offset, uint32(elf.R_X86_64_PC32), reloc.addend-4)
				default:
					panic("unknown relocation " + reloc.reloc.String() + " for amd64")
				}
			case elf.EM_ARM:
				switch reloc.reloc {
				case RelocADDR:
					addReloc(reloc.offset, uint32(elf.R_ARM_ABS32), reloc.addend)
				case RelocCALL:
					// TODO: I think this should be R_ARM_CALL for bl.
					// R_ARM_JUMP24 is only for b and bl<cond>.
					addReloc(reloc.offset, uint32(elf.R_ARM_JUMP24), reloc.addend)
				default:
					panic("unknown relocation " + reloc.reloc.String() + " for arm")
				}
			case elf.EM_AARCH64:
				switch reloc.reloc {
				case RelocADDR:
					addReloc(reloc.offset, uint32(elf.R_AARCH64_ADR_PREL_PG_HI21), reloc.addend)
					addReloc(reloc.offset+4, uint32(elf.R_AARCH64_ADD_ABS_LO12_NC), reloc.addend)
				case RelocCALL:
					addReloc(reloc.offset, uint32(elf.R_AARCH64_CALL26), reloc.addend)
				case RelocPCREL_LDST64:
					addReloc(reloc.offset, uint32(elf.R_AARCH64_ADR_PREL_PG_HI21), reloc.addend)
					addReloc(reloc.offset+4, uint32(elf.R_AARCH64_LDST64_ABS_LO12_NC), reloc.addend)
				default:
					panic("unknown relocation " + reloc.reloc.String() + " for arm64")
				}
			default:
				panic("unknown machine " + machine.String())
			}
		}
		if len(relocs) != 0 {
			var relocBuf bytes.Buffer
			err := binary.Write(&relocBuf, binary.LittleEndian, relocs)
			if err != nil {
				panic(err)
			}
			sections = append(sections, elfSection[addr]{
				header: elfSectionHeader[addr]{
					sh_name:      shstrtab.add(".rela." + symbol.section + "." + symbol.name),
					sh_type:      elf.SHT_RELA,
					sh_flags:     0,
					sh_addr:      0,
					sh_offset:    0, // to be filled in
					sh_size:      addr(relocBuf.Len()),
					sh_link:      uint32(symtabSection),
					sh_info:      uint32(symbolSectionIndices[symbol.name]),
					sh_addralign: 1,
					sh_entsize:   addr(unsafe.Sizeof(elfReloc[addr]{})),
				},
				data: relocBuf.Bytes(),
			})
		}
	}

	// Create the string table sections (.strtab and .shstrtab).
	strtabSection := len(sections)
	sections = append(sections, makeStringTableSection[addr](strtab, shstrtab.add(".strtab")))
	shstrtabSection := len(sections)
	sections = append(sections, makeStringTableSection[addr](shstrtab, shstrtab.add(".shstrtab")))

	// Patch the symtab section: it needs to know the associated strtab section.
	sections[symtabSection].header.sh_link = uint32(strtabSection)

	// Write out the ELF header.
	out := &bytes.Buffer{}
	elfHeaderSize := int(unsafe.Sizeof(elfHeader[addr]{}))
	header := elfHeader[addr]{
		ident_magic:      [4]byte{'\x7F', 'E', 'L', 'F'},
		ident_data:       elf.ELFDATA2LSB,
		ident_version:    elf.EV_CURRENT,
		ident_osabi:      elf.ELFOSABI_LINUX,
		ident_abiversion: 0,
		e_type:           elf.ET_REL,
		e_machine:        machine,
		e_version:        1,
		e_entry:          0,
		e_phoff:          0,
		e_shoff:          addr(elfHeaderSize),
		e_flags:          0,
		e_ehsize:         uint16(elfHeaderSize),
		e_phentsize:      0,
		e_phnum:          0,
		e_shentsize:      uint16(elfSectionHeaderSize),
		e_shnum:          uint16(len(sections)),
		e_shstrndx:       uint16(shstrtabSection),
	}
	if isELF64 {
		header.ident_class = elf.ELFCLASS64
	} else {
		header.ident_class = elf.ELFCLASS32
	}
	binary.Write(out, binary.LittleEndian, header)

	// Adjust section offsets.
	sectionHeaderSize := elfSectionHeaderSize * len(sections)
	dataOffset := elfHeaderSize + sectionHeaderSize
	for i := range sections {
		sections[i].header.sh_offset = addr(dataOffset)
		dataOffset += len(sections[i].data)
	}

	// Write section header.
	for _, section := range sections {
		binary.Write(out, binary.LittleEndian, section.header)
	}

	// Write contents of sections.
	for _, section := range sections {
		out.Write(section.data)
	}

	return out.Bytes()
}
