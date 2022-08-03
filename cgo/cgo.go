// Package cgo implements CGo by modifying a loaded AST. It does this by parsing
// the `import "C"` statements found in the source code with libclang and
// generating stub function and global declarations.
//
// There are a few advantages to modifying the AST directly instead of doing CGo
// as a preprocessing step, with the main advantage being that debug information
// is kept intact as much as possible.
package cgo

// This file extracts the `import "C"` statement from the source and modifies
// the AST for CGo. It does not use libclang directly: see libclang.go for the C
// source file parsing.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/shlex"
	"golang.org/x/tools/go/ast/astutil"
)

// cgoPackage holds all CGo-related information of a package.
type cgoPackage struct {
	generated       *ast.File
	generatedPos    token.Pos
	errors          []error
	currentDir      string // current working directory
	packageDir      string // full path to the package to process
	fset            *token.FileSet
	tokenFiles      map[string]*token.File
	definedGlobally map[string]ast.Node
	anonDecls       map[interface{}]string
	cflags          []string // CFlags from #cgo lines
	ldflags         []string // LDFlags from #cgo lines
	visitedFiles    map[string][]byte
}

// cgoFile holds information only for a single Go file (with one or more
// `import "C"` statements).
type cgoFile struct {
	*cgoPackage
	defined map[string]ast.Node
	names   map[string]clangCursor
}

// elaboratedTypeInfo contains some information about an elaborated type
// (struct, union) found in the C AST.
type elaboratedTypeInfo struct {
	typeExpr   *ast.StructType
	pos        token.Pos
	bitfields  []bitfieldInfo
	unionSize  int64 // union size in bytes, nonzero when union getters/setters should be created
	unionAlign int64 // union alignment in bytes
}

// bitfieldInfo contains information about a single bitfield in a struct. It
// keeps information about the start, end, and the special (renamed) base field
// of this bitfield.
type bitfieldInfo struct {
	field    *ast.Field
	name     string
	pos      token.Pos
	startBit int64
	endBit   int64 // may be 0 meaning "until the end of the field"
}

// cgoAliases list type aliases between Go and C, for types that are equivalent
// in both languages. See addTypeAliases.
var cgoAliases = map[string]string{
	"C.int8_t":    "int8",
	"C.int16_t":   "int16",
	"C.int32_t":   "int32",
	"C.int64_t":   "int64",
	"C.uint8_t":   "uint8",
	"C.uint16_t":  "uint16",
	"C.uint32_t":  "uint32",
	"C.uint64_t":  "uint64",
	"C.uintptr_t": "uintptr",
}

// builtinAliases are handled specially because they only exist on the Go side
// of CGo, not on the CGo side (they're prefixed with "_Cgo_" there).
var builtinAliases = []string{
	"char",
	"schar",
	"uchar",
	"short",
	"ushort",
	"int",
	"uint",
	"long",
	"ulong",
	"longlong",
	"ulonglong",
}

// builtinAliasTypedefs lists some C types with ambiguous sizes that must be
// retrieved somehow from C. This is done by adding some typedefs to get the
// size of each type.
const builtinAliasTypedefs = `
# 1 "<cgo>"
typedef char                _Cgo_char;
typedef signed char         _Cgo_schar;
typedef unsigned char       _Cgo_uchar;
typedef short               _Cgo_short;
typedef unsigned short      _Cgo_ushort;
typedef int                 _Cgo_int;
typedef unsigned int        _Cgo_uint;
typedef long                _Cgo_long;
typedef unsigned long       _Cgo_ulong;
typedef long long           _Cgo_longlong;
typedef unsigned long long  _Cgo_ulonglong;
`

// First part of the generated Go file. Written here as Go because that's much
// easier than constructing the entire AST in memory.
// The string/bytes functions below implement C.CString etc. To make sure the
// runtime doesn't need to know the C int type, lengths are converted to uintptr
// first.
// These functions will be modified to get a "C." prefix, so the source below
// doesn't reflect the final AST.
const generatedGoFilePrefix = `
import "unsafe"

var _ unsafe.Pointer

//go:linkname C.CString runtime.cgo_CString
func CString(string) *C.char

//go:linkname C.GoString runtime.cgo_GoString
func GoString(*C.char) string

//go:linkname C.__GoStringN runtime.cgo_GoStringN
func __GoStringN(*C.char, uintptr) string

func GoStringN(cstr *C.char, length C.int) string {
	return C.__GoStringN(cstr, uintptr(length))
}

//go:linkname C.__GoBytes runtime.cgo_GoBytes
func __GoBytes(unsafe.Pointer, uintptr) []byte

func GoBytes(ptr unsafe.Pointer, length C.int) []byte {
	return C.__GoBytes(ptr, uintptr(length))
}
`

