package goobj

// Parse the Go object file format.
// Some of the code here has been copied from the following file:
// https://github.com/golang/go/blob/master/src/cmd/internal/goobj/objfile.go
// It is licensed under the same license as the whole Go source tree:
//
//   Copyright 2019 The Go Authors. All rights reserved.
//   Use of this source code is governed by a BSD-style
//   license that can be found in the LICENSE file.
//
// For documentation on the file format, see the objfile.go file (linked above).

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// see objfile.go
const (
	stringRefSize = 8 // two uint32s
	symbolSize    = stringRefSize + 2 + 1 + 1 + 1 + 4 + 4
	auxSize       = 1 + 8
)

// Blocks
const (
	blkAutolib = iota
	blkPkgIdx
	blkFile
	blkSymdef
	blkHashed64def
	blkHasheddef
	blkNonpkgdef
	blkNonpkgref
	blkRefFlags
	blkHash64
	blkHash
	blkRelocIdx
	blkAuxIdx
	blkDataIdx
	blkReloc
	blkAux
	blkData
	blkPcdata // note: gone in Go 1.17
	blkRefName
	blkEnd
	numBlk
)

type symKind uint8

// See: https://github.com/golang/go/blob/master/src/cmd/internal/objabi/symkind.go
// This is only the first few symbols that should be more or less stable.
const (
	// Executable instructions
	symbolTEXT symKind = iota + 1
	// Read only static data
	symbolRODATA
	// Static data that does not contain any pointers
	symbolNOPTRDATA
	// Static data
	symbolDATA
	// Statically data that is initially all 0s
	symbolBSS
	// Statically data that is initially all 0s and does not contain pointers
	symbolNOPTRBSS
)

type relocType uint16

// Relocation constants. Not the same values as in reloctype.go, because those
// can change each release.
// See: https://github.com/golang/go/blob/master/src/cmd/internal/objabi/reloctype.go
const (
	r_ADDRARM64 relocType = iota + 1
	r_CALL
	r_CALLARM64
	r_CALLIND
	r_PCREL
)

// Package Index.
const (
	pkgIdxNone     = (1<<31 - 1) - iota // Non-package symbols
	pkgIdxHashed64                      // Short hashed (content-addressable) symbols
	pkgIdxHashed                        // Hashed (content-addressable) symbols
	pkgIdxBuiltin                       // Predefined runtime symbols (ex: runtime.newobject)
	pkgIdxSelf                          // Symbols defined in the current package
	pkgIdxInvalid  = 0
	// The index of other referenced packages starts from 1.
)

type stringRef struct {
	Len    uint32
	Offset uint32
}

type symbolDef struct {
	Index  uint32
	Name   string
	Type   symKind
	Size   uint32
	Align  uint32
	Data   []byte
	Relocs []relocation
}

type relocation struct {
	Offset uint32
	Size   uint8
	Type   relocType
	Addend int64
	Symbol *symbolDef
}

// GoObj is a Go object file loaded into memory.
type GoObj struct {
	goos        string
	goarch      string
	goversion   string
	magic       string
	offsets     [numBlk]uint32
	buf         []byte
	symdef      []*symbolDef
	hashed64def []*symbolDef
	hasheddef   []*symbolDef
	nonpkgdef   []*symbolDef
	nonpkgref   []*symbolDef
}

