package main

import (
	"debug/dwarf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
)

// wasmFile is a core module or a component. wasmtime's address map uses file
// offsets for both, so everything here does too.
type wasmFile struct {
	data       []byte
	modules    []*wasmModule // sorted by start
	bodyStarts map[int64]bool
}

// wasmModule holds the parts of a core module needed for symbolization.
type wasmModule struct {
	start, end int64 // file offsets of the module bytes
	// File offset of the first byte of the code section payload. DWARF
	// addresses in wasm are offsets relative to this.
	codeStart int64
	custom    map[string][]byte
	data      []byte // the whole file
	dw        *dwarfInfo
}

func readWasm(path string) (*wasmFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f := &wasmFile{data: data, bodyStarts: map[int64]bool{}}
	if err := f.parse(0, int64(len(data))); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(f.modules) == 0 {
		return nil, fmt.Errorf("%s: no core modules with code", path)
	}
	return f, nil
}

// parse reads the module or component in data[start:end], recording core
// modules (including those nested in components).
func (f *wasmFile) parse(start, end int64) error {
	data := f.data[:end]
	if end-start < 8 || string(data[start:start+4]) != "\x00asm" {
		return fmt.Errorf("not a wasm module or component at %#x", start)
	}
	component := false
	switch binary.LittleEndian.Uint32(data[start+4:]) {
	case 1:
	case 0x0001000d: // version 0xd, layer 1
		component = true
	default:
		return fmt.Errorf("unknown wasm version at %#x", start)
	}
	m := &wasmModule{start: start, end: end, codeStart: -1, custom: map[string][]byte{}, data: f.data}
	pos := start + 8
	for pos < end {
		id := data[pos]
		pos++
		size, n := binary.Uvarint(data[pos:])
		if n <= 0 || pos+int64(n)+int64(size) > end {
			return fmt.Errorf("malformed section header at %#x", pos)
		}
		pos += int64(n)
		body := data[pos : pos+int64(size)]
		switch {
		case component && (id == 1 || id == 4): // core module, component
			if err := f.parse(pos, pos+int64(size)); err != nil {
				return err
			}
		case !component && id == 0:
			nameLen, k := binary.Uvarint(body)
			if k <= 0 || k+int(nameLen) > len(body) {
				return fmt.Errorf("malformed custom section at %#x", pos)
			}
			m.custom[string(body[k:k+int(nameLen)])] = body[k+int(nameLen):]
		case !component && id == 10:
			m.codeStart = pos
			if err := f.readBodies(pos, body); err != nil {
				return err
			}
		}
		pos += int64(size)
	}
	if !component && m.codeStart >= 0 {
		if m.hasDWARF() {
			dw, err := newDWARFInfo(m)
			if err != nil {
				return fmt.Errorf("reading DWARF of module at %#x: %w", start, err)
			}
			m.dw = dw
		}
		i := sort.Search(len(f.modules), func(i int) bool { return f.modules[i].start > start })
		f.modules = slices.Insert(f.modules, i, m)
	}
	return nil
}

// readBodies records the file offset of each function body in a code section,
// which starts after the body size.
func (f *wasmFile) readBodies(pos int64, code []byte) error {
	count, n := binary.Uvarint(code)
	if n <= 0 {
		return fmt.Errorf("malformed code section at %#x", pos)
	}
	off := int64(n)
	for range count {
		size, n := binary.Uvarint(code[off:])
		if n <= 0 || off+int64(n)+int64(size) > int64(len(code)) {
			return fmt.Errorf("malformed function body at %#x", pos+off)
		}
		off += int64(n)
		f.bodyStarts[pos+off] = true
		off += int64(size)
	}
	return nil
}

// moduleFor returns the core module containing a file offset.
func (f *wasmFile) moduleFor(off int64) *wasmModule {
	i := sort.Search(len(f.modules), func(i int) bool { return f.modules[i].start > off }) - 1
	if i < 0 || off >= f.modules[i].end {
		return nil
	}
	return f.modules[i]
}