// Process extracts `import "C"` statements from the AST, parses the comment
// with libclang, and modifies the AST to use this information. It returns a
// newly created *ast.File that should be added to the list of to-be-parsed
// files, the CGo header snippets that should be compiled (for inline
// functions), the CFLAGS and LDFLAGS found in #cgo lines, and a map of file
// hashes of the accessed C header files. If there is one or more error, it
// returns these in the []error slice but still modifies the AST.
func Process(files []*ast.File, dir string, fset *token.FileSet, cflags []string, clangHeaders string) (*ast.File, []string, []string, []string, map[string][]byte, []error) {
	p := &cgoPackage{
		currentDir:      dir,
		fset:            fset,
		tokenFiles:      map[string]*token.File{},
		definedGlobally: map[string]ast.Node{},
		anonDecls:       map[interface{}]string{},
		visitedFiles:    map[string][]byte{},
	}

	// Add a new location for the following file.
	generatedTokenPos := p.fset.AddFile(dir+"/!cgo.go", -1, 0)
	generatedTokenPos.SetLines([]int{0})
	p.generatedPos = generatedTokenPos.Pos(0)

	// Find the absolute path for this package.
	packagePath, err := filepath.Abs(fset.File(files[0].Pos()).Name())
	if err != nil {
		return nil, nil, nil, nil, nil, []error{
			scanner.Error{
				Pos: fset.Position(files[0].Pos()),
				Msg: "cgo: cannot find absolute path: " + err.Error(), // TODO: wrap this error
			},
		}
	}
	p.packageDir = filepath.Dir(packagePath)

	// Construct a new in-memory AST for CGo declarations of this package.
	// The first part is written as Go code that is then parsed, but more code
	// is added later to the AST to declare functions, globals, etc.
	goCode := "package " + files[0].Name.Name + "\n\n" + generatedGoFilePrefix
	p.generated, err = parser.ParseFile(fset, dir+"/!cgo.go", goCode, parser.ParseComments)
	if err != nil {
		// This is always a bug in the cgo package.
		panic("unexpected error: " + err.Error())
	}
	// If the Comments field is not set to nil, the go/format package will get
	// confused about where comments should go.
	p.generated.Comments = nil
	// Adjust some of the functions in there.
	for _, decl := range p.generated.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			switch decl.Name.Name {
			case "CString", "GoString", "GoStringN", "__GoStringN", "GoBytes", "__GoBytes":
				// Adjust the name to have a "C." prefix so it is correctly
				// resolved.
				decl.Name.Name = "C." + decl.Name.Name
			}
		}
	}
	// Patch some types, for example *C.char in C.CString.
	cf := p.newCGoFile()
	astutil.Apply(p.generated, func(cursor *astutil.Cursor) bool {
		return cf.walker(cursor, nil)
	}, nil)

	// Find `import "C"` C fragments in the file.
	cgoHeaders := make([]string, len(files)) // combined CGo header fragment for each file
	for i, f := range files {
		var cgoHeader string
		for i := 0; i < len(f.Decls); i++ {
			decl := f.Decls[i]
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if len(genDecl.Specs) != 1 {
				continue
			}
			spec, ok := genDecl.Specs[0].(*ast.ImportSpec)
			if !ok {
				continue
			}
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				// This should not happen. An import path that is not properly
				// quoted should not exist in a correct AST.
				panic("could not parse import path: " + err.Error())
			}
			if path != "C" {
				continue
			}

			// Remove this import declaration.
			f.Decls = append(f.Decls[:i], f.Decls[i+1:]...)
			i--

			if genDecl.Doc == nil {
				continue
			}

			// Iterate through all parts of the CGo header. Note that every //
			// line is a new comment.
			position := fset.Position(genDecl.Doc.Pos())
			fragment := fmt.Sprintf("# %d %#v\n", position.Line, position.Filename)
			for _, comment := range genDecl.Doc.List {
				// Find all #cgo lines, extract and use their contents, and
				// replace the lines with spaces (to preserve locations).
				c := p.parseCGoPreprocessorLines(comment.Text, comment.Slash)

				// Change the comment (which is still in Go syntax, with // and
				// /* */ ) to a regular string by replacing the start/end
				// markers of comments with spaces.
				// It is similar to the Text() method but differs in that it
				// doesn't strip anything and tries to keep all offsets correct
				// by adding spaces and newlines where necessary.
				if c[1] == '/' { /* comment */
					c = "  " + c[2:]
				} else { // comment
					c = "  " + c[2:len(c)-2]
				}
				fragment += c + "\n"
			}
			cgoHeader += fragment
		}

		cgoHeaders[i] = cgoHeader
	}

	// Define CFlags that will be used while parsing the package.
	// Disable _FORTIFY_SOURCE as it causes problems on macOS.
	// Note that it is only disabled for memcpy (etc) calls made from Go, which
	// have better alternatives anyway.
	cflagsForCGo := append([]string{"-D_FORTIFY_SOURCE=0"}, cflags...)
	cflagsForCGo = append(cflagsForCGo, p.cflags...)
	if clangHeaders != "" {
		cflagsForCGo = append(cflagsForCGo, "-isystem", clangHeaders)
	}

	// Retrieve types such as C.int, C.longlong, etc from C.
	p.newCGoFile().readNames(builtinAliasTypedefs, cflagsForCGo, "", func(names map[string]clangCursor) {
		gen := &ast.GenDecl{
			TokPos: token.NoPos,
			Tok:    token.TYPE,
		}
		for _, name := range builtinAliases {
			typeSpec := p.getIntegerType("C."+name, names["_Cgo_"+name])
			gen.Specs = append(gen.Specs, typeSpec)
		}
		p.generated.Decls = append(p.generated.Decls, gen)
	})

	// Process CGo imports for each file.
	for i, f := range files {
		cf := p.newCGoFile()
		cf.readNames(cgoHeaders[i], cflagsForCGo, filepath.Base(fset.File(f.Pos()).Name()), func(names map[string]clangCursor) {
			for _, name := range builtinAliases {
				// Names such as C.int should not be obtained from C.
				// This works around an issue in picolibc that has `#define int`
				// in a header file.
				delete(names, name)
			}
			astutil.Apply(f, func(cursor *astutil.Cursor) bool {
				return cf.walker(cursor, names)
			}, nil)
		})
	}

	// Print the newly generated in-memory AST, for debugging.
	//ast.Print(fset, p.generated)

	return p.generated, cgoHeaders, p.cflags, p.ldflags, p.visitedFiles, p.errors
}

