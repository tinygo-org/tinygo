package compiler

// This file manages symbols, that is, functions and globals. It reads their
// pragmas, determines the link name, etc.

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

type inlineType int

// How much to inline.
const (
	// Default behavior. The compiler decides for itself whether any given
	// function will be inlined. Whether any function is inlined depends on the
	// optimization level.
	inlineDefault inlineType = iota

	// Inline hint, just like the C inline keyword (signalled using
	// //go:inline). The compiler will be more likely to inline this function,
	// but it is not a guarantee.
	inlineHint

	// Don't inline, just like the GCC noinline attribute. Signalled using
	// //go:noinline.
	inlineNone
)

// functionInfo contains some information about a function or method. In
// particular, it contains information obtained from pragmas.
//
// The linkName value contains a valid link name, even though //go:linkname is
// not present.
type functionInfo struct {
	linkName  string     // go:linkname, go:export, go:interrupt
	module    string     // go:wasm-module
	exported  bool       // go:export
	nobounds  bool       // go:nobounds
	flag      bool       // used by dead code elimination
	interrupt bool       // go:interrupt
	inline    inlineType // go:inline
}

func (c *Compiler) getFunction(fn *ssa.Function) llvm.Value {
	info := c.getFunctionInfo(fn)
	llvmFn := c.mod.NamedFunction(info.linkName)
	if llvmFn.IsNil() {
		llvmFn = c.declareFunction(fn, info)
	}
	if _, ok := c.functionsToCompileSet[fn]; !ok {
		if fn.Blocks != nil {
			c.functionsToCompile = append(c.functionsToCompile, fn)
			c.functionsToCompileSet[fn] = struct{}{}
		}
	}
	return llvmFn
}

func (c *Compiler) declareFunction(fn *ssa.Function, info functionInfo) llvm.Value {
	var paramTypes []llvm.Type
	for _, param := range fn.Params {
		paramType := c.getLLVMType(param.Type())
		paramTypeFragments := c.expandFormalParamType(paramType)
		paramTypes = append(paramTypes, paramTypeFragments...)
	}

	// Add an extra parameter as the function context. This context is used in
	// closures and bound methods, but should be optimized away when not used.
	if !info.exported {
		paramTypes = append(paramTypes, c.i8ptrType) // context
		paramTypes = append(paramTypes, c.i8ptrType) // parent coroutine
	}

	returnType := c.getFuncReturnType(fn.Signature)
	fnType := llvm.FunctionType(returnType, paramTypes, false)

	llvmFn := llvm.AddFunction(c.mod, info.linkName, fnType)

	// External/exported functions may not retain pointer values.
	// https://golang.org/cmd/cgo/#hdr-Passing_pointers
	if info.exported {
		// Set the wasm-import-module attribute if the function's module is set.
		if info.module != "" {
			wasmImportModuleAttr := c.ctx.CreateStringAttribute("wasm-import-module", info.module)
			llvmFn.AddFunctionAttr(wasmImportModuleAttr)
		}
		nocaptureKind := llvm.AttributeKindID("nocapture")
		nocapture := c.ctx.CreateEnumAttribute(nocaptureKind, 0)
		for i, typ := range fnType.ParamTypes() {
			if typ.TypeKind() == llvm.PointerTypeKind {
				llvmFn.AddAttributeAtIndex(i+1, nocapture)
			}
		}
	}
	return llvmFn
}

func (c *Compiler) getFunctionInfo(f *ssa.Function) functionInfo {
	info := functionInfo{}
	if strings.HasPrefix(f.Name(), "C.") {
		// Created by CGo: such a name cannot be created by regular C code.
		info.linkName = f.Name()[2:]
		info.exported = true
	} else {
		// Pick the default linkName.
		info.linkName = f.RelString(nil)
		// Check for //go: pragmas, which may change the link name (among
		// others).
		info.parsePragmas(f)
	}
	return info
}

