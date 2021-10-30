package builder

import (
	"debug/dwarf"
	"debug/elf"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
	"tinygo.org/x/go-llvm"
)

// programSize contains size statistics per package of a compiled program.
type programSize struct {
	Packages map[string]packageSize
	Code     uint64
	ROData   uint64
	Data     uint64
	BSS      uint64
}

// sortedPackageNames returns the list of package names (ProgramSize.Packages)
// sorted alphabetically.
func (ps *programSize) sortedPackageNames() []string {
	names := make([]string, 0, len(ps.Packages))
	for name := range ps.Packages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Flash usage in regular microcontrollers.
func (ps *programSize) Flash() uint64 {
	return ps.Code + ps.ROData + ps.Data
}

// Static RAM usage in regular microcontrollers.
func (ps *programSize) RAM() uint64 {
	return ps.Data + ps.BSS
}

// packageSize contains the size of a package, calculated from the linked object
// file.
type packageSize struct {
	Code   uint64
	ROData uint64
	Data   uint64
	BSS    uint64
}

// Flash usage in regular microcontrollers.
func (ps *packageSize) Flash() uint64 {
	return ps.Code + ps.ROData + ps.Data
}

// Static RAM usage in regular microcontrollers.
func (ps *packageSize) RAM() uint64 {
	return ps.Data + ps.BSS
}

// A mapping of a single chunk of code or data to a file path.
type addressLine struct {
	Address    uint64
	Length     uint64 // length of this chunk
	File       string // file path as stored in DWARF
	IsVariable bool   // true if this is a variable (or constant), false if it is code
}

// Regular expressions to match particular symbol names. These are not stored as
// DWARF variables because they have no mapping to source code global variables.
var (
	// Various globals that aren't a variable but nonetheless need to be stored
	// somewhere:
	//   alloc:  heap allocations during init interpretation
	//   pack:   data created when storing a constant in an interface for example
	//   string: buffer behind strings
	packageSymbolRegexp = regexp.MustCompile(`\$(alloc|pack|string)(\.[0-9]+)?$`)

	// Reflect sidetables. Created by the reflect lowering pass.
	// See src/reflect/sidetables.go.
	reflectDataRegexp = regexp.MustCompile(`^reflect\.[a-zA-Z]+Sidetable$`)
)

// readProgramSizeFromDWARF reads the source location for each line of code and
// each variable in the program, as far as this is stored in the DWARF debug
// information.
func readProgramSizeFromDWARF(data *dwarf.Data) ([]addressLine, error) {
	r := data.Reader()
	var lines []*dwarf.LineFile
	var addresses []addressLine
	for {
		e, err := r.Next()
		if err != nil {
			return nil, err
		}
		if e == nil {
			break
		}
		switch e.Tag {
		case dwarf.TagCompileUnit:
			// Found a compile unit.
			// We can read the .debug_line section using it, which contains a
			// mapping for most instructions to their file/line/column - even
			// for inlined functions!
			lr, err := data.LineReader(e)
			if err != nil {
				return nil, err
			}
			lines = lr.Files()
			var lineEntry = dwarf.LineEntry{
				EndSequence: true,
			}

			// Line tables are organized as sequences of line entries until an
			// end sequence. A single line table can contain multiple such
			// sequences. The last line entry is an EndSequence to indicate the
			// end.
			for {
				// Read the next .debug_line entry.
				prevLineEntry := lineEntry
				err := lr.Next(&lineEntry)
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}

				if prevLineEntry.EndSequence && lineEntry.Address == 0 {
					// Tombstone value. This symbol has been removed, for
					// example by the --gc-sections linker flag. It is still
					// here in the debug information because the linker can't
					// just remove this reference.
					// Read until the next EndSequence so that this sequence is
					// skipped.
					// For more details, see (among others):
					// https://reviews.llvm.org/D84825
					for {
						err := lr.Next(&lineEntry)
						if err != nil {
							return nil, err
						}
						if lineEntry.EndSequence {
							break
						}
					}
				}

				if !prevLineEntry.EndSequence {
					// The chunk describes the code from prevLineEntry to
					// lineEntry.
					line := addressLine{
						Address: prevLineEntry.Address,
						Length:  lineEntry.Address - prevLineEntry.Address,
						File:    prevLineEntry.File.Name,
					}
					if line.Length != 0 {
						addresses = append(addresses, line)
					}
				}
			}
		case dwarf.TagVariable:
			// Global variable (or constant). Most of these are not actually
			// stored in the binary, because they have been optimized out. Only
			// the ones with a location are still present.
			r.SkipChildren()

			file := e.AttrField(dwarf.AttrDeclFile)
			location := e.AttrField(dwarf.AttrLocation)
			globalType := e.AttrField(dwarf.AttrType)
			if file == nil || location == nil || globalType == nil {
				// Doesn't contain the requested information.
				continue
			}

			// Try to parse the location. While this could in theory be a very
			// complex expression, usually it's just a DW_OP_addr opcode
			// followed by an address.
			locationCode := location.Val.([]uint8)
			if locationCode[0] != 3 { // DW_OP_addr
				continue
			}
			var addr uint64
			switch len(locationCode) {
			case 1 + 2:
				addr = uint64(binary.LittleEndian.Uint16(locationCode[1:]))
			case 1 + 4:
				addr = uint64(binary.LittleEndian.Uint32(locationCode[1:]))
			case 1 + 8:
				addr = binary.LittleEndian.Uint64(locationCode[1:])
			default:
				continue // unknown address
			}

			// Parse the type of the global variable, which (importantly)
			// contains the variable size. We're not interested in the type,
			// only in the size.
			typ, err := data.Type(globalType.Val.(dwarf.Offset))
			if err != nil {
				return nil, err
			}

			addresses = append(addresses, addressLine{
				Address:    addr,
				Length:     uint64(typ.Size()),
				File:       lines[file.Val.(int64)].Name,
				IsVariable: true,
			})
		default:
			r.SkipChildren()
		}
	}
	return addresses, nil
}