func (p *cgoPackage) newCGoFile() *cgoFile {
	return &cgoFile{
		cgoPackage: p,
		defined:    make(map[string]ast.Node),
		names:      make(map[string]clangCursor),
	}
}

// makePathsAbsolute converts some common path compiler flags (-I, -L) from
// relative flags into absolute flags, if they are relative. This is necessary
// because the C compiler is usually not invoked from the package path.
func (p *cgoPackage) makePathsAbsolute(args []string) {
	nextIsPath := false
	for i, arg := range args {
		if nextIsPath {
			if !filepath.IsAbs(arg) {
				args[i] = filepath.Join(p.packageDir, arg)
			}
		}
		if arg == "-I" || arg == "-L" {
			nextIsPath = true
			continue
		}
		if strings.HasPrefix(arg, "-I") || strings.HasPrefix(arg, "-L") {
			path := arg[2:]
			if !filepath.IsAbs(path) {
				args[i] = arg[:2] + filepath.Join(p.packageDir, path)
			}
		}
	}
}

// parseCGoPreprocessorLines reads #cgo pseudo-preprocessor lines in the source
// text (import "C" fragment), stores their information such as CFLAGS, and
// returns the same text but with those #cgo lines replaced by spaces (to keep
// position offsets the same).
func (p *cgoPackage) parseCGoPreprocessorLines(text string, pos token.Pos) string {
	for {
		// Extract the #cgo line, and replace it with spaces.
		// Replacing with spaces makes sure that error locations are
		// still correct, while not interfering with parsing in any way.
		lineStart := strings.Index(text, "#cgo ")
		if lineStart < 0 {
			break
		}
		lineLen := strings.IndexByte(text[lineStart:], '\n')
		if lineLen < 0 {
			lineLen = len(text) - lineStart
		}
		lineEnd := lineStart + lineLen
		line := text[lineStart:lineEnd]
		spaces := make([]byte, len(line))
		for i := range spaces {
			spaces[i] = ' '
		}
		text = text[:lineStart] + string(spaces) + text[lineEnd:]

		// Get the text before the colon in the #cgo directive.
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			p.addErrorAfter(pos, text[:lineStart], "missing colon in #cgo line")
			continue
		}

		// Extract the fields before the colon. These fields are a list
		// of build tags and the C environment variable.
		fields := strings.Fields(line[4:colon])
		if len(fields) == 0 {
			p.addErrorAfter(pos, text[:lineStart+colon-1], "invalid #cgo line")
			continue
		}

		if len(fields) > 1 {
			p.addErrorAfter(pos, text[:lineStart+5], "not implemented: build constraints in #cgo line")
			continue
		}

		name := fields[len(fields)-1]
		value := line[colon+1:]
		switch name {
		case "CFLAGS":
			flags, err := shlex.Split(value)
			if err != nil {
				// TODO: find the exact location where the error happened.
				p.addErrorAfter(pos, text[:lineStart+colon+1], "failed to parse flags in #cgo line: "+err.Error())
				continue
			}
			if err := checkCompilerFlags(name, flags); err != nil {
				p.addErrorAfter(pos, text[:lineStart+colon+1], err.Error())
				continue
			}
			p.makePathsAbsolute(flags)
			p.cflags = append(p.cflags, flags...)
		case "LDFLAGS":
			flags, err := shlex.Split(value)
			if err != nil {
				// TODO: find the exact location where the error happened.
				p.addErrorAfter(pos, text[:lineStart+colon+1], "failed to parse flags in #cgo line: "+err.Error())
				continue
			}
			if err := checkLinkerFlags(name, flags); err != nil {
				p.addErrorAfter(pos, text[:lineStart+colon+1], err.Error())
				continue
			}
			p.makePathsAbsolute(flags)
			p.ldflags = append(p.ldflags, flags...)
		default:
			startPos := strings.LastIndex(line[4:colon], name) + 4
			p.addErrorAfter(pos, text[:lineStart+startPos], "invalid #cgo line: "+name)
			continue
		}
	}
	return text
}

