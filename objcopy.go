package main

import (
	"debug/elf"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/marcinbor85/gohex"
)

// ObjcopyError is an error returned by functions that act like objcopy.
type ObjcopyError struct {
	Op  string
	Err error
}

func (e ObjcopyError) Error() string {
	if e.Err == nil {
		return e.Op
	}
	return e.Op + ": " + e.Err.Error()
}

type ProgSlice []*elf.Prog

func (s ProgSlice) Len() int           { return len(s) }
func (s ProgSlice) Less(i, j int) bool { return s[i].Paddr < s[j].Paddr }
func (s ProgSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// ExtractROM extracts a firmware image and the first load address from the
// given ELF file. It tries to emulate the behavior of objcopy.
func ExtractROM(path string) (uint64, []byte, error) {
	f, err := elf.Open(path)
	if err != nil {
		return 0, nil, ObjcopyError{"failed to open ELF file to extract text segment", err}
	}
	defer f.Close()

	// The GNU objcopy command does the following for firmware extraction (from
	// the man page):
	// > When objcopy generates a raw binary file, it will essentially produce a
	// > memory dump of the contents of the input object file. All symbols and
	// > relocation information will be discarded. The memory dump will start at
	// > the load address of the lowest section copied into the output file.

	// Find the lowest section address.
	startAddr := ^uint64(0)
	for _, section := range f.Sections {
		if section.Type != elf.SHT_PROGBITS || section.Flags&elf.SHF_ALLOC == 0 {
			continue
		}
		if section.Addr < startAddr {
			startAddr = section.Addr
		}
	}

	progs := make(ProgSlice, 0, 2)
	for _, prog := range f.Progs {
		if prog.Type != elf.PT_LOAD || prog.Filesz == 0 {
			continue
		}
		progs = append(progs, prog)
	}
	if len(progs) == 0 {
		return 0, nil, ObjcopyError{"file does not contain ROM segments: " + path, nil}
	}
	sort.Sort(progs)

	var rom []byte
	for _, prog := range progs {
		if prog.Paddr != progs[0].Paddr+uint64(len(rom)) {
			return 0, nil, ObjcopyError{"ROM segments are non-contiguous: " + path, nil}
		}
		data, err := ioutil.ReadAll(prog.Open())
		if err != nil {
			return 0, nil, ObjcopyError{"failed to extract segment from ELF file: " + path, err}
		}
		rom = append(rom, data...)
	}
	if progs[0].Paddr < startAddr {
		// The lowest memory address is before the first section. This means
		// that there is some extra data loaded at the start of the image that
		// should be discarded.
		// Example: ELF files where .text doesn't start at address 0 because
		// there is a bootloader at the start.
		return startAddr, rom[startAddr-progs[0].Paddr:], nil
	} else {
		return progs[0].Paddr, rom, nil
	}
}

// Objcopy converts an ELF file to a different (simpler) output file format:
// .bin or .hex. It extracts only the .text section.
func Objcopy(infile, outfile string) error {
	f, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the .text segment.
	addr, data, err := ExtractROM(infile)
	if err != nil {
		return err
	}

	// Write to the file, in the correct format.
	switch filepath.Ext(outfile) {
	case ".gba":
		// The address is not stored in a .gba file.
		_, err := f.Write(data)
		return err
	case ".bin":
		// The address is not stored in a .bin file (therefore you
		// should use .hex files in most cases).
		_, err := f.Write(data)
		return err
	case ".hex":
		mem := gohex.NewMemory()
		err := mem.AddBinary(uint32(addr), data)
		if err != nil {
			return ObjcopyError{"failed to create .hex file", err}
		}
		mem.DumpIntelHex(f, 16) // TODO: handle error
		return nil
	default:
		panic("unreachable")
	}
}
