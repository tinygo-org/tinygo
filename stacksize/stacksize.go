// Package stacksize tries to determine the call graph for ELF binaries and
// tries to parse stack size information from DWARF call frame information.
package stacksize

import (
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
)

// set to true to print information useful for debugging
const debugPrint = false

// SizeType indicates whether a stack or frame size could be determined and if
// not, why.
type SizeType uint8

// Results after trying to determine the stack size of a function in the call
// graph. The goal is to find a maximum (bounded) stack size, but sometimes this
// is not possible for some reasons such as recursion or indirect calls.
const (
	Undefined SizeType = iota // not yet calculated
	Unknown                   // child has unknown stack size
	Bounded                   // stack size is fixed at compile time (no recursion etc)
	Recursive
	IndirectCall
)

func (s SizeType) String() string {
	switch s {
	case Undefined:
		return "undefined"
	case Unknown:
		return "unknown"
	case Bounded:
		return "bounded"
	case Recursive:
		return "recursive"
	case IndirectCall:
		return "indirect call"
	default:
		return "<?>"
	}
}

// CallNode is a node in the call graph (that is, a function). Because this is
// determined after linking, there may be multiple names for a single function
// (due to aliases). It is also possible multiple functions have the same name
// (but are in fact different), for example for static functions in C.
type CallNode struct {
	Names            []string
	Address          uint64      // address at which the function is linked (without T bit on ARM)
	Size             uint64      // symbol size, in bytes
	Children         []*CallNode // functions this function calls
	FrameSize        uint64      // frame size, if FrameSizeType is Bounded
	FrameSizeType    SizeType    // can be Undefined or Bounded
	stackSize        uint64
	stackSizeType    SizeType
	missingFrameInfo *CallNode // the child function that is the cause for not being able to determine the stack size
}

func (n *CallNode) String() string {
	if n == nil {
		return "<nil>"
	}
	return n.Names[0]
}

// CallGraph parses the ELF file and reads DWARF call frame information to
// determine frame sizes for each function, as far as that's possible. Because
// at this point it is not possible to determine indirect calls, a list of
// indirect function calling functions needs to be supplied separately.
//
// This function does not attempt to determine the stack size for functions.
// This is done by calling StackSize on a function in the call graph.
func CallGraph(f *elf.File, callsIndirectFunction []string) (map[string][]*CallNode, error) {
	// Sanity check that there is exactly one symbol table.
	// Multiple symbol tables are possible, but aren't yet supported below.
	numSymbolTables := 0
	for _, section := range f.Sections {
		if section.Type == elf.SHT_SYMTAB {
			numSymbolTables++
		}
	}
	if numSymbolTables != 1 {
		return nil, fmt.Errorf("expected exactly one symbol table, got %d", numSymbolTables)
	}

	// Collect all symbols in the executable.
	symbols := make(map[uint64]*CallNode)
	symbolList := make([]*CallNode, 0)
	symbolNames := make(map[string][]*CallNode)
	elfSymbols, err := f.Symbols()
	if err != nil {
		return nil, err
	}
	for _, elfSymbol := range elfSymbols {
		if elf.ST_TYPE(elfSymbol.Info) != elf.STT_FUNC {
			continue
		}
		address := elfSymbol.Value
		if f.Machine == elf.EM_ARM {
			address = address &^ 1
		}
		var node *CallNode
		if n, ok := symbols[address]; ok {
			// Existing symbol.
			if n.Size != elfSymbol.Size {
				return nil, fmt.Errorf("symbol at 0x%x has inconsistent size (%d for %s and %d for %s)", address, n.Size, n.Names[0], elfSymbol.Size, elfSymbol.Name)
			}
			node = n
			node.Names = append(node.Names, elfSymbol.Name)
		} else {
			// New symbol.
			node = &CallNode{
				Names:   []string{elfSymbol.Name},
				Address: address,
				Size:    elfSymbol.Size,
			}
			symbols[address] = node
			symbolList = append(symbolList, node)
		}
		symbolNames[elfSymbol.Name] = append(symbolNames[elfSymbol.Name], node)
	}

	// Sort symbols by address, for binary searching.
	sort.Slice(symbolList, func(i, j int) bool {
		return symbolList[i].Address < symbolList[j].Address
	})

	// Load relocations and construct the call graph.
	for _, section := range f.Sections {
		if section.Type != elf.SHT_REL {
			continue
		}
		if section.Entsize != 8 {
			// Assume ELF32, this should be fixed.
			return nil, fmt.Errorf("only ELF32 is supported at this time")
		}
		data, err := section.Data()
		if err != nil {
			return nil, err
		}
		for i := uint64(0); i < section.Size/section.Entsize; i++ {
			offset := binary.LittleEndian.Uint32(data[i*section.Entsize:])
			info := binary.LittleEndian.Uint32(data[i*section.Entsize+4:])
			if elf.R_SYM32(info) == 0 {
				continue
			}
			elfSymbol := elfSymbols[elf.R_SYM32(info)-1]
			if elf.ST_TYPE(elfSymbol.Info) != elf.STT_FUNC {
				continue
			}
			address := elfSymbol.Value
			if f.Machine == elf.EM_ARM {
				address = address &^ 1
			}
			childSym := symbols[address]
			switch f.Machine {
			case elf.EM_ARM:
				relocType := elf.R_ARM(elf.R_TYPE32(info))
				parentSym := findSymbol(symbolList, uint64(offset))
				if debugPrint {
					fmt.Fprintf(os.Stderr, "found relocation %-24s at %s (0x%x) to %s (0x%x)\n", relocType, parentSym, offset, childSym, childSym.Address)
				}
				isCall := true
				switch relocType {
				case elf.R_ARM_THM_PC22: // actually R_ARM_THM_CALL
					// used for bl calls
				case elf.R_ARM_THM_JUMP24:
					// used for b.w jumps
					isCall = parentSym != childSym
				case elf.R_ARM_THM_JUMP11:
					// used for b.n jumps
					isCall = parentSym != childSym
				case elf.R_ARM_THM_MOVW_ABS_NC, elf.R_ARM_THM_MOVT_ABS:
					// used for getting a function pointer
					isCall = false
				case elf.R_ARM_ABS32:
					// when compiling with -Oz (minsize), used for calling
					isCall = true
				default:
					return nil, fmt.Errorf("unknown relocation: %s", relocType)
				}
				if isCall {
					if parentSym != nil {
						parentSym.Children = append(parentSym.Children, childSym)
					}
				}
			default:
				return nil, fmt.Errorf("unknown architecture: %s", f.Machine)
			}
		}
	}

	// Set fixed frame size information, depending on the architecture.
	switch f.Machine {
	case elf.EM_ARM:
		knownFrameSizes := map[string]uint64{
			// implemented with assembly in compiler-rt
			"__aeabi_idivmod":  3 * 4, // 3 registers on thumb1 but 1 register on thumb2
			"__aeabi_uidivmod": 3 * 4, // 3 registers on thumb1 but 1 register on thumb2
			"__aeabi_ldivmod":  2 * 4,
			"__aeabi_uldivmod": 2 * 4,
			"__aeabi_memclr":   2 * 4, // 2 registers on thumb1
			"__aeabi_memset":   2 * 4, // 2 registers on thumb1
			"__aeabi_memcmp":   2 * 4, // 2 registers on thumb1
			"__aeabi_memcpy":   2 * 4, // 2 registers on thumb1
			"__aeabi_memmove":  2 * 4, // 2 registers on thumb1
			"__aeabi_dcmpeq":   2 * 4,
			"__aeabi_dcmplt":   2 * 4,
			"__aeabi_dcmple":   2 * 4,
			"__aeabi_dcmpge":   2 * 4,
			"__aeabi_dcmpgt":   2 * 4,
			"__aeabi_fcmpeq":   2 * 4,
			"__aeabi_fcmplt":   2 * 4,
			"__aeabi_fcmple":   2 * 4,
			"__aeabi_fcmpge":   2 * 4,
			"__aeabi_fcmpgt":   2 * 4,
		}
		for name, size := range knownFrameSizes {
			if sym, ok := symbolNames[name]; ok {
				if len(sym) > 1 {
					return nil, fmt.Errorf("expected zero or one occurence of the symbol %s, found %d", name, len(sym))
				}
				sym[0].FrameSize = size
				sym[0].FrameSizeType = Bounded
			}
		}
	}

	// Mark functions that do indirect calls (which cannot be determined
	// directly from ELF/DWARF information).
	for _, name := range callsIndirectFunction {
		for _, fn := range symbolNames[name] {
			fn.stackSizeType = IndirectCall
			fn.missingFrameInfo = fn
		}
	}

	// Read the .debug_frame section.
	section := f.Section(".debug_frame")
	if section == nil {
		return nil, errors.New("no .debug_frame section present, binary was compiled without debug information")
	}
	data, err := section.Data()
	if err != nil {
		return nil, fmt.Errorf("could not read .debug_frame section: %w", err)
	}
	err = parseFrames(f, data, symbols)
	if err != nil {
		return nil, err
	}

	return symbolNames, nil
}

