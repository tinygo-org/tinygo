package loader

// This file extracts the `import "C"` statement from the source and modifies
// the AST for Cgo. It does not use libclang directly (see libclang.go).

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// fileInfo holds all Cgo-related information of a given *ast.File.
type fileInfo struct {
	*ast.File
	*Package
	filename        string
	functions       map[string]*functionInfo
	globals         map[string]*globalInfo
	typedefs        map[string]*typedefInfo
	elaboratedTypes map[string]ast.Expr
	importCPos      token.Pos
}

// functionInfo stores some information about a Cgo function found by libclang
// and declared in the AST.
type functionInfo struct {
	args    []paramInfo
	results *ast.FieldList
}

// paramInfo is a parameter of a Cgo function (see functionInfo).
type paramInfo struct {
	name     string
	typeExpr ast.Expr
}

// typedefInfo contains information about a single typedef in C.
type typedefInfo struct {
	typeExpr ast.Expr
}

// globalInfo contains information about a declared global variable in C.
type globalInfo struct {
	typeExpr ast.Expr
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

// cgoTypes lists some C types with ambiguous sizes that must be retrieved
// somehow from C. This is done by adding some typedefs to get the size of each
// type.
const cgoTypes = `
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

// processCgo extracts the `import "C"` statement from the AST, parses the
// comment with libclang, and modifies the AST to use this information.
func (p *Package) processCgo(filename string, f *ast.File, cflags []string) []error {
	info := &fileInfo{
		File:            f,
		Package:         p,
		filename:        filename,
		functions:       map[string]*functionInfo{},
		globals:         map[string]*globalInfo{},
		typedefs:        map[string]*typedefInfo{},
		elaboratedTypes: map[string]ast.Expr{},
	}

	// Find `import "C"` statements in the file.
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
			panic("could not parse import path: " + err.Error())
		}
		if path != "C" {
			continue
		}
		cgoComment := genDecl.Doc.Text()

		// Stored for later use by generated functions, to use a somewhat sane
		// source location.
		info.importCPos = spec.Path.ValuePos

		pos := info.fset.PositionFor(genDecl.Doc.Pos(), true)
		errs := info.parseFragment(cgoComment+cgoTypes, cflags, pos.Filename, pos.Line)
		if errs != nil {
			return errs
		}

		// Remove this import declaration.
		f.Decls = append(f.Decls[:i], f.Decls[i+1:]...)
		i--
	}

	// Print the AST, for debugging.
	//ast.Print(p.fset, f)

	// Declare functions found by libclang.
	info.addFuncDecls()

	// Declare stub function pointer values found by libclang.
	info.addFuncPtrDecls()

	// Declare globals found by libclang.
	info.addVarDecls()

	// Forward C types to Go types (like C.uint32_t -> uint32).
	info.addTypeAliases()

	// Add type declarations for C types, declared using typedef in C.
	info.addTypedefs()

	// Add elaborated types for C structs and unions.
	info.addElaboratedTypes()

	// Patch the AST to use the declared types and functions.
	f = astutil.Apply(f, info.walker, nil).(*ast.File)

	return nil
}

// addFuncDecls adds the C function declarations found by libclang in the
// comment above the `import "C"` statement.
func (info *fileInfo) addFuncDecls() {
	// TODO: replace all uses of importCPos with the real locations from
	// libclang.
	names := make([]string, 0, len(info.functions))
	for name := range info.functions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fn := info.functions[name]
		obj := &ast.Object{
			Kind: ast.Fun,
			Name: "C." + name,
		}
		args := make([]*ast.Field, len(fn.args))
		decl := &ast.FuncDecl{
			Name: &ast.Ident{
				NamePos: info.importCPos,
				Name:    "C." + name,
				Obj:     obj,
			},
			Type: &ast.FuncType{
				Func: info.importCPos,
				Params: &ast.FieldList{
					Opening: info.importCPos,
					List:    args,
					Closing: info.importCPos,
				},
				Results: fn.results,
			},
		}
		obj.Decl = decl
		for i, arg := range fn.args {
			args[i] = &ast.Field{
				Names: []*ast.Ident{
					&ast.Ident{
						NamePos: info.importCPos,
						Name:    arg.name,
						Obj: &ast.Object{
							Kind: ast.Var,
							Name: arg.name,
							Decl: decl,
						},
					},
				},
				Type: arg.typeExpr,
			}
		}
		info.Decls = append(info.Decls, decl)
	}
}

// addFuncPtrDecls creates stub declarations of function pointer values. These
// values will later be replaced with the real values in the compiler.
// It adds code like the following to the AST:
//
//     var (
//         C.add unsafe.Pointer
//         C.mul unsafe.Pointer
//         // ...
//     )
func (info *fileInfo) addFuncPtrDecls() {
	gen := &ast.GenDecl{
		TokPos: info.importCPos,
		Tok:    token.VAR,
		Lparen: info.importCPos,
		Rparen: info.importCPos,
	}
	names := make([]string, 0, len(info.functions))
	for name := range info.functions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: "C." + name + "$funcaddr",
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{&ast.Ident{
				NamePos: info.importCPos,
				Name:    "C." + name + "$funcaddr",
				Obj:     obj,
			}},
			Type: &ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: info.importCPos,
					Name:    "unsafe",
				},
				Sel: &ast.Ident{
					NamePos: info.importCPos,
					Name:    "Pointer",
				},
			},
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
	}
	info.Decls = append(info.Decls, gen)
}

// addVarDecls declares external C globals in the Go source.
// It adds code like the following to the AST:
//
//     var (
//         C.globalInt  int
//         C.globalBool bool
//         // ...
//     )
func (info *fileInfo) addVarDecls() {
	gen := &ast.GenDecl{
		TokPos: info.importCPos,
		Tok:    token.VAR,
		Lparen: info.importCPos,
		Rparen: info.importCPos,
	}
	names := make([]string, 0, len(info.globals))
	for name := range info.globals {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		global := info.globals[name]
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: "C." + name,
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{&ast.Ident{
				NamePos: info.importCPos,
				Name:    "C." + name,
				Obj:     obj,
			}},
			Type: global.typeExpr,
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
	}
	info.Decls = append(info.Decls, gen)
}

// addTypeAliases aliases some built-in Go types with their equivalent C types.
// It adds code like the following to the AST:
//
//     type (
//         C.int8_t  = int8
//         C.int16_t = int16
//         // ...
//     )
func (info *fileInfo) addTypeAliases() {
	aliasKeys := make([]string, 0, len(cgoAliases))
	for key := range cgoAliases {
		aliasKeys = append(aliasKeys, key)
	}
	sort.Strings(aliasKeys)
	gen := &ast.GenDecl{
		TokPos: info.importCPos,
		Tok:    token.TYPE,
		Lparen: info.importCPos,
		Rparen: info.importCPos,
	}
	for _, typeName := range aliasKeys {
		goTypeName := cgoAliases[typeName]
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: typeName,
		}
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: info.importCPos,
				Name:    typeName,
				Obj:     obj,
			},
			Assign: info.importCPos,
			Type: &ast.Ident{
				NamePos: info.importCPos,
				Name:    goTypeName,
			},
		}
		obj.Decl = typeSpec
		gen.Specs = append(gen.Specs, typeSpec)
	}
	info.Decls = append(info.Decls, gen)
}

func (info *fileInfo) addTypedefs() {
	gen := &ast.GenDecl{
		TokPos: info.importCPos,
		Tok:    token.TYPE,
	}
	names := make([]string, 0, len(info.typedefs))
	for name := range info.typedefs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		typedef := info.typedefs[name]
		typeName := "C." + name
		if strings.HasPrefix(name, "_Cgo_") {
			typeName = "C." + name[len("_Cgo_"):]
		}
		if _, ok := cgoAliases[typeName]; ok {
			// This is a type that also exists in Go (defined in stdint.h).
			continue
		}
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: typeName,
		}
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: info.importCPos,
				Name:    typeName,
				Obj:     obj,
			},
			Type: typedef.typeExpr,
		}
		obj.Decl = typeSpec
		gen.Specs = append(gen.Specs, typeSpec)
	}
	info.Decls = append(info.Decls, gen)
}

// addElaboratedTypes adds C elaborated types as aliases. These are the "struct
// foo" or "union foo" types, often used in a typedef.
//
// See also:
// https://en.cppreference.com/w/cpp/language/elaborated_type_specifier
func (info *fileInfo) addElaboratedTypes() {
	gen := &ast.GenDecl{
		TokPos: info.importCPos,
		Tok:    token.TYPE,
	}
	names := make([]string, 0, len(info.elaboratedTypes))
	for name := range info.elaboratedTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		typ := info.elaboratedTypes[name]
		typeName := "C.struct_" + name
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: typeName,
		}
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: info.importCPos,
				Name:    typeName,
				Obj:     obj,
			},
			Type: typ,
		}
		obj.Decl = typeSpec
		gen.Specs = append(gen.Specs, typeSpec)
	}
	info.Decls = append(info.Decls, gen)
}

// walker replaces all "C".<something> expressions to literal "C.<something>"
// expressions. Such expressions are impossible to write in Go (a dot cannot be
// used in the middle of a name) so in practice all C identifiers live in a
// separate namespace (no _Cgo_ hacks like in gc).
func (info *fileInfo) walker(cursor *astutil.Cursor) bool {
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
		if _, ok := info.functions[fun.Sel.Name]; ok && x.Name == "C" {
			node.Fun = &ast.Ident{
				NamePos: x.NamePos,
				Name:    "C." + fun.Sel.Name,
			}
		}
	case *ast.SelectorExpr:
		x, ok := node.X.(*ast.Ident)
		if !ok {
			return true
		}
		if x.Name == "C" {
			name := "C." + node.Sel.Name
			if _, ok := info.functions[node.Sel.Name]; ok {
				name += "$funcaddr"
			}
			cursor.Replace(&ast.Ident{
				NamePos: x.NamePos,
				Name:    name,
			})
		}
	}
	return true
}
