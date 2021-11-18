package builder

import (
	"bytes"
	"debug/dwarf"
	"debug/elf"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/aykevl/go-wasm"
	"github.com/tinygo-org/tinygo/goenv"
)

// Set to true to print extra debug logs.
const sizesDebug = false

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

// Sections defined in the input file. This struct defines them in a
// filetype-agnostic way but roughly follow the ELF types (.text, .data, .bss,
// etc).
type memorySection struct {
	Type    memoryType
	Address uint64
	Size    uint64
}

type memoryType int

const (
	memoryCode memoryType = iota + 1
	memoryData
	memoryROData
	memoryBSS
	memoryStack
)

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
func readProgramSizeFromDWARF(data *dwarf.Data, codeOffset uint64) ([]addressLine, error) {
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
						Address: prevLineEntry.Address + codeOffset,
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
	// Open the binary file.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// This stores all chunks of addresses found in the binary.
	var addresses []addressLine

	// Load the binary file, which could be in a number of file formats.
	var sections []memorySection
	if file, err := elf.NewFile(f); err == nil {
		// Read DWARF information. The error is intentionally ignored.
		data, _ := file.DWARF()
		if data != nil {
			addresses, err = readProgramSizeFromDWARF(data, 0)
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
			if symbol.Section >= elf.SHN_LORESERVE {
				// Not a regular section, so skip it.
				// One example is elf.SHN_ABS, which is used for symbols
				// declared with an absolute value such as the memset function
				// on the ESP32 which is defined in the mask ROM.
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

		// Load allocated sections.
		for _, section := range file.Sections {
			if section.Flags&elf.SHF_ALLOC == 0 {
				continue
			}
			if section.Type == elf.SHT_NOBITS {
				if section.Name == ".stack" {
					// TinyGo emits stack sections on microcontroller using the
					// ".stack" name.
					// This is a bit ugly, but I don't think there is a way to
					// mark the stack section in a linker script.
					sections = append(sections, memorySection{
						Address: section.Addr,
						Size:    section.Size,
						Type:    memoryStack,
					})
				} else {
					// Regular .bss section.
					sections = append(sections, memorySection{
						Address: section.Addr,
						Size:    section.Size,
						Type:    memoryBSS,
					})
				}
			} else if section.Type == elf.SHT_PROGBITS && section.Flags&elf.SHF_EXECINSTR != 0 {
				// .text
				sections = append(sections, memorySection{
					Address: section.Addr,
					Size:    section.Size,
					Type:    memoryCode,
				})
			} else if section.Type == elf.SHT_PROGBITS && section.Flags&elf.SHF_WRITE != 0 {
				// .data
				sections = append(sections, memorySection{
					Address: section.Addr,
					Size:    section.Size,
					Type:    memoryData,
				})
			} else if section.Type == elf.SHT_PROGBITS {
				// .rodata
				sections = append(sections, memorySection{
					Address: section.Addr,
					Size:    section.Size,
					Type:    memoryROData,
				})
			}
		}
	} else if file, err := pe.NewFile(f); err == nil {
		// Read DWARF information. The error is intentionally ignored.
		data, _ := file.DWARF()
		if data != nil {
			addresses, err = readProgramSizeFromDWARF(data, 0)
			if err != nil {
				// However, _do_ report an error here. Something must have gone
				// wrong while trying to parse DWARF data.
				return nil, err
			}
		}

		// Read COFF sections.
		optionalHeader := file.OptionalHeader.(*pe.OptionalHeader64)
		for _, section := range file.Sections {
			// For more information:
			// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-image_section_header
			const (
				IMAGE_SCN_CNT_CODE             = 0x00000020
				IMAGE_SCN_CNT_INITIALIZED_DATA = 0x00000040
				IMAGE_SCN_MEM_DISCARDABLE      = 0x02000000
				IMAGE_SCN_MEM_READ             = 0x40000000
				IMAGE_SCN_MEM_WRITE            = 0x80000000
			)
			if section.Characteristics&IMAGE_SCN_MEM_DISCARDABLE != 0 {
				// Debug sections, etc.
				continue
			}
			address := uint64(section.VirtualAddress) + optionalHeader.ImageBase
			if section.Characteristics&IMAGE_SCN_CNT_CODE != 0 {
				// .text
				sections = append(sections, memorySection{
					Address: address,
					Size:    uint64(section.VirtualSize),
					Type:    memoryCode,
				})
			} else if section.Characteristics&IMAGE_SCN_CNT_INITIALIZED_DATA != 0 {
				if section.Characteristics&IMAGE_SCN_MEM_WRITE != 0 {
					// .data
					sections = append(sections, memorySection{
						Address: address,
						Size:    uint64(section.Size),
						Type:    memoryData,
					})
					if section.Size < section.VirtualSize {
						// Equivalent of a .bss section.
						// Note: because of how the PE/COFF format is
						// structured, not all zero-initialized data is marked
						// as such. A portion may be at the end of the .data
						// section and is thus marked as initialized data.
						sections = append(sections, memorySection{
							Address: address + uint64(section.Size),
							Size:    uint64(section.VirtualSize) - uint64(section.Size),
							Type:    memoryBSS,
						})
					}
				} else if section.Characteristics&IMAGE_SCN_MEM_READ != 0 {
					// .rdata, .buildid, .pdata
					sections = append(sections, memorySection{
						Address: address,
						Size:    uint64(section.VirtualSize),
						Type:    memoryROData,
					})
				}
			}
		}
	} else if file, err := wasm.Parse(f); err == nil {
		// File is in WebAssembly format.

		// Put code at a very high address, so that it won't conflict with the
		// data in the memory section.
		const codeOffset = 0x8000_0000_0000_0000

		// Read DWARF information. The error is intentionally ignored.
		data, err := file.DWARF()
		if data != nil {
			addresses, err = readProgramSizeFromDWARF(data, codeOffset)
			if err != nil {
				// However, _do_ report an error here. Something must have gone
				// wrong while trying to parse DWARF data.
				return nil, err
			}
		}

		var linearMemorySize uint64
		for _, section := range file.Sections {
			switch section := section.(type) {
			case *wasm.SectionCode:
				sections = append(sections, memorySection{
					Address: codeOffset,
					Size:    uint64(section.Size()),
					Type:    memoryCode,
				})
			case *wasm.SectionMemory:
				// This value is used when processing *wasm.SectionData (which
				// always comes after *wasm.SectionMemory).
				linearMemorySize = uint64(section.Entries[0].Limits.Initial) * 64 * 1024
			case *wasm.SectionData:
				// Data sections contain initial values for linear memory.
				// First load the list of data sections, and sort them by
				// address for easier processing.
				var dataSections []memorySection
				for _, entry := range section.Entries {
					address, err := wasm.Eval(bytes.NewBuffer(entry.Offset))
					if err != nil {
						return nil, fmt.Errorf("could not parse data section address: %w", err)
					}
					dataSections = append(dataSections, memorySection{
						Address: uint64(address[0].(int32)),
						Size:    uint64(len(entry.Data)),
						Type:    memoryData,
					})
				}
				sort.Slice(dataSections, func(i, j int) bool {
					return dataSections[i].Address < dataSections[j].Address
				})

				// And now add all data sections for linear memory.
				// Parts that are in the slice of data sections are added as
				// memoryData, and parts that are not are added as memoryBSS.
				addr := uint64(0)
				for _, section := range dataSections {
					if addr < section.Address {
						sections = append(sections, memorySection{
							Address: addr,
							Size:    section.Address - addr,
							Type:    memoryBSS,
						})
					}
					if addr > section.Address {
						// This might be allowed, I'm not sure.
						// It certainly doesn't make a lot of sense.
						return nil, fmt.Errorf("overlapping data section")
					}
					// addr == section.Address
					sections = append(sections, section)
					addr = section.Address + section.Size
				}
				if addr < linearMemorySize {
					sections = append(sections, memorySection{
						Address: addr,
						Size:    linearMemorySize - addr,
						Type:    memoryBSS,
					})
				}
			}
		}
	} else {
		return nil, fmt.Errorf("could not parse file: %w", err)
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
	for _, section := range sections {
		switch section.Type {
		case memoryCode:
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				if isVariable {
					field.ROData += size
				} else {
					field.Code += size
				}
				sizes[path] = field
			}, packagePathMap)
		case memoryROData:
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				field.ROData += size
				sizes[path] = field
			}, packagePathMap)
		case memoryData:
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				field.Data += size
				sizes[path] = field
			}, packagePathMap)
		case memoryBSS:
			readSection(section, addresses, func(path string, size uint64, isVariable bool) {
				field := sizes[path]
				field.BSS += size
				sizes[path] = field
			}, packagePathMap)
		case memoryStack:
			// We store the C stack as a pseudo-package.
			sizes["C stack"] = packageSize{
				BSS: section.Size,
			}
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
func readSection(section memorySection, addresses []addressLine, addSize func(string, uint64, bool), packagePathMap map[string]string) {
	// The addr variable tracks at which address we are while going through this
	// section. We start at the beginning.
	addr := section.Address
	sectionEnd := section.Address + section.Size
	for _, line := range addresses {
		if line.Address < section.Address || line.Address+line.Length >= sectionEnd {
			// Check that this line is entirely within the section.
			// Don't bother dealing with line entries that cross sections (that
			// seems rather unlikely anyway).
			continue
		}
		if addr < line.Address {
			// There is a gap: there is a space between the current and the
			// previous line entry.
			addSize("(unknown)", line.Address-addr, false)
			if sizesDebug {
				fmt.Printf("%08x..%08x %4d: unknown (gap)\n", addr, line.Address, line.Address-addr)
			}
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
		if sizesDebug {
			fmt.Printf("%08x..%08x %4d: unknown (end)\n", addr, sectionEnd, sectionEnd-addr)
		}
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
		} else if path == "<Go interface assert>" {
			// Interface type assert, generated by the interface lowering pass.
			packagePath = "Go interface assert"
		} else if path == "<Go interface method>" {
			// Interface method wrapper (switch over all concrete types),
			// generated by the interface lowering pass.
			packagePath = "Go interface method"
		} else if path == "<stdin>" {
			// This can happen when the source code (in Go) doesn't have a
			// source file and uses "-" as the location. Somewhere this is
			// converted to "<stdin>".
			// Convert this back to the "-" string. Eventually, this should be
			// fixed in the compiler.
			packagePath = "-"
		} else {
			// This is some other path. Not sure what it is, so just emit its directory.
			packagePath = filepath.Dir(path) // fallback
		}
	}
	return packagePath
}