// loadProgramSize calculate a program/data size breakdown of each package for a
// given ELF file.
// If the file doesn't contain DWARF debug information, the returned program
// size will still have valid summaries but won't have complete size information
// per package.
func loadProgramSize(path string, packagePathMap map[string]string) (*programSize, error) {
	// Open the ELF file.
	file, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// This stores all chunks of addresses found in the binary.
	var addresses []addressLine

	// Read DWARF information.
	// Intentionally ignoring the error here: if DWARF couldn't be loaded, just
	// don't load symbol information from DWARF metadata.
	data, _ := file.DWARF()
	if file.Machine == elf.EM_AVR && strings.Split(llvm.Version, ".")[0] <= "10" {
		// Hack to work around broken DWARF support for AVR in LLVM 10.
		// This should be removed once support for LLVM 10 is dropped.
		data = nil
	}
	if data != nil {
		addresses, err = readProgramSizeFromDWARF(data)
		if err != nil {
			// However, _do_ report an error here. Something must have gone
			// wrong while trying to parse DWARF data.
			return nil, err
		}
	}

	// Read the ELF symbols for some more chunks of location information.
	// Some globals (such as strings) aren't stored in the DWARF debug
	// information and therefore need to be obtained in a different way.
	allSymbols, err := file.Symbols()
	if err != nil {
		return nil, err
	}
	for _, symbol := range allSymbols {
		symType := elf.ST_TYPE(symbol.Info)
		if symbol.Size == 0 {
			continue
		}
		if symType != elf.STT_FUNC && symType != elf.STT_OBJECT && symType != elf.STT_NOTYPE {
			continue
		}
		section := file.Sections[symbol.Section]
		if section.Flags&elf.SHF_ALLOC == 0 {
			continue
		}
		if packageSymbolRegexp.MatchString(symbol.Name) || reflectDataRegexp.MatchString(symbol.Name) {
			addresses = append(addresses, addressLine{
				Address:    symbol.Value,
				Length:     symbol.Size,
				File:       symbol.Name,
				IsVariable: true,
			})
		}
	}

	// Sort the slice of address chunks by address, so that we can iterate
	// through it to calculate section sizes.
	sort.Slice(addresses, func(i, j int) bool {
		if addresses[i].Address == addresses[j].Address {
			// Very rarely, there might be duplicate addresses.
			// If that happens, sort the largest chunks first.
			return addresses[i].Length > addresses[j].Length
		}
		return addresses[i].Address < addresses[j].Address
	})

	// Now finally determine the binary/RAM size usage per package by going
	// through each allocated section.
	sizes := make(map[string]packageSize)
	for _, section := range file.Sections {
		if section.Flags&elf.SHF_ALLOC == 0 {
			continue
		}
		if section.Type != elf.SHT_PROGBITS && section.Type != elf.SHT_NOBITS {
			continue
		}
		if section.Name == ".stack" {
			// This is a bit ugly, but I don't think there is a way to mark the
			// stack section in a linker script.
			// We store the C stack as a pseudo-section.
			sizes["C stack"] = packageSize{
				BSS: section.Size,
			}
			continue
		}
		if section.Type == elf.SHT_NOBITS {
			// .bss
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				field.BSS += size
				sizes[path] = field
			}, packagePathMap)
		} else if section.Type == elf.SHT_PROGBITS && section.Flags&elf.SHF_EXECINSTR != 0 {
			// .text
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				if isVariable {
					field.ROData += size
				} else {
					field.Code += size
				}
				sizes[path] = field
			}, packagePathMap)
		} else if section.Type == elf.SHT_PROGBITS && section.Flags&elf.SHF_WRITE != 0 {
			// .data
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				field.Data += size
				sizes[path] = field
			}, packagePathMap)
		} else if section.Type == elf.SHT_PROGBITS {
			// .rodata
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				field.ROData += size
				sizes[path] = field
			}, packagePathMap)
		}
	}

	// ...and summarize the results.
	program := &programSize{
		Packages: sizes,
	}
	for _, pkg := range sizes {
		program.Code += pkg.Code
		program.ROData += pkg.ROData
		program.Data += pkg.Data
		program.BSS += pkg.BSS
	}
	return program, nil
}

