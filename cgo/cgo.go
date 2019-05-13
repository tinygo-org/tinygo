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
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// cgoPackage holds all CGo-related information of a package.
type cgoPackage struct {
	generated       *ast.File
	generatedPos    token.Pos
	errors          []error
	dir             string
	fset            *token.FileSet
	tokenFiles      map[string]*token.File
	missingSymbols  map[string]struct{}
	constants       map[string]constantInfo
	functions       map[string]*functionInfo
	globals         map[string]globalInfo
	typedefs        map[string]*typedefInfo
	elaboratedTypes map[string]*elaboratedTypeInfo
}

// constantInfo stores some information about a CGo constant found by libclang
// and declared in the Go AST.
type constantInfo struct {
	expr *ast.BasicLit
	pos  token.Pos
}

// functionInfo stores some information about a CGo function found by libclang
// and declared in the AST.
type functionInfo struct {
	args    []paramInfo
	results *ast.FieldList
	pos     token.Pos
}

// paramInfo is a parameter of a CGo function (see functionInfo).
type paramInfo struct {
	name     string
	typeExpr ast.Expr
}

// typedefInfo contains information about a single typedef in C.
type typedefInfo struct {
	typeExpr ast.Expr
	pos      token.Pos
}

// elaboratedTypeInfo contains some information about an elaborated type
// (struct, union) found in the C AST.
type elaboratedTypeInfo struct {
	typeExpr ast.Expr
	pos      token.Pos
}

// globalInfo contains information about a declared global variable in C.
type globalInfo struct {
	typeExpr ast.Expr
	pos      token.Pos
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
var builtinAliases = map[string]struct{}{
	"char":      struct{}{},
	"schar":     struct{}{},
	"uchar":     struct{}{},
	"short":     struct{}{},
	"ushort":    struct{}{},
	"int":       struct{}{},
	"uint":      struct{}{},
	"long":      struct{}{},
	"ulong":     struct{}{},
	"longlong":  struct{}{},
	"ulonglong": struct{}{},
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

// Process extracts `import "C"` statements from the AST, parses the comment
// with libclang, and modifies the AST to use this information. It returns a
// newly created *ast.File that should be added to the list of to-be-parsed
// files. If there is one or more error, it returns these in the []error slice
// but still modifies the AST.
func Process(files []*ast.File, dir string, fset *token.FileSet, cflags []string) (*ast.File, []error) {
	p := &cgoPackage{
		dir:             dir,
		fset:            fset,
		tokenFiles:      map[string]*token.File{},
		missingSymbols:  map[string]struct{}{},
		constants:       map[string]constantInfo{},
		functions:       map[string]*functionInfo{},
		globals:         map[string]globalInfo{},
		typedefs:        map[string]*typedefInfo{},
		elaboratedTypes: map[string]*elaboratedTypeInfo{},
	}

	// Add a new location for the following file.
	generatedTokenPos := p.fset.AddFile(dir+"/!cgo.go", -1, 0)
	generatedTokenPos.SetLines([]int{0})
	p.generatedPos = generatedTokenPos.Pos(0)

	// Construct a new in-memory AST for CGo declarations of this package.
	unsafeImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			ValuePos: p.generatedPos,
			Kind:     token.STRING,
			Value:    "\"unsafe\"",
		},
		EndPos: p.generatedPos,
	}
	p.generated = &ast.File{
		Package: p.generatedPos,
		Name: &ast.Ident{
			NamePos: p.generatedPos,
			Name:    files[0].Name.Name,
		},
		Decls: []ast.Decl{
			&ast.GenDecl{
				TokPos: p.generatedPos,
				Tok:    token.IMPORT,
				Specs: []ast.Spec{
					unsafeImport,
				},
			},
		},
		Imports: []*ast.ImportSpec{unsafeImport},
	}

	// Find all C.* symbols.
	for _, f := range files {
		astutil.Apply(f, p.findMissingCGoNames, nil)
	}
	for name := range builtinAliases {
		p.missingSymbols["_Cgo_"+name] = struct{}{}
	}

	// Find `import "C"` statements in the file.
	for _, f := range files {
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

			pos := genDecl.Pos()
			if genDecl.Doc != nil {
				pos = genDecl.Doc.Pos()
			}
			position := fset.PositionFor(pos, true)
			p.parseFragment(cgoComment+cgoTypes, cflags, position.Filename, position.Line)

			// Remove this import declaration.
			f.Decls = append(f.Decls[:i], f.Decls[i+1:]...)
			i--
		}

		// Print the AST, for debugging.
		//ast.Print(fset, f)
	}

	// Declare functions found by libclang.
	p.addFuncDecls()

	// Declare stub function pointer values found by libclang.
	p.addFuncPtrDecls()

	// Declare globals found by libclang.
	p.addConstDecls()

	// Declare globals found by libclang.
	p.addVarDecls()

	// Forward C types to Go types (like C.uint32_t -> uint32).
	p.addTypeAliases()

	// Add type declarations for C types, declared using typedef in C.
	p.addTypedefs()

	// Add elaborated types for C structs and unions.
	p.addElaboratedTypes()

	// Patch the AST to use the declared types and functions.
	for _, f := range files {
		astutil.Apply(f, p.walker, nil)
	}

	// Print the newly generated in-memory AST, for debugging.
	//ast.Print(fset, p.generated)

	return p.generated, p.errors
}