// makeUnionField creates a new struct from an existing *elaboratedTypeInfo,
// that has just a single field that must be accessed through special accessors.
// It returns nil when there is an error. In case of an error, that error has
// already been added to the list of errors using p.addError.
func (p *cgoPackage) makeUnionField(typ *elaboratedTypeInfo) *ast.StructType {
	unionFieldTypeName, ok := map[int64]string{
		1: "uint8",
		2: "uint16",
		4: "uint32",
		8: "uint64",
	}[typ.unionAlign]
	if !ok {
		p.addError(typ.typeExpr.Struct, fmt.Sprintf("expected union alignment to be one of 1, 2, 4, or 8, but got %d", typ.unionAlign))
		return nil
	}
	var unionFieldType ast.Expr = &ast.Ident{
		NamePos: token.NoPos,
		Name:    unionFieldTypeName,
	}
	if typ.unionSize != typ.unionAlign {
		// A plain struct{uintX} isn't enough, we have to make a
		// struct{[N]uintX} to make the union big enough.
		if typ.unionSize/typ.unionAlign*typ.unionAlign != typ.unionSize {
			p.addError(typ.typeExpr.Struct, fmt.Sprintf("union alignment (%d) must be a multiple of union alignment (%d)", typ.unionSize, typ.unionAlign))
			return nil
		}
		unionFieldType = &ast.ArrayType{
			Len: &ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.FormatInt(typ.unionSize/typ.unionAlign, 10),
			},
			Elt: unionFieldType,
		}
	}
	return &ast.StructType{
		Struct: typ.typeExpr.Struct,
		Fields: &ast.FieldList{
			Opening: typ.typeExpr.Fields.Opening,
			List: []*ast.Field{{
				Names: []*ast.Ident{
					{
						NamePos: typ.typeExpr.Fields.Opening,
						Name:    "$union",
					},
				},
				Type: unionFieldType,
			}},
			Closing: typ.typeExpr.Fields.Closing,
		},
	}
}

// createUnionAccessor creates a function that returns a typed pointer to a
// union field for each field in a union. For example:
//
//	func (union *C.union_1) unionfield_d() *float64 {
//	    return (*float64)(unsafe.Pointer(&union.$union))
//	}
//
// Where C.union_1 is defined as:
//
//	type C.union_1 struct{
//	    $union uint64
//	}
//
// The returned pointer can be used to get or set the field, or get the pointer
// to a subfield.
func (p *cgoPackage) createUnionAccessor(field *ast.Field, typeName string) {
	if len(field.Names) != 1 {
		panic("number of names in union field must be exactly 1")
	}
	fieldName := field.Names[0]
	pos := fieldName.NamePos

	// The method receiver.
	receiver := &ast.SelectorExpr{
		X: &ast.Ident{
			NamePos: pos,
			Name:    "union",
			Obj:     nil,
		},
		Sel: &ast.Ident{
			NamePos: pos,
			Name:    "$union",
		},
	}

	// Get the address of the $union field.
	receiverPtr := &ast.UnaryExpr{
		Op: token.AND,
		X:  receiver,
	}

	// Cast to unsafe.Pointer.
	sourcePointer := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "unsafe"},
			Sel: &ast.Ident{Name: "Pointer"},
		},
		Args: []ast.Expr{receiverPtr},
	}

	// Cast to the target pointer type.
	targetPointer := &ast.CallExpr{
		Lparen: pos,
		Fun: &ast.ParenExpr{
			Lparen: pos,
			X: &ast.StarExpr{
				X: field.Type,
			},
			Rparen: pos,
		},
		Args:   []ast.Expr{sourcePointer},
		Rparen: pos,
	}

	// Create the accessor function.
	accessor := &ast.FuncDecl{
		Recv: &ast.FieldList{
			Opening: pos,
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{
							NamePos: pos,
							Name:    "union",
						},
					},
					Type: &ast.StarExpr{
						Star: pos,
						X: &ast.Ident{
							NamePos: pos,
							Name:    typeName,
							Obj:     nil,
						},
					},
				},
			},
			Closing: pos,
		},
		Name: &ast.Ident{
			NamePos: pos,
			Name:    "unionfield_" + fieldName.Name,
		},
		Type: &ast.FuncType{
			Func: pos,
			Params: &ast.FieldList{
				Opening: pos,
				Closing: pos,
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							Star: pos,
							X:    field.Type,
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			Lbrace: pos,
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Return: pos,
					Results: []ast.Expr{
						targetPointer,
					},
				},
			},
			Rbrace: pos,
		},
	}
	p.generated.Decls = append(p.generated.Decls, accessor)
}