// readSection determines for each byte in this section to which package it
// belongs. It reports this usage through the addSize callback.
func readSection(section *elf.Section, addresses []addressLine, addSize func(string, uint64, bool), packagePathMap map[string]string) {
	// The addr variable tracks at which address we are while going through this
	// section. We start at the beginning.
	addr := section.Addr
	sectionEnd := section.Addr + section.Size
	for _, line := range addresses {
		if line.Address < section.Addr || line.Address+line.Length >= sectionEnd {
			// Check that this line is entirely within the section.
			// Don't bother dealing with line entries that cross sections (that
			// seems rather unlikely anyway).
			continue
		}
		if addr < line.Address {
			// There is a gap: there is a space between the current and the
			// previous line entry.
			addSize("(unknown)", line.Address-addr, false)
		}
		if addr > line.Address+line.Length {
			// The current line is already covered by a previous line entry.
			// Simply skip it.
			continue
		}
		// At this point, addr falls within the current line (probably at the
		// start).
		length := line.Length
		if addr > line.Address {
			// There is some overlap: the previous line entry already covered
			// part of this line entry. So reduce the length to add to the
			// remaining bit of the line entry.
			length = line.Length - (addr - line.Address)
		}
		// Finally, mark this chunk of memory as used by the given package.
		addSize(findPackagePath(line.File, packagePathMap), length, line.IsVariable)
		addr = line.Address + line.Length
	}
	if addr < sectionEnd {
		// There is a gap at the end of the section.
		addSize("(unknown)", sectionEnd-addr, false)
	}
}

// findPackagePath returns the Go package (or a pseudo package) for the given
// path. It uses some heuristics, for example for some C libraries.
func findPackagePath(path string, packagePathMap map[string]string) string {
	// Check whether this path is part of one of the compiled packages.
	packagePath, ok := packagePathMap[filepath.Dir(path)]
	if !ok {
		if strings.HasPrefix(path, filepath.Join(goenv.Get("TINYGOROOT"), "lib")) {
			// Emit C libraries (in the lib subdirectory of TinyGo) as a single
			// package, with a "C" prefix. For example: "C compiler-rt" for the
			// compiler runtime library from LLVM.
			packagePath = "C " + strings.Split(strings.TrimPrefix(path, filepath.Join(goenv.Get("TINYGOROOT"), "lib")), string(os.PathSeparator))[1]
		} else if packageSymbolRegexp.MatchString(path) {
			// Parse symbol names like main$alloc or runtime$string.
			packagePath = path[:strings.LastIndex(path, "$")]
		} else if reflectDataRegexp.MatchString(path) {
			// Parse symbol names like reflect.structTypesSidetable.
			packagePath = "Go reflect data"
		} else {
			// This is some other path. Not sure what it is, so just emit its directory.
			packagePath = filepath.Dir(path) // fallback
		}
	}
	return packagePath
}