// addFuncDecls adds the C function declarations found by libclang in the
// comment above the `import "C"` statement.
func (p *cgoPackage) addFuncDecls() {
	names := make([]string, 0, len(p.functions))
	for name := range p.functions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fn := p.functions[name]
		obj := &ast.Object{
			Kind: ast.Fun,
			Name: "C." + name,
		}
		args := make([]*ast.Field, len(fn.args))
		decl := &ast.FuncDecl{
			Name: &ast.Ident{
				NamePos: fn.pos,
				Name:    "C." + name,
				Obj:     obj,
			},
			Type: &ast.FuncType{
				Func: fn.pos,
				Params: &ast.FieldList{
					Opening: fn.pos,
					List:    args,
					Closing: fn.pos,
				},
				Results: fn.results,
			},
		}
		obj.Decl = decl
		for i, arg := range fn.args {
			args[i] = &ast.Field{
				Names: []*ast.Ident{
					&ast.Ident{
						NamePos: fn.pos,
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
		p.generated.Decls = append(p.generated.Decls, decl)
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
func (p *cgoPackage) addFuncPtrDecls() {
	if len(p.functions) == 0 {
		return
	}
	gen := &ast.GenDecl{
		TokPos: token.NoPos,
		Tok:    token.VAR,
		Lparen: token.NoPos,
		Rparen: token.NoPos,
	}
	names := make([]string, 0, len(p.functions))
	for name := range p.functions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fn := p.functions[name]
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: "C." + name + "$funcaddr",
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{&ast.Ident{
				NamePos: fn.pos,
				Name:    "C." + name + "$funcaddr",
				Obj:     obj,
			}},
			Type: &ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: fn.pos,
					Name:    "unsafe",
				},
				Sel: &ast.Ident{
					NamePos: fn.pos,
					Name:    "Pointer",
				},
			},
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
	}
	p.generated.Decls = append(p.generated.Decls, gen)
}

// addConstDecls declares external C constants in the Go source.
// It adds code like the following to the AST:
//
//     const (
//         C.CONST_INT = 5
//         C.CONST_FLOAT = 5.8
//         // ...
//     )
func (p *cgoPackage) addConstDecls() {
	if len(p.constants) == 0 {
		return
	}
	gen := &ast.GenDecl{
		TokPos: token.NoPos,
		Tok:    token.CONST,
		Lparen: token.NoPos,
		Rparen: token.NoPos,
	}
	names := make([]string, 0, len(p.constants))
	for name := range p.constants {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		constVal := p.constants[name]
		obj := &ast.Object{
			Kind: ast.Con,
			Name: "C." + name,
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{&ast.Ident{
				NamePos: constVal.pos,
				Name:    "C." + name,
				Obj:     obj,
			}},
			Values: []ast.Expr{constVal.expr},
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
	}
	p.generated.Decls = append(p.generated.Decls, gen)
}

// addVarDecls declares external C globals in the Go source.
// It adds code like the following to the AST:
//
//     var (
//         C.globalInt  int
//         C.globalBool bool
//         // ...
//     )
func (p *cgoPackage) addVarDecls() {
	if len(p.globals) == 0 {
		return
	}
	gen := &ast.GenDecl{
		TokPos: token.NoPos,
		Tok:    token.VAR,
		Lparen: token.NoPos,
		Rparen: token.NoPos,
	}
	names := make([]string, 0, len(p.globals))
	for name := range p.globals {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		global := p.globals[name]
		obj := &ast.Object{
			Kind: ast.Var,
			Name: "C." + name,
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{&ast.Ident{
				NamePos: global.pos,
				Name:    "C." + name,
				Obj:     obj,
			}},
			Type: global.typeExpr,
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
	}
	p.generated.Decls = append(p.generated.Decls, gen)
}

// addTypeAliases aliases some built-in Go types with their equivalent C types.
// It adds code like the following to the AST:
//
//     type (
//         C.int8_t  = int8
//         C.int16_t = int16
//         // ...
//     )
func (p *cgoPackage) addTypeAliases() {
	aliasKeys := make([]string, 0, len(cgoAliases))
	for key := range cgoAliases {
		aliasKeys = append(aliasKeys, key)
	}
	sort.Strings(aliasKeys)
	gen := &ast.GenDecl{
		TokPos: token.NoPos,
		Tok:    token.TYPE,
		Lparen: token.NoPos,
		Rparen: token.NoPos,
	}
	for _, typeName := range aliasKeys {
		goTypeName := cgoAliases[typeName]
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: typeName,
		}
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: token.NoPos,
				Name:    typeName,
				Obj:     obj,
			},
			Assign: p.generatedPos,
			Type: &ast.Ident{
				NamePos: token.NoPos,
				Name:    goTypeName,
			},
		}
		obj.Decl = typeSpec
		gen.Specs = append(gen.Specs, typeSpec)
	}
	p.generated.Decls = append(p.generated.Decls, gen)
}

func (p *cgoPackage) addTypedefs() {
	if len(p.typedefs) == 0 {
		return
	}
	gen := &ast.GenDecl{
		TokPos: token.NoPos,
		Tok:    token.TYPE,
	}
	names := make([]string, 0, len(p.typedefs))
	for name := range p.typedefs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		typedef := p.typedefs[name]
		typeName := "C." + name
		isAlias := true
		if strings.HasPrefix(name, "_Cgo_") {
			typeName = "C." + name[len("_Cgo_"):]
			isAlias = false // C.short etc. should not be aliased to the equivalent Go type (not portable)
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
				NamePos: typedef.pos,
				Name:    typeName,
				Obj:     obj,
			},
			Type: typedef.typeExpr,
		}
		if isAlias {
			typeSpec.Assign = typedef.pos
		}
		obj.Decl = typeSpec
		gen.Specs = append(gen.Specs, typeSpec)
	}
	p.generated.Decls = append(p.generated.Decls, gen)
}

