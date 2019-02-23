package main

import (
	"debug/elf"
	"os"
	"path/filepath"

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

// ExtractTextSegment returns the .text segment and the first address from the
// ELF file in the given path.
func ExtractTextSegment(path string) (uint64, []byte, error) {
	f, err := elf.Open(path)
	if err != nil {
		return 0, nil, ObjcopyError{"failed to open ELF file to extract text segment", err}
	}
	defer f.Close()

	text := f.Section(".text")
	if text == nil {
		return 0, nil, ObjcopyError{"file does not contain .text segment: " + path, nil}
	}
	data, err := text.Data()
	if err != nil {
		return 0, nil, ObjcopyError{"failed to extract .text segment from ELF file", err}
	}
	return text.Addr, data, nil
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
	addr, data, err := ExtractTextSegment(infile)
	if err != nil {
		return err
	}

	// Write to the file, in the correct format.
	switch filepath.Ext(outfile) {
	case ".bin":
		// The address is not stored in a .bin file (therefore you
		// should use .hex files in most cases).
		_, err := f.Write(data)
		return err
	case ".hex":
		mem := gohex.NewMemory()
		mem.SetStartAddress(uint32(addr)) // ignored in most cases (Intel-specific)
		err := mem.AddBinary(uint32(addr), data)
		if err != nil {
			return ObjcopyError{"failed to create .hex file", err}
		}
		mem.DumpIntelHex(f, 32) // TODO: handle error
		return nil
	default:
		panic("unreachable")
	}
}