// createBitfieldGetter creates a bitfield getter function like the following:
//
//	func (s *C.struct_foo) bitfield_b() byte {
//	    return (s.__bitfield_1 >> 5) & 0x1
//	}
func (p *cgoPackage) createBitfieldGetter(bitfield bitfieldInfo, typeName string) {
	// The value to return from the getter.
	// Not complete: this is just an expression to get the complete field.
	var result ast.Expr = &ast.SelectorExpr{
		X: &ast.Ident{
			NamePos: bitfield.pos,
			Name:    "s",
			Obj:     nil,
		},
		Sel: &ast.Ident{
			NamePos: bitfield.pos,
			Name:    bitfield.field.Names[0].Name,
		},
	}
	if bitfield.startBit != 0 {
		// Shift to the right by .startBit so that fields that come before are
		// shifted off.
		result = &ast.BinaryExpr{
			X:     result,
			OpPos: bitfield.pos,
			Op:    token.SHR,
			Y: &ast.BasicLit{
				ValuePos: bitfield.pos,
				Kind:     token.INT,
				Value:    strconv.FormatInt(bitfield.startBit, 10),
			},
		}
	}
	if bitfield.endBit != 0 {
		// Mask off the high bits so that fields that come after this field are
		// masked off.
		and := (uint64(1) << uint64(bitfield.endBit-bitfield.startBit)) - 1
		result = &ast.BinaryExpr{
			X:     result,
			OpPos: bitfield.pos,
			Op:    token.AND,
			Y: &ast.BasicLit{
				ValuePos: bitfield.pos,
				Kind:     token.INT,
				Value:    "0x" + strconv.FormatUint(and, 16),
			},
		}
	}

	// Create the getter function.
	getter := &ast.FuncDecl{
		Recv: &ast.FieldList{
			Opening: bitfield.pos,
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{
							NamePos: bitfield.pos,
							Name:    "s",
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: "s",
								Decl: nil,
							},
						},
					},
					Type: &ast.StarExpr{
						Star: bitfield.pos,
						X: &ast.Ident{
							NamePos: bitfield.pos,
							Name:    typeName,
							Obj:     nil,
						},
					},
				},
			},
			Closing: bitfield.pos,
		},
		Name: &ast.Ident{
			NamePos: bitfield.pos,
			Name:    "bitfield_" + bitfield.name,
		},
		Type: &ast.FuncType{
			Func: bitfield.pos,
			Params: &ast.FieldList{
				Opening: bitfield.pos,
				Closing: bitfield.pos,
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: bitfield.field.Type,
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			Lbrace: bitfield.pos,
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Return: bitfield.pos,
					Results: []ast.Expr{
						result,
					},
				},
			},
			Rbrace: bitfield.pos,
		},
	}
	p.generated.Decls = append(p.generated.Decls, getter)
}