// SupportsTarget returns whether the given GOOS/GOARCH pair is supported by the
// goobj package.
func SupportsTarget(goos, goarch string) bool {
	switch goos {
	case "linux":
		switch goarch {
		case "amd64", "arm64":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// ReadGoObj reads the given Go object file into memory.
func ReadGoObj(b []byte) (*GoObj, error) {
	obj := &GoObj{}
	r := bytes.NewBuffer(b)

	// Read first line. Format: "go object linux arm64 go1.16.2"
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) < 5 {
		return nil, fmt.Errorf("expected file to have a line with at least 5 fields, got %#v (%d)", line, len(parts))
	}
	obj.goos = parts[2]
	obj.goarch = parts[3]
	obj.goversion = parts[4]
	// Read and ignore second line. Format: "!"
	_, err = r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	obj.buf = r.Bytes()
	r = bytes.NewBuffer(obj.buf)
	obj.magic, err = r.ReadString('d')
	if err != nil {
		return nil, err
	}
	if obj.magic != "\x00go116ld" && obj.magic != "\x00go117ld" {
		return nil, fmt.Errorf("unexpected magic: %#v", obj.magic)
	}
	r.Next(8) // skip the fingerprint
	r.Next(4) // skip the flags
	err = binary.Read(r, binary.LittleEndian, &obj.offsets)
	if err != nil {
		return nil, err
	}

	// Read all symbols.
	obj.symdef = obj.readSymbols(blkSymdef, 0)
	symbolIndex := uint32(len(obj.symdef))
	obj.hashed64def = obj.readSymbols(blkHashed64def, symbolIndex)
	symbolIndex += uint32(len(obj.hashed64def))
	obj.hasheddef = obj.readSymbols(blkHasheddef, symbolIndex)
	symbolIndex += uint32(len(obj.hasheddef))
	obj.nonpkgdef = obj.readSymbols(blkNonpkgdef, symbolIndex)
	symbolIndex += uint32(len(obj.nonpkgdef))
	obj.nonpkgref = obj.readSymbols(blkNonpkgref, symbolIndex)

	for _, symbol := range obj.nonpkgdef {
		obj.readRelocs(symbol)
	}

	return obj, nil
}

func (obj *GoObj) readNumSymbols(block uint32) uint32 {
	return (obj.offsets[block+1] - obj.offsets[block]) / symbolSize
}

func (obj *GoObj) readSymbols(block, symbolOffset uint32) []*symbolDef {
	n := obj.readNumSymbols(block)
	r := bytes.NewReader(obj.buf)
	r.Seek(int64(obj.offsets[block]), io.SeekStart)
	refs := make([]struct {
		Name  stringRef
		ABI   uint16
		Type  symKind
		Flag  uint8
		Flag2 uint8
		Siz   uint32
		Align uint32
	}, n)
	binary.Read(r, binary.LittleEndian, refs)
	symbols := make([]*symbolDef, n)
	for i, ref := range refs {
		name := obj.readString(ref.Name.Len, ref.Name.Offset)
		var data []byte
		if block != blkNonpkgref {
			data = obj.readData(symbolOffset + uint32(i))
		}
		symbols[i] = &symbolDef{
			Index: symbolOffset + uint32(i),
			Name:  name,
			Type:  ref.Type,
			Size:  ref.Siz,
			Align: ref.Align,
			Data:  data,
		}
	}
	return symbols
}

func (obj *GoObj) readData(index uint32) []byte {
	dataIdxOff := obj.offsets[blkDataIdx] + index*4
	base := obj.offsets[blkData]
	off := binary.LittleEndian.Uint32(obj.buf[dataIdxOff:])
	end := binary.LittleEndian.Uint32(obj.buf[dataIdxOff+4:])
	return obj.buf[base+off : base+end]
}

func (obj *GoObj) readRelocs(symbol *symbolDef) {
	relocIdxOff := obj.offsets[blkRelocIdx] + symbol.Index*4
	off := binary.LittleEndian.Uint32(obj.buf[relocIdxOff:])
	end := binary.LittleEndian.Uint32(obj.buf[relocIdxOff+4:])
	for i := off; i < end; i++ {
		offset := obj.offsets[blkReloc] + i*obj.relocSize()
		relocHeader := obj.buf[offset : offset+obj.relocSize()]
		relocOffset := binary.LittleEndian.Uint32(relocHeader[0:])
		relocSize := relocHeader[4]
		var relocType uint16
		var relocAddend int64
		var relocPkgIdx, relocSymIdx uint32
		if obj.magic <= "\x00go116ld" {
			relocType = uint16(relocHeader[5])
			relocAddend = int64(binary.LittleEndian.Uint64(relocHeader[6:]))
			relocPkgIdx = binary.LittleEndian.Uint32(relocHeader[14:])
			relocSymIdx = binary.LittleEndian.Uint32(relocHeader[18:])
		} else {
			relocType = binary.LittleEndian.Uint16(relocHeader[5:])
			relocAddend = int64(binary.LittleEndian.Uint64(relocHeader[7:]))
			relocPkgIdx = binary.LittleEndian.Uint32(relocHeader[15:])
			relocSymIdx = binary.LittleEndian.Uint32(relocHeader[19:])
		}
		var relocSymbol *symbolDef
		switch relocPkgIdx {
		case pkgIdxHashed64:
			relocSymbol = obj.hashed64def[relocSymIdx]
		case pkgIdxHashed:
			relocSymbol = obj.hasheddef[relocSymIdx]
		case pkgIdxNone:
			if relocSymIdx < uint32(len(obj.nonpkgdef)) {
				relocSymbol = obj.nonpkgdef[relocSymIdx]
			} else {
				relocSymbol = obj.nonpkgref[relocSymIdx-uint32(len(obj.nonpkgdef))]
			}
		case pkgIdxSelf:
			relocSymbol = obj.symdef[relocSymIdx]
		case 0:
			if relocSymIdx != 0 {
				// According to objapi.go:
				// > {0, 0} represents a nil symbol. Otherwise PkgIdx should not
				// > be 0.
				panic("expected symbol index to be 0 when package index is 0")
			}
			relocSymbol = nil
		default:
			panic("unknown symbol reference")
		}
		symbol.Relocs = append(symbol.Relocs, relocation{
			Offset: relocOffset,
			Size:   relocSize,
			Type:   obj.relocationFor(relocType),
			Addend: relocAddend,
			Symbol: relocSymbol,
		})
	}
}

func (obj *GoObj) readString(len, off uint32) string {
	return string(obj.buf[off : off+len])
}

// findExportedSymbols returns the list of symbols that should be exported from
// the object file, such as defined assembly functions.
func (obj *GoObj) findExportedSymbols() []*symbolDef {
	var symbols []*symbolDef
	for _, symbol := range obj.nonpkgdef {
		switch symbol.Type {
		case symbolTEXT:
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

// findLocalSymbols returns a list of all local symbols that should be included
// in the resulting object file.
func (obj *GoObj) findLocalSymbols() []*symbolDef {
	var symbols []*symbolDef
	for _, symbol := range obj.symdef {
		if symbol.Type == symbolRODATA {
			if symbol.Name == "" {
				continue
			}
			symbols = append(symbols, symbol)
		}
	}
	for _, symbol := range obj.hashed64def {
		if symbol.Type == symbolRODATA {
			if symbol.Name == "" {
				continue
			}
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

// findExported returns a list of symbols referenced from local or exported but
// not defined in one of those.
func (obj *GoObj) findExternal(local, exported []*symbolDef) []*symbolDef {
	// Get a list of all referenced symbols (including defined symbols).
	relocSymbolSet := make(map[*symbolDef]struct{})
	for _, symbols := range [][]*symbolDef{local, exported} {
		for _, symbol := range symbols {
			for _, reloc := range symbol.Relocs {
				if reloc.Type == r_CALLIND {
					continue
				}
				if _, ok := relocSymbolSet[reloc.Symbol]; !ok {
					relocSymbolSet[reloc.Symbol] = struct{}{}
				}
			}
		}
	}

	// Go through all non-defined references and if they are one of the referend
	// symbols above, add them to the list.
	// This avoids including lots of external symbols that aren't referenced from anywhere.
	var relocs []*symbolDef
	for _, symbol := range obj.nonpkgref {
		if symbol.Type == 0 { // external symbol
			if symbol.Name == "" {
				continue // shouldn't happen
			}
			if _, ok := relocSymbolSet[symbol]; ok {
				relocs = append(relocs, symbol)
			}
		}
	}
	return relocs
}

// References returns the imported and exported references in this object file.
func (obj *GoObj) References() map[string]string {
	// Create a few lists of symbols.
	exported := obj.findExportedSymbols()
	local := obj.findLocalSymbols()
	referenced := obj.findExternal(exported, local)

	// Add both exported and referend symbols to the same map.
	references := make(map[string]string)
	for _, symbol := range exported {
		references[symbol.Name] = mangleName(symbol.Name)
	}
	for _, symbol := range referenced {
		references[symbol.Name] = mangleName(symbol.Name)
	}

	return references
}

func (obj *GoObj) relocationFor(reloc uint16) relocType {
	switch obj.magic {
	case "\x00go116ld":
		// See: https://github.com/golang/go/blob/release-branch.go1.16/src/cmd/internal/objabi/reloctype.go
		return [...]relocType{
			3:  r_ADDRARM64,
			8:  r_CALL,
			10: r_CALLARM64,
			11: r_CALLIND,
			16: r_PCREL,
		}[reloc]
	case "\x00go117ld":
		// See: https://github.com/golang/go/blob/release-branch.go1.17/src/cmd/internal/objabi/reloctype.go
		return [...]relocType{
			3:  r_ADDRARM64,
			7:  r_CALL,
			9:  r_CALLARM64,
			10: r_CALLIND,
			15: r_PCREL,
		}[reloc]
	default:
		panic("unkown magic: " + obj.magic)
	}
}

func (obj *GoObj) relocSize() uint32 {
	switch obj.magic {
	case "\x00go116ld":
		return 4 + 1 + 1 + 8 + 8
	case "\x00go117ld":
		return 4 + 1 + 2 + 8 + 8
	default:
		panic("unknown magic")
	}
}

// Return mangled name for the Go ABI0 ABI. For example, convert math.Sqrt to
// __GoABI0_math.Sqrt.
func mangleName(name string) string {
	return "__GoABI0_" + name
}