// sampleOffset moves a sample at a loop instruction (where wasmtime puts epoch
// checks) to the first instruction of the loop body, as the loop has no line.
func (m *wasmModule) sampleOffset(off int64) int64 {
	if off < 0 || off+1 >= int64(len(m.data)) || m.data[off] != 0x03 {
		return off
	}
	// The block type is a signed LEB128 value. Only its length is used.
	// https://webassembly.github.io/spec/core/binary/instructions.html#control-instructions
	_, n := binary.Varint(m.data[off+1:])
	if n <= 0 {
		return off
	}
	return off + 1 + int64(n)
}

func (m *wasmModule) hasDWARF() bool {
	return m.custom[".debug_info"] != nil && m.custom[".debug_line"] != nil
}

// A source position for one level of a (possibly inlined) call stack.
type frame struct {
	function string
	file     string // file of the function's declaration
	declLine int64  // line of the function's declaration
	line     int64  // line being executed in this function
	srcFile  string // file containing line
	guessed  bool   // function guessed from declarations (missing inline scope)
	prologue bool   // line is the declaration line, as there is no line row
}

// dwarfInfo answers "which (inlined) functions and lines are at this code
// offset" from the module's DWARF.
type dwarfInfo struct {
	data  *dwarf.Data
	rows  []lineRow
	progs []*scope // top-level subprograms, sorted by first range start
	cus   []cuInfo // sorted by offset
	// Every named function declaration per file, sorted by line, to guess the
	// function for lines whose inline scope LLVM dropped.
	decls map[string][]decl
}

type decl struct {
	line int64
	name string
}

type cuInfo struct {
	offset dwarf.Offset
	files  []*dwarf.LineFile
}

type lineRow struct {
	addr   uint64
	file   string
	line   int64
	endSeq bool
}

type scope struct {
	ranges   [][2]uint64
	name     string
	declFile string
	declLine int64
	callFile string // where an inlined subroutine is called from
	callLine int64
	children []*scope
}

func (s *scope) contains(pc uint64) bool {
	for _, r := range s.ranges {
		if pc >= r[0] && pc < r[1] {
			return true
		}
	}
	return false
}

func newDWARFInfo(m *wasmModule) (*dwarfInfo, error) {
	d, err := dwarf.New(m.custom[".debug_abbrev"], m.custom[".debug_aranges"], nil,
		m.custom[".debug_info"], m.custom[".debug_line"], nil,
		m.custom[".debug_ranges"], m.custom[".debug_str"])
	if err != nil {
		return nil, err
	}
	for _, name := range []string{".debug_addr", ".debug_line_str", ".debug_str_offsets", ".debug_rnglists"} {
		if sec := m.custom[name]; sec != nil {
			if err := d.AddSection(name, sec); err != nil {
				return nil, err
			}
		}
	}
	di := &dwarfInfo{data: d, decls: map[string][]decl{}}
	if err := di.load(); err != nil {
		return nil, err
	}
	return di, nil
}

func (di *dwarfInfo) load() error {
	// Read all file tables first, as abstract origins can refer to later CUs
	// and their decl_file indexes into that CU's file table.
	r := di.data.Reader()
	for {
		cu, err := r.Next()
		if err != nil {
			return err
		}
		if cu == nil {
			break
		}
		r.SkipChildren()
		if cu.Tag != dwarf.TagCompileUnit {
			continue
		}
		info := cuInfo{offset: cu.Offset}
		lr, err := di.data.LineReader(cu)
		if err != nil {
			return err
		}
		if lr != nil {
			info.files = lr.Files()
			var le dwarf.LineEntry
			var seq []lineRow
			for {
				if err := lr.Next(&le); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}
				row := lineRow{addr: le.Address, line: int64(le.Line), endSeq: le.EndSequence}
				if le.File != nil {
					row.file = le.File.Name
				}
				seq = append(seq, row)
				if le.EndSequence {
					// Drop sequences for code the linker removed.
					if !isTombstone(seq[0].addr) {
						di.rows = append(di.rows, seq...)
					}
					seq = seq[:0]
				}
			}
		}
		di.cus = append(di.cus, info)
	}

	r = di.data.Reader()
	for {
		cu, err := r.Next()
		if err != nil {
			return err
		}
		if cu == nil {
			break
		}
		if cu.Tag != dwarf.TagCompileUnit || !cu.Children {
			r.SkipChildren()
			continue
		}
		if err := di.loadChildren(r, di.filesFor(cu.Offset), nil); err != nil {
			return err
		}
	}
	// End-of-sequence rows sort before a sequence starting at the same
	// address, so a lookup at that address finds the new sequence.
	sort.SliceStable(di.rows, func(i, j int) bool {
		a, b := di.rows[i], di.rows[j]
		if a.addr != b.addr {
			return a.addr < b.addr
		}
		return a.endSeq && !b.endSeq
	})
	sort.Slice(di.progs, func(i, j int) bool { return di.progs[i].ranges[0][0] < di.progs[j].ranges[0][0] })
	for f, ds := range di.decls {
		sort.Slice(ds, func(i, j int) bool { return ds[i].line < ds[j].line })
		di.decls[f] = slices.CompactFunc(ds, func(a, b decl) bool { return a.line == b.line })
	}
	return nil
}