// createBitfieldSetter creates a bitfield setter function like the following:
//
//	func (s *C.struct_foo) set_bitfield_b(value byte) {
//	    s.__bitfield_1 = s.__bitfield_1 ^ 0x60 | ((value & 1) << 5)
//	}
//
// Or the following:
//
//	func (s *C.struct_foo) set_bitfield_c(value byte) {
//	    s.__bitfield_1 = s.__bitfield_1 & 0x3f | (value << 6)
//	}
func (p *cgoPackage) createBitfieldSetter(bitfield bitfieldInfo, typeName string) {
	// The full field with all bitfields.
	var field ast.Expr = &ast.SelectorExpr{
		X: &ast.Ident{
			NamePos: bitfield.pos,
			Name:    "s",
			Obj:     nil,
		},
		Sel: &ast.Ident{
			NamePos: bitfield.pos,
			Name:    bitfield.field.Names[0].Name,
		},
	}
	// The value to insert into the field.
	var valueToInsert ast.Expr = &ast.Ident{
		NamePos: bitfield.pos,
		Name:    "value",
	}

	if bitfield.endBit != 0 {
		// Make sure the value is in range with a mask.
		valueToInsert = &ast.BinaryExpr{
			X:     valueToInsert,
			OpPos: bitfield.pos,
			Op:    token.AND,
			Y: &ast.BasicLit{
				ValuePos: bitfield.pos,
				Kind:     token.INT,
				Value:    "0x" + strconv.FormatUint((uint64(1)<<uint64(bitfield.endBit-bitfield.startBit))-1, 16),
			},
		}
		// Create a mask for the AND NOT operation.
		mask := ((uint64(1) << uint64(bitfield.endBit-bitfield.startBit)) - 1) << uint64(bitfield.startBit)
		// Zero the bits in the field that will soon be inserted.
		field = &ast.BinaryExpr{
			X:     field,
			OpPos: bitfield.pos,
			Op:    token.AND_NOT,
			Y: &ast.BasicLit{
				ValuePos: bitfield.pos,
				Kind:     token.INT,
				Value:    "0x" + strconv.FormatUint(mask, 16),
			},
		}
	} else { // bitfield.endBit == 0
		// We don't know exactly how many high bits should be zeroed. So we do
		// something different: keep the low bits with a mask and OR the new
		// value with it.
		mask := (uint64(1) << uint64(bitfield.startBit)) - 1
		// Extract the lower bits.
		field = &ast.BinaryExpr{
			X:     field,
			OpPos: bitfield.pos,
			Op:    token.AND,
			Y: &ast.BasicLit{
				ValuePos: bitfield.pos,
				Kind:     token.INT,
				Value:    "0x" + strconv.FormatUint(mask, 16),
			},
		}
	}

	// Bitwise OR with the new value (after the new value has been shifted).
	field = &ast.BinaryExpr{
		X:     field,
		OpPos: bitfield.pos,
		Op:    token.OR,
		Y: &ast.BinaryExpr{
			X:     valueToInsert,
			OpPos: bitfield.pos,
			Op:    token.SHL,
			Y: &ast.BasicLit{
				ValuePos: bitfield.pos,
				Kind:     token.INT,
				Value:    strconv.FormatInt(bitfield.startBit, 10),
			},
		},
	}

	// Create the setter function.
	setter := &ast.FuncDecl{
		Recv: &ast.FieldList{
			Opening: bitfield.pos,
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{
							NamePos: bitfield.pos,
							Name:    "s",
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: "s",
								Decl: nil,
							},
						},
					},
					Type: &ast.StarExpr{
						Star: bitfield.pos,
						X: &ast.Ident{
							NamePos: bitfield.pos,
							Name:    typeName,
							Obj:     nil,
						},
					},
				},
			},
			Closing: bitfield.pos,
		},
		Name: &ast.Ident{
			NamePos: bitfield.pos,
			Name:    "set_bitfield_" + bitfield.name,
		},
		Type: &ast.FuncType{
			Func: bitfield.pos,
			Params: &ast.FieldList{
				Opening: bitfield.pos,
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								NamePos: bitfield.pos,
								Name:    "value",
								Obj:     nil,
							},
						},
						Type: bitfield.field.Type,
					},
				},
				Closing: bitfield.pos,
			},
		},
		Body: &ast.BlockStmt{
			Lbrace: bitfield.pos,
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X: &ast.Ident{
								NamePos: bitfield.pos,
								Name:    "s",
								Obj:     nil,
							},
							Sel: &ast.Ident{
								NamePos: bitfield.pos,
								Name:    bitfield.field.Names[0].Name,
							},
						},
					},
					TokPos: bitfield.pos,
					Tok:    token.ASSIGN,
					Rhs: []ast.Expr{
						field,
					},
				},
			},
			Rbrace: bitfield.pos,
		},
	}
	p.generated.Decls = append(p.generated.Decls, setter)
}

