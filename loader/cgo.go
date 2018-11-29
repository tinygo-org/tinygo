package loader

// This file extracts the `import "C"` statement from the source and modifies
// the AST for Cgo. It does not use libclang directly (see libclang.go).

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

// fileInfo holds all Cgo-related information of a given *ast.File.
type fileInfo struct {
	*ast.File
	filename   string
	functions  []*functionInfo
	typedefs   []*typedefInfo
	importCPos token.Pos
}

// functionInfo stores some information about a Cgo function found by libclang
// and declared in the AST.
type functionInfo struct {
	name   string
	args   []paramInfo
	result string
}

// paramInfo is a parameter of a Cgo function (see functionInfo).
type paramInfo struct {
	name     string
	typeName string
}

// typedefInfo contains information about a single typedef in C.
type typedefInfo struct {
	newName string // newly defined type name
	oldName string // old type name, may be something like "unsigned long long"
	size    int    // size in bytes
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
func (p *Package) processCgo(filename string, f *ast.File) error {
	info := &fileInfo{
		File:     f,
		filename: filename,
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

		err = info.parseFragment(cgoComment + cgoTypes)
		if err != nil {
			return err
		}

		// Remove this import declaration.
		f.Decls = append(f.Decls[:i], f.Decls[i+1:]...)
		i--
	}

	// Print the AST, for debugging.
	//ast.Print(p.fset, f)

	// Declare functions found by libclang.
	info.addFuncDecls()

	// Forward C types to Go types (like C.uint32_t -> uint32).
	info.addTypeAliases()

	// Add type declarations for C types, declared using typeef in C.
	info.addTypedefs()

	// Patch the AST to use the declared types and functions.
	ast.Inspect(f, info.walker)

	return nil
}

// addFuncDecls adds the C function declarations found by libclang in the
// comment above the `import "C"` statement.
func (info *fileInfo) addFuncDecls() {
	// TODO: replace all uses of importCPos with the real locations from
	// libclang.
	for _, fn := range info.functions {
		obj := &ast.Object{
			Kind: ast.Fun,
			Name: "C." + fn.name,
		}
		args := make([]*ast.Field, len(fn.args))
		decl := &ast.FuncDecl{
			Name: &ast.Ident{
				NamePos: info.importCPos,
				Name:    "C." + fn.name,
				Obj:     obj,
			},
			Type: &ast.FuncType{
				Func: info.importCPos,
				Params: &ast.FieldList{
					Opening: info.importCPos,
					List:    args,
					Closing: info.importCPos,
				},
				Results: &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{
							Type: &ast.Ident{
								NamePos: info.importCPos,
								Name:    "C." + fn.result,
							},
						},
					},
				},
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
							Name: "C." + arg.name,
							Decl: decl,
						},
					},
				},
				Type: &ast.Ident{
					NamePos: info.importCPos,
					Name:    "C." + arg.typeName,
				},
			}
		}
		info.Decls = append(info.Decls, decl)
	}
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
	for _, typedef := range info.typedefs {
		newType := "C." + typedef.newName
		oldType := "C." + typedef.oldName
		switch oldType {
		// TODO: plain char (may be signed or unsigned)
		case "C.signed char", "C.short", "C.int", "C.long", "C.long long":
			switch typedef.size {
			case 1:
				oldType = "int8"
			case 2:
				oldType = "int16"
			case 4:
				oldType = "int32"
			case 8:
				oldType = "int64"
			}
		case "C.unsigned char", "C.unsigned short", "C.unsigned int", "C.unsigned long", "C.unsigned long long":
			switch typedef.size {
			case 1:
				oldType = "uint8"
			case 2:
				oldType = "uint16"
			case 4:
				oldType = "uint32"
			case 8:
				oldType = "uint64"
			}
		}
		if strings.HasPrefix(newType, "C._Cgo_") {
			newType = "C." + newType[len("C._Cgo_"):]
		}
		if _, ok := cgoAliases[newType]; ok {
			// This is a type that also exists in Go (defined in stdint.h).
			continue
		}
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: newType,
		}
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: info.importCPos,
				Name:    newType,
				Obj:     obj,
			},
			Type: &ast.Ident{
				NamePos: info.importCPos,
				Name:    oldType,
			},
		}
		obj.Decl = typeSpec
		gen.Specs = append(gen.Specs, typeSpec)
	}
	info.Decls = append(info.Decls, gen)
}

// walker replaces all "C".<something> call expressions to literal
// "C.<something>" expressions. This is impossible to write in Go (a dot cannot
// be used in the middle of a name) so is used as a new namespace for C call
// expressions.
func (info *fileInfo) walker(node ast.Node) bool {
	switch node := node.(type) {
	case *ast.CallExpr:
		fun, ok := node.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := fun.X.(*ast.Ident)
		if !ok {
			return true
		}
		if x.Name == "C" {
			node.Fun = &ast.Ident{
				NamePos: x.NamePos,
				Name:    "C." + fun.Sel.Name,
			}
		}
	case *ast.ValueSpec:
		typ, ok := node.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := typ.X.(*ast.Ident)
		if !ok {
			return true
		}
		if x.Name == "C" {
			node.Type = &ast.Ident{
				NamePos: x.NamePos,
				Name:    "C." + typ.Sel.Name,
			}
		}
	}
	return true
}
