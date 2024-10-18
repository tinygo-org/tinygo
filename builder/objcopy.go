package builder

import (
	"debug/elf"
	"fmt"
	"os"

	"github.com/marcinbor85/gohex"
	"github.com/soypat/tinyboot/build/elfutil"
)

// maxPadBytes is the maximum allowed bytes to be padded in a rom extraction
// this value is currently defined by Nintendo Switch Page Alignment (4096 bytes)
const maxPadBytes = 4095

// objcopyError is an error returned by functions that act like objcopy.
type objcopyError struct {
	Op  string
	Err error
}

func (e objcopyError) Error() string {
	if e.Err == nil {
		return e.Op
	}
	return e.Op + ": " + e.Err.Error()
}

// extractROM extracts a firmware image and the first load address from the
// given ELF file. It tries to emulate the behavior of objcopy.
func extractROM(path string) (uint64, []byte, error) {
	f, err := elf.Open(path)
	if err != nil {
		return 0, nil, objcopyError{Op: "failed to open ELF file to extract text segment", Err: err}
	}
	defer f.Close()

	// The GNU objcopy command does the following for firmware extraction (from
	// the man page):
	// > When objcopy generates a raw binary file, it will essentially produce a
	// > memory dump of the contents of the input object file. All symbols and
	// > relocation information will be discarded. The memory dump will start at
	// > the load address of the lowest section copied into the output file.
	start, end, err := elfutil.ROMAddr(f)
	if err != nil {
		return 0, nil, objcopyError{Op: "failed to calculate ELF ROM addresses", Err: err}
	}
	err = elfutil.EnsureROMContiguous(f, start, end, maxPadBytes)
	if err != nil {
		return 0, nil, objcopyError{Op: "checking if ELF ROM contiguous", Err: err}
	}
	const (
		_ = 1 << (iota * 10)
		kB
		MB
		GB
	)
	const maxSize = 1 * GB
	if end-start > maxSize {
		return 0, nil, objcopyError{Op: fmt.Sprintf("obj size exceeds max %d/%d, bad ELF address calculation?", end-start, maxSize)}
	}

	ROM := make([]byte, end-start)
	n, err := elfutil.ReadROMAt(f, ROM, start)
	if err != nil {
		return 0, nil, objcopyError{Op: "reading ELF ROM", Err: err}
	} else if n != len(ROM) {
		return 0, nil, objcopyError{Op: "short ELF ROM read"}
	}
	return start, ROM, nil
}

// objcopy converts an ELF file to a different (simpler) output file format:
// .bin or .hex. It extracts only the .text section.
func objcopy(infile, outfile, binaryFormat string) error {
	f, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the .text segment.
	addr, data, err := extractROM(infile)
	if err != nil {
		return err
	}

	// Write to the file, in the correct format.
	switch binaryFormat {
	case "hex":
		// Intel hex file, includes the firmware start address.
		mem := gohex.NewMemory()
		err := mem.AddBinary(uint32(addr), data)
		if err != nil {
			return objcopyError{"failed to create .hex file", err}
		}
		return mem.DumpIntelHex(f, 16)
	case "bin":
		// The start address is not stored in raw firmware files (therefore you
		// should use .hex files in most cases).
		_, err := f.Write(data)
		return err
	default:
		panic("unreachable")
	}
}