// isEquivalentAST returns whether the given two AST nodes are equivalent as far
// as CGo is concerned. This is used to check that C types, globals, etc defined
// in different CGo header snippets are actually the same type (and probably
// even defined in the same header file, just in different translation units).
func (p *cgoPackage) isEquivalentAST(a, b ast.Node) bool {
	switch node := a.(type) {
	case *ast.ArrayType:
		b, ok := b.(*ast.ArrayType)
		if !ok {
			return false
		}
		if !p.isEquivalentAST(node.Len, b.Len) {
			return false
		}
		return p.isEquivalentAST(node.Elt, b.Elt)
	case *ast.BasicLit:
		b, ok := b.(*ast.BasicLit)
		if !ok {
			return false
		}
		// Note: this comparison is not correct in general ("1e2" equals "100"),
		// but is correct for its use in the cgo package.
		return node.Value == b.Value
	case *ast.CommentGroup:
		b, ok := b.(*ast.CommentGroup)
		if !ok {
			return false
		}
		if len(node.List) != len(b.List) {
			return false
		}
		for i, c := range node.List {
			if c.Text != b.List[i].Text {
				return false
			}
		}
		return true
	case *ast.FieldList:
		b, ok := b.(*ast.FieldList)
		if !ok {
			return false
		}
		if len(node.List) != len(b.List) {
			return false
		}
		for i, f := range node.List {
			if !p.isEquivalentAST(f, b.List[i]) {
				return false
			}
		}
		return true
	case *ast.Field:
		b, ok := b.(*ast.Field)
		if !ok {
			return false
		}
		if !p.isEquivalentAST(node.Type, b.Type) {
			return false
		}
		if len(node.Names) != len(b.Names) {
			return false
		}
		for i, name := range node.Names {
			if name.Name != b.Names[i].Name {
				return false
			}
		}
		return true
	case *ast.FuncDecl:
		b, ok := b.(*ast.FuncDecl)
		if !ok {
			return false
		}
		if node.Name.Name != b.Name.Name {
			return false
		}
		if node.Doc != b.Doc {
			if !p.isEquivalentAST(node.Doc, b.Doc) {
				return false
			}
		}
		if node.Recv != b.Recv {
			if !p.isEquivalentAST(node.Recv, b.Recv) {
				return false
			}
		}
		if !p.isEquivalentAST(node.Type.Params, b.Type.Params) {
			return false
		}
		return p.isEquivalentAST(node.Type.Results, b.Type.Results)
	case *ast.GenDecl:
		b, ok := b.(*ast.GenDecl)
		if !ok {
			return false
		}
		if node.Doc != b.Doc {
			if !p.isEquivalentAST(node.Doc, b.Doc) {
				return false
			}
		}
		if len(node.Specs) != len(b.Specs) {
			return false
		}
		for i, s := range node.Specs {
			if !p.isEquivalentAST(s, b.Specs[i]) {
				return false
			}
		}
		return true
	case *ast.Ident:
		b, ok := b.(*ast.Ident)
		if !ok {
			return false
		}
		return node.Name == b.Name
	case *ast.SelectorExpr:
		b, ok := b.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if !p.isEquivalentAST(node.Sel, b.Sel) {
			return false
		}
		return p.isEquivalentAST(node.X, b.X)
	case *ast.StarExpr:
		b, ok := b.(*ast.StarExpr)
		if !ok {
			return false
		}
		return p.isEquivalentAST(node.X, b.X)
	case *ast.StructType:
		b, ok := b.(*ast.StructType)
		if !ok {
			return false
		}
		return p.isEquivalentAST(node.Fields, b.Fields)
	case *ast.TypeSpec:
		b, ok := b.(*ast.TypeSpec)
		if !ok {
			return false
		}
		if node.Name.Name != b.Name.Name {
			return false
		}
		if node.Assign.IsValid() != b.Assign.IsValid() {
			return false
		}
		return p.isEquivalentAST(node.Type, b.Type)
	case *ast.ValueSpec:
		b, ok := b.(*ast.ValueSpec)
		if !ok {
			return false
		}
		if len(node.Names) != len(b.Names) {
			return false
		}
		for i, name := range node.Names {
			if name.Name != b.Names[i].Name {
				return false
			}
		}
		if node.Type != b.Type && !p.isEquivalentAST(node.Type, b.Type) {
			return false
		}
		if len(node.Values) != len(b.Values) {
			return false
		}
		for i, value := range node.Values {
			if !p.isEquivalentAST(value, b.Values[i]) {
				return false
			}
		}
		return true
	case nil:
		p.addError(token.NoPos, "internal error: AST node is nil")
		return true
	default:
		p.addError(a.Pos(), fmt.Sprintf("internal error: unknown AST node: %T", a))
		return true
	}
}

// getPos returns node.Pos(), and tries to obtain a closely related position if
// that fails.
func getPos(node ast.Node) token.Pos {
	pos := node.Pos()
	if pos.IsValid() {
		return pos
	}
	if decl, ok := node.(*ast.GenDecl); ok {
		// *ast.GenDecl often doesn't have TokPos defined, so look at the first
		// spec.
		return getPos(decl.Specs[0])
	}
	return token.NoPos
}

// getUnnamedDeclName creates a name (with the given prefix) for the given C
// declaration. This is used for structs, unions, and enums that are often
// defined without a name and used in a typedef.
func (p *cgoPackage) getUnnamedDeclName(prefix string, itf interface{}) string {
	if name, ok := p.anonDecls[itf]; ok {
		return name
	}
	name := prefix + strconv.Itoa(len(p.anonDecls))
	p.anonDecls[itf] = name
	return name
}

// getASTDeclName will declare the given C AST node (if not already defined) and
// will return its name, in the form of C.foo.
func (f *cgoFile) getASTDeclName(name string, found clangCursor, iscall bool) string {
	// Some types are defined in stdint.h and map directly to a particular Go
	// type.
	if alias := cgoAliases["C."+name]; alias != "" {
		return alias
	}
	node := f.getASTDeclNode(name, found, iscall)
	if _, ok := node.(*ast.FuncDecl); ok && !iscall {
		return "C." + name + "$funcaddr"
	}
	return "C." + name
}

