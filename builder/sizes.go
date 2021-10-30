package builder

import (
	"debug/elf"
	"sort"
	"strings"
)

// programSize contains size statistics per package of a compiled program.
type programSize struct {
	Packages map[string]*packageSize
	Sum      *packageSize
	Code     uint64
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

type symbolList []elf.Symbol

func (l symbolList) Len() int {
	return len(l)
}

func (l symbolList) Less(i, j int) bool {
	bind_i := elf.ST_BIND(l[i].Info)
	bind_j := elf.ST_BIND(l[j].Info)
	if l[i].Value == l[j].Value && bind_i != elf.STB_WEAK && bind_j == elf.STB_WEAK {
		// sort weak symbols after non-weak symbols
		return true
	}
	return l[i].Value < l[j].Value
}

func (l symbolList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// loadProgramSize calculate a program/data size breakdown of each package for a
// given ELF file.
func loadProgramSize(path string) (*programSize, error) {
	file, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sumCode uint64
	var sumData uint64
	var sumBSS uint64
	for _, section := range file.Sections {
		if section.Flags&elf.SHF_ALLOC == 0 {
			continue
		}
		if section.Type != elf.SHT_PROGBITS && section.Type != elf.SHT_NOBITS {
			continue
		}
		if section.Type == elf.SHT_NOBITS {
			sumBSS += section.Size
		} else if section.Flags&elf.SHF_EXECINSTR != 0 {
			sumCode += section.Size
		} else if section.Flags&elf.SHF_WRITE != 0 {
			sumData += section.Size
		}
	}

	allSymbols, err := file.Symbols()
	if err != nil {
		return nil, err
	}
	symbols := make([]elf.Symbol, 0, len(allSymbols))
	for _, symbol := range allSymbols {
		symType := elf.ST_TYPE(symbol.Info)
		if symbol.Size == 0 {
			continue
		}
		if symType != elf.STT_FUNC && symType != elf.STT_OBJECT && symType != elf.STT_NOTYPE {
			continue
		}
		if symbol.Section >= elf.SectionIndex(len(file.Sections)) {
			continue
		}
		section := file.Sections[symbol.Section]
		if section.Flags&elf.SHF_ALLOC == 0 {
			continue
		}
		symbols = append(symbols, symbol)
	}
	sort.Sort(symbolList(symbols))

	sizes := map[string]*packageSize{}
	var lastSymbolValue uint64
	for _, symbol := range symbols {
		symType := elf.ST_TYPE(symbol.Info)
		//bind := elf.ST_BIND(symbol.Info)
		section := file.Sections[symbol.Section]
		pkgName := "(bootstrap)"
		symName := strings.TrimLeft(symbol.Name, "(*")
		dot := strings.IndexByte(symName, '.')
		if dot > 0 {
			pkgName = symName[:dot]
		}
		pkgSize := sizes[pkgName]
		if pkgSize == nil {
			pkgSize = &packageSize{}
			sizes[pkgName] = pkgSize
		}
		if lastSymbolValue != symbol.Value || lastSymbolValue == 0 {
			if symType == elf.STT_FUNC {
				pkgSize.Code += symbol.Size
			} else if section.Flags&elf.SHF_WRITE != 0 {
				if section.Type == elf.SHT_NOBITS {
					pkgSize.BSS += symbol.Size
				} else {
					pkgSize.Data += symbol.Size
				}
			} else {
				pkgSize.ROData += symbol.Size
			}
		}
		lastSymbolValue = symbol.Value
	}

	sum := &packageSize{}
	for _, pkg := range sizes {
		sum.Code += pkg.Code
		sum.ROData += pkg.ROData
		sum.Data += pkg.Data
		sum.BSS += pkg.BSS
	}

	return &programSize{Packages: sizes, Code: sumCode, Data: sumData, BSS: sumBSS, Sum: sum}, nil
}