// loadChildren reads subprograms with code as top-level scopes and inlined
// subroutines as their children, looking through lexical blocks.
func (di *dwarfInfo) loadChildren(r *dwarf.Reader, files []*dwarf.LineFile, parent *scope) error {
	for {
		e, err := r.Next()
		if err != nil {
			return err
		}
		if e == nil || e.Tag == 0 {
			return nil
		}
		if e.Tag == dwarf.TagSubprogram && e.Val(dwarf.AttrAbstractOrigin) == nil && e.Val(dwarf.AttrSpecification) == nil {
			if name, file, line := di.entryName(e); name != "" && file != "" && line > 0 {
				di.decls[file] = append(di.decls[file], decl{line, name})
			}
		}
		next := parent
		switch {
		case parent == nil && e.Tag == dwarf.TagSubprogram,
			parent != nil && e.Tag == dwarf.TagInlinedSubroutine:
			ranges, err := di.data.Ranges(e)
			if err != nil {
				return err
			}
			ranges = slices.DeleteFunc(ranges, func(r [2]uint64) bool { return isTombstone(r[0]) })
			if len(ranges) == 0 {
				next = nil
				break
			}
			s := &scope{ranges: ranges}
			s.name, s.declFile, s.declLine = di.entryName(e)
			if parent == nil {
				sort.Slice(s.ranges, func(i, j int) bool { return s.ranges[i][0] < s.ranges[j][0] })
				di.progs = append(di.progs, s)
			} else {
				s.callFile = fileName(files, e.Val(dwarf.AttrCallFile))
				if l, ok := e.Val(dwarf.AttrCallLine).(int64); ok {
					s.callLine = l
				}
				parent.children = append(parent.children, s)
			}
			next = s
		case e.Tag == dwarf.TagSubprogram, e.Tag == dwarf.TagInlinedSubroutine:
			// Nested subprogram, or inlined subroutine outside a function.
			next = nil
		}
		if !e.Children {
			continue
		}
		if next == nil && parent != nil || next == nil && (e.Tag == dwarf.TagSubprogram || e.Tag == dwarf.TagInlinedSubroutine) {
			r.SkipChildren()
			continue
		}
		if err := di.loadChildren(r, files, next); err != nil {
			return err
		}
	}
}

// entryName returns the name and declaration position of a function,
// following abstract_origin and specification links.
func (di *dwarfInfo) entryName(e *dwarf.Entry) (name, file string, line int64) {
	for depth := 0; e != nil && depth < 4; depth++ {
		if name == "" {
			if n, ok := e.Val(dwarf.AttrName).(string); ok {
				name = n
			} else if n, ok := e.Val(dwarf.AttrLinkageName).(string); ok {
				name = n
			}
		}
		if file == "" {
			file = fileName(di.filesFor(e.Offset), e.Val(dwarf.AttrDeclFile))
			if l, ok := e.Val(dwarf.AttrDeclLine).(int64); ok {
				line = l
			}
		}
		if name != "" && file != "" {
			return name, file, line
		}
		off, ok := e.Val(dwarf.AttrAbstractOrigin).(dwarf.Offset)
		if !ok {
			off, ok = e.Val(dwarf.AttrSpecification).(dwarf.Offset)
		}
		if !ok {
			break
		}
		r := di.data.Reader()
		r.Seek(off)
		e, _ = r.Next()
	}
	return name, file, line
}