func (info *functionInfo) parsePragmas(f *ssa.Function) {
	// Parse compiler directives in the preceding comments.
	if f.Syntax() == nil {
		return
	}
	if decl, ok := f.Syntax().(*ast.FuncDecl); ok && decl.Doc != nil {
		for _, comment := range decl.Doc.List {
			text := comment.Text
			if strings.HasPrefix(text, "//export ") {
				// Rewrite '//export' to '//go:export' for compatibility with
				// gc.
				text = "//go:" + text[2:]
			}
			if !strings.HasPrefix(text, "//go:") {
				continue
			}
			parts := strings.Fields(text)
			switch parts[0] {
			case "//go:export":
				if len(parts) != 2 {
					continue
				}
				info.linkName = parts[1]
				info.exported = true
			case "//go:wasm-module":
				// Alternative comment for setting the import module.
				if len(parts) != 2 {
					continue
				}
				info.module = parts[1]
			case "//go:inline":
				info.inline = inlineHint
			case "//go:interrupt":
				if len(parts) != 2 {
					continue
				}
				name := parts[1]
				if strings.HasSuffix(name, "_vect") {
					// AVR vector naming
					name = "__vector_" + name[:len(name)-5]
				}
				info.linkName = name
				info.exported = true
				info.interrupt = true
			case "//go:linkname":
				if len(parts) != 3 || parts[1] != f.Name() {
					continue
				}
				// Only enable go:linkname when the package imports "unsafe".
				// This is a slightly looser requirement than what gc uses: gc
				// requires the file to import "unsafe", not the package as a
				// whole.
				if hasUnsafeImport(f.Pkg.Pkg) {
					info.linkName = parts[2]
				}
			case "//go:nobounds":
				// Skip bounds checking in this function. Useful for some
				// runtime functions.
				// This is somewhat dangerous and thus only imported in packages
				// that import unsafe.
				if hasUnsafeImport(f.Pkg.Pkg) {
					info.nobounds = true
				}
			}
		}
	}
}

// globalInfo contains some information about a specific global. By default,
// linkName is equal to .RelString(nil) on a global and extern is false, but for
// some symbols this is different (due to //go:extern for example).
type globalInfo struct {
	linkName string // go:extern
	extern   bool   // go:extern
}

// loadASTComments loads comments on globals from the AST, for use later in the
// program. In particular, they are required for //go:extern pragmas on globals.
func (c *Compiler) loadASTComments(lprogram *loader.Program) {
	c.astComments = map[string]*ast.CommentGroup{}
	for _, pkgInfo := range lprogram.Sorted() {
		for _, file := range pkgInfo.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					switch decl.Tok {
					case token.VAR:
						if len(decl.Specs) != 1 {
							continue
						}
						for _, spec := range decl.Specs {
							switch spec := spec.(type) {
							case *ast.ValueSpec: // decl.Tok == token.VAR
								for _, name := range spec.Names {
									id := pkgInfo.Pkg.Path() + "." + name.Name
									c.astComments[id] = decl.Doc
								}
							}
						}
					}
				}
			}
		}
	}
}

// getGlobal returns a LLVM IR global value for a Go SSA global. It is added to
// the LLVM IR if it has not been added already.
func (c *Compiler) getGlobal(g *ssa.Global) llvm.Value {
	info := c.getGlobalInfo(g)
	llvmGlobal := c.mod.NamedGlobal(info.linkName)
	if llvmGlobal.IsNil() {
		llvmType := c.getLLVMType(g.Type().(*types.Pointer).Elem())
		llvmGlobal = llvm.AddGlobal(c.mod, llvmType, info.linkName)
		if !info.extern {
			llvmGlobal.SetInitializer(llvm.ConstNull(llvmType))
			llvmGlobal.SetLinkage(llvm.InternalLinkage)
		}
	}
	return llvmGlobal
}

// getGlobalInfo returns some information about a specific global.
func (c *Compiler) getGlobalInfo(g *ssa.Global) globalInfo {
	info := globalInfo{}
	if strings.HasPrefix(g.Name(), "C.") {
		// Created by CGo: such a name cannot be created by regular C code.
		info.linkName = g.Name()[2:]
		info.extern = true
	} else {
		// Pick the default linkName.
		info.linkName = g.RelString(nil)
		// Check for //go: pragmas, which may change the link name (among
		// others).
		doc := c.astComments[info.linkName]
		if doc != nil {
			info.parsePragmas(doc)
		}
	}
	return info
}

// Parse //go: pragma comments from the source. In particular, it parses the
// //go:extern pragma on globals.
func (info *globalInfo) parsePragmas(doc *ast.CommentGroup) {
	for _, comment := range doc.List {
		if !strings.HasPrefix(comment.Text, "//go:") {
			continue
		}
		parts := strings.Fields(comment.Text)
		switch parts[0] {
		case "//go:extern":
			info.extern = true
			if len(parts) == 2 {
				info.linkName = parts[1]
			}
		}
	}
}

// Get all methods of a type.
func getAllMethods(prog *ssa.Program, typ types.Type) []*types.Selection {
	ms := prog.MethodSets.MethodSet(typ)
	methods := make([]*types.Selection, ms.Len())
	for i := 0; i < ms.Len(); i++ {
		methods[i] = ms.At(i)
	}
	return methods
}

// Return true if this package imports "unsafe", false otherwise.
func hasUnsafeImport(pkg *types.Package) bool {
	for _, imp := range pkg.Imports() {
		if imp == types.Unsafe {
			return true
		}
	}
	return false
}