// addElaboratedTypes adds C elaborated types as aliases. These are the "struct
// foo" or "union foo" types, often used in a typedef.
//
// See also:
// https://en.cppreference.com/w/cpp/language/elaborated_type_specifier
func (p *cgoPackage) addElaboratedTypes() {
	if len(p.elaboratedTypes) == 0 {
		return
	}
	gen := &ast.GenDecl{
		TokPos: token.NoPos,
		Tok:    token.TYPE,
	}
	names := make([]string, 0, len(p.elaboratedTypes))
	for name := range p.elaboratedTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		typ := p.elaboratedTypes[name]
		typeName := "C." + name
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: typeName,
		}
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: typ.pos,
				Name:    typeName,
				Obj:     obj,
			},
			Type: typ.typeExpr,
		}
		obj.Decl = typeSpec
		gen.Specs = append(gen.Specs, typeSpec)
	}
	p.generated.Decls = append(p.generated.Decls, gen)
}

// findMissingCGoNames traverses the AST and finds all C.something names. Only
// these symbols are extracted from the parsed C AST and converted to the Go
// equivalent.
func (p *cgoPackage) findMissingCGoNames(cursor *astutil.Cursor) bool {
	switch node := cursor.Node().(type) {
	case *ast.SelectorExpr:
		x, ok := node.X.(*ast.Ident)
		if !ok {
			return true
		}
		if x.Name == "C" {
			name := node.Sel.Name
			if _, ok := builtinAliases[name]; ok {
				name = "_Cgo_" + name
			}
			p.missingSymbols[name] = struct{}{}
		}
	}
	return true
}

// walker replaces all "C".<something> expressions to literal "C.<something>"
// expressions. Such expressions are impossible to write in Go (a dot cannot be
// used in the middle of a name) so in practice all C identifiers live in a
// separate namespace (no _Cgo_ hacks like in gc).
func (p *cgoPackage) walker(cursor *astutil.Cursor) bool {
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
		if _, ok := p.functions[fun.Sel.Name]; ok && x.Name == "C" {
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
			if _, ok := p.functions[node.Sel.Name]; ok {
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