// filesFor returns the file table of the CU containing the entry at off.
func (di *dwarfInfo) filesFor(off dwarf.Offset) []*dwarf.LineFile {
	i := sort.Search(len(di.cus), func(i int) bool { return di.cus[i].offset > off }) - 1
	if i < 0 {
		return nil
	}
	return di.cus[i].files
}

// isTombstone reports whether wasm-ld wrote addr for removed code. Offset 0 is
// the function count in the code section, so it is never code.
func isTombstone(addr uint64) bool {
	return addr == 0 || addr >= 0xfffffffe
}

func fileName(files []*dwarf.LineFile, v any) string {
	i, ok := v.(int64)
	if !ok || i < 0 || int(i) >= len(files) || files[i] == nil {
		return ""
	}
	return files[i].Name
}

// loopBodyPC returns the first line row shortly after the start of a loop body,
// as instructions like a range bounds check often have no row of their own.
func (di *dwarfInfo) loopBodyPC(pc uint64) uint64 {
	i := sort.Search(len(di.rows), func(i int) bool { return di.rows[i].addr >= pc })
	for ; i < len(di.rows) && di.rows[i].addr < pc+64; i++ {
		if di.rows[i].endSeq {
			break
		}
		if di.rows[i].line > 0 {
			return di.rows[i].addr
		}
	}
	return pc
}

// guessFunc returns the function declared closest before line in file.
func (di *dwarfInfo) guessFunc(file string, line int64) (string, int64) {
	ds := di.decls[file]
	i := sort.Search(len(ds), func(i int) bool { return ds[i].line > line }) - 1
	if i < 0 {
		return "", 0
	}
	return ds[i].name, ds[i].line
}

// line returns the source position for a code offset.
func (di *dwarfInfo) line(pc uint64) (string, int64) {
	i := sort.Search(len(di.rows), func(i int) bool { return di.rows[i].addr > pc }) - 1
	if i < 0 || di.rows[i].endSeq {
		return "", 0
	}
	return di.rows[i].file, di.rows[i].line
}

// frames returns the inlined call stack at pc, innermost function first, as
// pprof wants it. It returns nil if pc is not in any known function.
func (di *dwarfInfo) frames(pc uint64) []frame {
	i := sort.Search(len(di.progs), func(i int) bool { return di.progs[i].ranges[0][0] > pc }) - 1
	var prog *scope
	// Subprograms can have several ranges, so pc may be in a later range of
	// an earlier subprogram.
	for j := i; j >= 0 && j > i-64; j-- {
		if di.progs[j].contains(pc) {
			prog = di.progs[j]
			break
		}
	}
	if prog == nil {
		// Code with line rows but no subprogram, such as assembly.
		if file, line := di.line(pc); line > 0 {
			return []frame{{line: line, srcFile: file}}
		}
		return nil
	}
	chain := []*scope{prog}
	for s := prog; ; {
		var next *scope
		for _, c := range s.children {
			if c.contains(pc) {
				next = c
				break
			}
		}
		if next == nil {
			break
		}
		chain = append(chain, next)
		s = next
	}
	file, line := di.line(pc)
	out := make([]frame, 0, len(chain)+1)
	if inner := chain[len(chain)-1]; line > 0 && file != "" && inner.declFile != "" && file != inner.declFile {
		// An inline scope is missing, so guess the function from the
		// declarations in the line's file. The call line is unknown.
		if name, declLine := di.guessFunc(file, line); name != "" {
			out = append(out, frame{function: name, file: file, declLine: declLine, line: line, srcFile: file, guessed: true})
			file, line = "", 0
		}
	}
	for k := len(chain) - 1; k >= 0; k-- {
		s := chain[k]
		prologue := false
		if line == 0 && len(out) == 0 {
			prologue = true
			// No line row, as in the prologue where wasmtime's epoch check is.
			// Use the declaration line like Go does.
			line, file = s.declLine, s.declFile
		}
		out = append(out, frame{function: s.name, file: s.declFile, declLine: s.declLine, line: line, srcFile: file, prologue: prologue})
		// The caller is executing the line the inlined call is on.
		file, line = s.callFile, s.callLine
	}
	return out
}
