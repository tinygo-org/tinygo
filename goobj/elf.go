package goobj

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"strconv"
)

// CreateELF converts the Go object file to an ELF file and returns this as a
// byte slice. Errors should not normally happen.
func (obj *GoObj) CreateELF() ([]byte, error) {
	// Go through all symbols present in the Go object file and filter them. Not
	// all symbols should be included.
	// TODO: include DWARF data, such as line tables.
	local := obj.findLocalSymbols()
	exported := obj.findExportedSymbols()
	external := obj.findExternal(local, exported)
	numLocalSymbols := len(local)
	defined := append(local, exported...)
	symbols := append(defined, external...)

	// Determine number of sections.
	numSections := uint16(1) // section 0 is a dummy (all-zeroes) section
	symbolToSectionMapping := make(map[*symbolDef]uint16)
	symbolToIndexMapping := make(map[*symbolDef]uint32)
	for i, symbol := range symbols {
		symbolToSectionMapping[symbol] = numSections
		symbolToIndexMapping[symbol] = uint32(i + 1)
		if symbol.Type != 0 { // defined symbol
			numSections++
			if len(symbol.Relocs) != 0 {
				numSections++
			}
		}
	}
	numSections += 2 // .symtab and .strtab

	var e_machine elf.Machine
	switch obj.goarch {
	case "amd64":
		e_machine = elf.EM_X86_64
	case "arm64":
		e_machine = elf.EM_AARCH64
	default:
		return nil, fmt.Errorf("unknown GOARCH: %s", obj.goarch)
	}

	// This ELF file has the following format:
	// - ELF header
	// - ELF sections
	// - section data
	order := binary.LittleEndian
	symtabIndex := numSections - 2
	strtabIndex := numSections - 1
	headerSize := uint16(64)
	sectionSize := uint16(64)
	sectionOffset := headerSize // sections follow header directly

	dataStart := uint64(headerSize) + uint64(sectionSize)*uint64(numSections)
	dataBuf := &bytes.Buffer{}

	w := &bytes.Buffer{}
	// Write the ELF file header.
	// https://en.wikipedia.org/wiki/Executable_and_Linkable_Format#File_header
	w.WriteString(elf.ELFMAG)
	w.WriteByte(byte(elf.ELFCLASS64))
	w.WriteByte(byte(elf.ELFDATA2LSB))
	w.WriteByte(byte(elf.EV_CURRENT))
	w.WriteByte(0)                                // OS ABI
	w.WriteByte(0)                                // ABI version
	w.Write(make([]byte, 7))                      // padding
	binary.Write(w, order, elf.ET_REL)            // e_type
	binary.Write(w, order, e_machine)             // e_machine
	binary.Write(w, order, uint32(1))             // e_version
	binary.Write(w, order, uint64(0))             // e_entry (no entry point)
	binary.Write(w, order, uint64(0))             // e_phoff (no program header)
	binary.Write(w, order, uint64(sectionOffset)) // e_shoff
	binary.Write(w, order, uint32(0))             // e_flags
	binary.Write(w, order, uint16(headerSize))    // e_ehsize
	binary.Write(w, order, uint16(0))             // e_phentsize (no program header)
	binary.Write(w, order, uint16(0))             // e_phnum (no program header)
	binary.Write(w, order, uint16(sectionSize))   // e_shentsize: section entry size
	binary.Write(w, order, uint16(numSections))   // e_shnum: number of sections
	binary.Write(w, order, uint16(strtabIndex))   // e_shstrndx

	// Section strings.
	strtab := &bytes.Buffer{}
	strtab.WriteByte(0)
	makeString := func(name string) uint32 {
		index := uint32(strtab.Len())
		strtab.WriteString(name + "\x00")
		return index
	}

	// Write section headers, with first dummy section.
	w.Write(make([]byte, sectionSize)) // section index 0 is a dummy section

	// Write section for each symbol.
	for i, symbol := range symbols {
		if symbol.Type == 0 {
			continue // external symbol
		}
		sectionType := elf.SHT_NULL
		name := symbol.Name
		if name == "" {
			name = "__$" + strconv.Itoa(i)
		}
		var sectionName string
		var sectionFlags elf.SectionFlag
		switch symbol.Type {
		case symbolTEXT:
			sectionType = elf.SHT_PROGBITS
			sectionName = ".text." + name
			sectionFlags = elf.SHF_ALLOC | elf.SHF_EXECINSTR
		case symbolRODATA:
			sectionType = elf.SHT_PROGBITS
			sectionName = ".rodata." + name
			sectionFlags = elf.SHF_ALLOC
		default:
			// Unreachable, this has already checked above.
			panic("unknown symbol type")
		}
		binary.Write(w, order, makeString(sectionName))         // sh_name
		binary.Write(w, order, sectionType)                     // sh_type
		binary.Write(w, order, uint64(sectionFlags))            // sh_flags
		binary.Write(w, order, uint64(0))                       // sh_addr
		binary.Write(w, order, dataStart+uint64(dataBuf.Len())) // sh_offset
		binary.Write(w, order, uint64(len(symbol.Data)))        // sh_size
		binary.Write(w, order, uint32(0))                       // sh_link
		binary.Write(w, order, uint32(0))                       // sh_info
		binary.Write(w, order, uint64(symbol.Align))            // sh_addralign
		binary.Write(w, order, uint64(0))                       // sh_entsize
		dataBuf.Write(symbol.Data)
		if len(symbol.Relocs) != 0 {
			for dataBuf.Len()%8 != 0 {
				dataBuf.WriteByte(0) // align section
			}
			sectionStart := dataBuf.Len()
			for _, reloc := range symbol.Relocs {
				if reloc.Type == r_CALLIND {
					// This relocation seems to be emitted for indirect calls.
					// I wish it was common on ELF, that would make some forms
					// of static analysis a lot easier. However, it isn't, and
					// isn't necessary for ELF in most cases, so skip this
					// relocation.
					continue
				}
				symbolIndex, ok := symbolToIndexMapping[reloc.Symbol]
				if !ok {
					return nil, fmt.Errorf("relocation to unknown symbol: %s", reloc.Symbol.Name)
				}
				switch obj.goarch {
				case "amd64":
					switch reloc.Type {
					case r_CALL:
						r_info := elf.R_INFO(symbolIndex, uint32(elf.R_X86_64_PLT32))
						binary.Write(dataBuf, order, uint64(reloc.Offset))           // r_offset
						binary.Write(dataBuf, order, r_info)                         // r_info
						binary.Write(dataBuf, order, reloc.Addend-int64(reloc.Size)) // r_info
					case r_PCREL:
						r_info := elf.R_INFO(symbolIndex, uint32(elf.R_X86_64_PC32))
						binary.Write(dataBuf, order, uint64(reloc.Offset))           // r_offset
						binary.Write(dataBuf, order, r_info)                         // r_info
						binary.Write(dataBuf, order, reloc.Addend-int64(reloc.Size)) // r_info
					default:
						return nil, fmt.Errorf("unknown relocation type for amd64: %d", reloc.Type)
					}
				case "arm64":
					switch reloc.Type {
					case r_ADDRARM64:
						r_info := elf.R_INFO(symbolIndex, uint32(elf.R_AARCH64_ADR_PREL_PG_HI21))
						binary.Write(dataBuf, order, uint64(reloc.Offset)) // r_offset
						binary.Write(dataBuf, order, r_info)               // r_info
						binary.Write(dataBuf, order, reloc.Addend)         // r_info
						r_info = elf.R_INFO(symbolIndex, uint32(elf.R_AARCH64_ADD_ABS_LO12_NC))
						binary.Write(dataBuf, order, uint64(reloc.Offset+4)) // r_offset
						binary.Write(dataBuf, order, r_info)                 // r_info
						binary.Write(dataBuf, order, reloc.Addend)           // r_info
					case r_CALLARM64:
						r_info := elf.R_INFO(symbolIndex, uint32(elf.R_AARCH64_JUMP26))
						binary.Write(dataBuf, order, uint64(reloc.Offset)) // r_offset
						binary.Write(dataBuf, order, r_info)               // r_info
						binary.Write(dataBuf, order, reloc.Addend)         // r_info
					default:
						return nil, fmt.Errorf("unknown relocation type for arm64: %d", reloc.Type)
					}
				default:
					return nil, fmt.Errorf("unknown architecture: %s", obj.goarch)
				}
			}
			sectionSize := dataBuf.Len() - sectionStart
			binary.Write(w, order, makeString(".rela"+sectionName))        // sh_name
			binary.Write(w, order, elf.SHT_RELA)                           // sh_type
			binary.Write(w, order, uint64(0))                              // sh_flags
			binary.Write(w, order, uint64(0))                              // sh_addr
			binary.Write(w, order, dataStart+uint64(sectionStart))         // sh_offset
			binary.Write(w, order, uint64(sectionSize))                    // sh_size
			binary.Write(w, order, uint32(symtabIndex))                    // sh_link
			binary.Write(w, order, uint32(symbolToSectionMapping[symbol])) // sh_info
			binary.Write(w, order, uint64(1))                              // sh_addralign
			binary.Write(w, order, uint64(24))                             // sh_entsize
		}
	}

	// Create symbol section (.symtab).
	symtab := &bytes.Buffer{}
	symtab.Write(make([]byte, 24))
	for _, symbol := range symbols {
		sectionIndex := symbolToSectionMapping[symbol]
		switch symbol.Type {
		case symbolTEXT:
			binary.Write(symtab, order, makeString(mangleName(symbol.Name)))       // st_name
			binary.Write(symtab, order, elf.ST_INFO(elf.STB_GLOBAL, elf.STT_FUNC)) // st_info
			binary.Write(symtab, order, uint8(0))                                  // st_other
			binary.Write(symtab, order, sectionIndex)                              // st_shndx
			binary.Write(symtab, order, uint64(0))                                 // st_value
			binary.Write(symtab, order, uint64(len(symbol.Data)))                  // st_size
		case symbolRODATA:
			binary.Write(symtab, order, makeString(symbol.Name))                    // st_name
			binary.Write(symtab, order, elf.ST_INFO(elf.STB_LOCAL, elf.STT_OBJECT)) // st_info
			binary.Write(symtab, order, uint8(0))                                   // st_other
			binary.Write(symtab, order, sectionIndex)                               // st_shndx
			binary.Write(symtab, order, uint64(0))                                  // st_value
			binary.Write(symtab, order, uint64(len(symbol.Data)))                   // st_size
		case 0:
			// External symbol.
			binary.Write(symtab, order, makeString(mangleName(symbol.Name)))         // st_name
			binary.Write(symtab, order, elf.ST_INFO(elf.STB_GLOBAL, elf.STT_NOTYPE)) // st_info
			binary.Write(symtab, order, uint8(0))                                    // st_other
			binary.Write(symtab, order, uint16(0))                                   // st_shndx
			binary.Write(symtab, order, uint64(0))                                   // st_value
			binary.Write(symtab, order, uint64(0))                                   // st_size
		default:
			// Symbols are filtered at the beginning, so this is normally
			// unreachable.
			panic("unknown symbol type!")
		}
	}

	// Write symbol section (.symtab).
	binary.Write(w, order, makeString(".symtab"))           // sh_name
	binary.Write(w, order, elf.SHT_SYMTAB)                  // sh_type
	binary.Write(w, order, uint64(0))                       // sh_flags
	binary.Write(w, order, uint64(0))                       // sh_addr
	binary.Write(w, order, dataStart+uint64(dataBuf.Len())) // sh_offset
	binary.Write(w, order, uint64(symtab.Len()))            // sh_size
	binary.Write(w, order, uint32(strtabIndex))             // sh_link
	binary.Write(w, order, uint32(numLocalSymbols+1))       // sh_info
	binary.Write(w, order, uint64(1))                       // sh_addralign
	binary.Write(w, order, uint64(24))                      // sh_entsize
	dataBuf.Write(symtab.Bytes())

	// Write string table (.strtab).
	binary.Write(w, order, makeString(".strtab"))           // sh_name
	binary.Write(w, order, elf.SHT_STRTAB)                  // sh_type
	binary.Write(w, order, uint64(0))                       // sh_flags
	binary.Write(w, order, uint64(0))                       // sh_addr
	binary.Write(w, order, dataStart+uint64(dataBuf.Len())) // sh_offset
	binary.Write(w, order, uint64(strtab.Len()))            // sh_size
	binary.Write(w, order, uint32(0))                       // sh_link
	binary.Write(w, order, uint32(0))                       // sh_info
	binary.Write(w, order, uint64(1))                       // sh_addralign
	binary.Write(w, order, uint64(0))                       // sh_entsize
	dataBuf.Write(strtab.Bytes())

	// Write all section contents.
	w.Write(dataBuf.Bytes())

	return w.Bytes(), nil
}