// getASTDeclNode will declare the given C AST node (if not already defined) and
// returns it.
func (f *cgoFile) getASTDeclNode(name string, found clangCursor, iscall bool) ast.Node {
	if node, ok := f.defined[name]; ok {
		// Declaration was found in the current file, so return it immediately.
		return node
	}

	if node, ok := f.definedGlobally[name]; ok {
		// Declaration was previously created, but not for the current file. It
		// may be different (because it comes from a different CGo snippet), so
		// we need to check whether the AST for this definition is equivalent.
		f.defined[name] = nil
		newNode, _ := f.createASTNode(name, found)
		if !f.isEquivalentAST(node, newNode) {
			// It's not. Return a nice error with both locations.
			// Original cgo reports an error like
			//   cgo: inconsistent definitions for C.myint
			// which is far less helpful.
			f.addError(getPos(node), "defined previously at "+f.fset.Position(getPos(newNode)).String()+" with a different type")
		}
		f.defined[name] = node
		return node
	}

	// The declaration has no AST node. Create it now.
	f.defined[name] = nil
	node, elaboratedType := f.createASTNode(name, found)
	f.defined[name] = node
	f.definedGlobally[name] = node
	switch node := node.(type) {
	case *ast.FuncDecl:
		f.generated.Decls = append(f.generated.Decls, node)
		// Also add a declaration like the following:
		//   var C.foo$funcaddr unsafe.Pointer
		f.generated.Decls = append(f.generated.Decls, &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{{Name: "C." + name + "$funcaddr"}},
					Type: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "unsafe"},
						Sel: &ast.Ident{Name: "Pointer"},
					},
				},
			},
		})
	case *ast.GenDecl:
		f.generated.Decls = append(f.generated.Decls, node)
	case *ast.TypeSpec:
		f.generated.Decls = append(f.generated.Decls, &ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{node},
		})
	case nil:
		// Node may be nil in case of an error. In that case, just don't add it
		// as a declaration.
	default:
		panic("unexpected AST node")
	}

	// If this is a struct or union it may need bitfields or union accessor
	// methods.
	if elaboratedType != nil {
		// Add struct bitfields.
		for _, bitfield := range elaboratedType.bitfields {
			f.createBitfieldGetter(bitfield, "C."+name)
			f.createBitfieldSetter(bitfield, "C."+name)
		}
		if elaboratedType.unionSize != 0 {
			// Create union getters/setters.
			for _, field := range elaboratedType.typeExpr.Fields.List {
				if len(field.Names) != 1 {
					f.addError(elaboratedType.pos, fmt.Sprintf("union must have field with a single name, it has %d names", len(field.Names)))
					continue
				}
				f.createUnionAccessor(field, "C."+name)
			}
		}
	}

	return node
}

// walker replaces all "C".<something> expressions to literal "C.<something>"
// expressions. Such expressions are impossible to write in Go (a dot cannot be
// used in the middle of a name) so in practice all C identifiers live in a
// separate namespace (no _Cgo_ hacks like in gc).
func (f *cgoFile) walker(cursor *astutil.Cursor, names map[string]clangCursor) bool {
	switch node := cursor.Node().(type) {
	case *ast.CallExpr:
		fun, ok := node.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := fun.X.(*ast.Ident)
		if !ok {
			return true
		}
		if found, ok := names[fun.Sel.Name]; ok && x.Name == "C" {
			node.Fun = &ast.Ident{
				NamePos: x.NamePos,
				Name:    f.getASTDeclName(fun.Sel.Name, found, true),
			}
		}
	case *ast.SelectorExpr:
		x, ok := node.X.(*ast.Ident)
		if !ok {
			return true
		}
		if x.Name == "C" {
			name := "C." + node.Sel.Name
			if found, ok := names[node.Sel.Name]; ok {
				name = f.getASTDeclName(node.Sel.Name, found, false)
			}
			cursor.Replace(&ast.Ident{
				NamePos: x.NamePos,
				Name:    name,
			})
		}
	}
	return true
}

// renameFieldKeywords renames all reserved words in Go to some other field name
// with a "_" prefix. For example, it renames `type` to `_type`.
//
// See: https://golang.org/cmd/cgo/#hdr-Go_references_to_C
func renameFieldKeywords(fieldList *ast.FieldList) {
	renameFieldName(fieldList, "type")
}

// renameFieldName renames a given field name to a name with a "_" prepended. It
// makes sure to do the same thing for any field sharing the same name.
func renameFieldName(fieldList *ast.FieldList, name string) {
	var ident *ast.Ident
	for _, f := range fieldList.List {
		for _, n := range f.Names {
			if n.Name == name {
				ident = n
			}
		}
	}
	if ident == nil {
		return
	}
	renameFieldName(fieldList, "_"+name)
	ident.Name = "_" + ident.Name
}