// findSymbol determines in which symbol the given address lies.
func findSymbol(symbolList []*CallNode, address uint64) *CallNode {
	// TODO: binary search
	for _, sym := range symbolList {
		if address >= sym.Address && address < sym.Address+sym.Size {
			return sym
		}
	}
	return nil
}

// StackSize tries to determine the stack size of the given call graph node. It
// returns the maximum stack size, whether this size can be known at compile
// time and the call node responsible for failing to determine the maximum stack
// usage. The stack size is only valid if sizeType is Bounded.
func (node *CallNode) StackSize() (uint64, SizeType, *CallNode) {
	if node.stackSizeType == Undefined {
		node.determineStackSize(make(map[*CallNode]struct{}))
	}
	return node.stackSize, node.stackSizeType, node.missingFrameInfo
}

// determineStackSize tries to determine the maximum stack size for this
// function, recursively.
func (node *CallNode) determineStackSize(parents map[*CallNode]struct{}) {
	if _, ok := parents[node]; ok {
		// The function calls itself (directly or indirectly).
		node.stackSizeType = Recursive
		node.missingFrameInfo = node
		return
	}
	parents[node] = struct{}{}
	defer func() {
		delete(parents, node)
	}()
	switch node.FrameSizeType {
	case Bounded:
		// Determine the stack size recursively.
		childMaxStackSize := uint64(0)
		for _, child := range node.Children {
			if child.stackSizeType == Undefined {
				child.determineStackSize(parents)
			}
			switch child.stackSizeType {
			case Bounded:
				if child.stackSize > childMaxStackSize {
					childMaxStackSize = child.stackSize
				}
			case Unknown, Recursive, IndirectCall:
				node.stackSizeType = child.stackSizeType
				node.missingFrameInfo = child.missingFrameInfo
				return
			default:
				panic("unknown child stack size type")
			}
		}
		node.stackSize = node.FrameSize + childMaxStackSize
		node.stackSizeType = Bounded
	case Undefined:
		node.stackSizeType = Unknown
		node.missingFrameInfo = node
	default:
		panic("unknown frame size type") // unreachable
	}
}
