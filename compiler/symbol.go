package compiler

// This file manages symbols, that is, functions and globals. It reads their
// pragmas, determines the link name, etc.

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
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
	linkName string     // go:linkname, go:export
	module   string     // go:wasm-module
	exported bool       // go:export
	nobounds bool       // go:nobounds
	inline   inlineType // go:inline
}

// getFunction returns the LLVM function for the given *ssa.Function, creating
// it if needed. It can later be filled with compilerContext.createFunction().
func (c *compilerContext) getFunction(fn *ssa.Function) llvm.Value {
	info := c.getFunctionInfo(fn)
	return c.getFunctionRaw(fn.Signature, info)
}

func (c *compilerContext) getFunctionRaw(sig *types.Signature, info functionInfo) llvm.Value {
	llvmFn := c.mod.NamedFunction(info.linkName)
	if !llvmFn.IsNil() {
		return llvmFn
	}

	var retType llvm.Type
	if sig.Results() == nil {
		retType = c.ctx.VoidType()
	} else if sig.Results().Len() == 1 {
		retType = c.getLLVMType(sig.Results().At(0).Type())
	} else {
		results := make([]llvm.Type, 0, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			results = append(results, c.getLLVMType(sig.Results().At(i).Type()))
		}
		retType = c.ctx.StructType(results, false)
	}

	var paramInfos []paramInfo
	params := []*types.Var{}
	if sig.Recv() != nil {
		params = append(params, sig.Recv())
	}
	for i := 0; i < sig.Params().Len(); i++ {
		params = append(params, sig.Params().At(i))
	}
	for _, param := range params {
		paramType := c.getLLVMType(param.Type())
		paramFragmentInfos := expandFormalParamType(paramType, param.Name(), param.Type())
		paramInfos = append(paramInfos, paramFragmentInfos...)
	}

	// Add an extra parameter as the function context. This context is used in
	// closures and bound methods, but should be optimized away when not used.
	if !info.exported {
		paramInfos = append(paramInfos, paramInfo{llvmType: c.i8ptrType, name: "context", flags: 0})
		paramInfos = append(paramInfos, paramInfo{llvmType: c.i8ptrType, name: "parentHandle", flags: 0})
	}

	var paramTypes []llvm.Type
	for _, info := range paramInfos {
		paramTypes = append(paramTypes, info.llvmType)
	}

	fnType := llvm.FunctionType(retType, paramTypes, false)
	llvmFn = llvm.AddFunction(c.mod, info.linkName, fnType)

	dereferenceableOrNullKind := llvm.AttributeKindID("dereferenceable_or_null")
	for i, info := range paramInfos {
		if info.flags&paramIsDeferenceableOrNull == 0 {
			continue
		}
		if info.llvmType.TypeKind() == llvm.PointerTypeKind {
			el := info.llvmType.ElementType()
			size := c.targetData.TypeAllocSize(el)
			if size == 0 {
				// dereferenceable_or_null(0) appears to be illegal in LLVM.
				continue
			}
			dereferenceableOrNull := c.ctx.CreateEnumAttribute(dereferenceableOrNullKind, size)
			llvmFn.AddAttributeAtIndex(i+1, dereferenceableOrNull)
		}
	}

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
		for i, typ := range paramTypes {
			if typ.TypeKind() == llvm.PointerTypeKind {
				llvmFn.AddAttributeAtIndex(i+1, nocapture)
			}
		}
	}
	return llvmFn
}

// getFunctionInfo returns information about a function that is not directly
// present in *ssa.Function, such as the link name and whether it should be
// exported.
func (c *compilerContext) getFunctionInfo(f *ssa.Function) functionInfo {
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

// parsePragmas is used by getFunctionInfo to parse function pragmas such as
// //export or //go:noinline.
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
			case "//go:noinline":
				info.inline = inlineNone
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
	align    int    // go:align
}

// loadASTComments loads comments on globals from the AST, for use later in the
// program. In particular, they are required for //go:extern pragmas on globals.
func (c *compilerContext) loadASTComments(lprogram *loader.Program) {
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
func (c *compilerContext) getGlobal(g *ssa.Global) llvm.Value {
	info := c.getGlobalInfo(g)
	llvmGlobal := c.mod.NamedGlobal(info.linkName)
	if llvmGlobal.IsNil() {
		typ := g.Type().(*types.Pointer).Elem()
		llvmType := c.getLLVMType(typ)
		llvmGlobal = llvm.AddGlobal(c.mod, llvmType, info.linkName)
		if !info.extern {
			llvmGlobal.SetInitializer(llvm.ConstNull(llvmType))
			llvmGlobal.SetLinkage(llvm.InternalLinkage)
		}

		// Set alignment from the //go:align comment.
		var alignInBits uint32
		if info.align < 0 || info.align&(info.align-1) != 0 {
			// Check for power-of-two (or 0).
			// See: https://stackoverflow.com/a/108360
			c.addError(g.Pos(), "global variable alignment must be a positive power of two")
		} else {
			// Set the alignment only when it is a power of two.
			alignInBits = uint32(info.align) ^ uint32(info.align-1)
			if info.align > c.targetData.ABITypeAlignment(llvmType) {
				llvmGlobal.SetAlignment(info.align)
			}
		}

		if c.Debug() && !info.extern {
			// Add debug info.
			// TODO: this should be done for every global in the program, not just
			// the ones that are referenced from some code.
			pos := c.program.Fset.Position(g.Pos())
			diglobal := c.dibuilder.CreateGlobalVariableExpression(c.difiles[pos.Filename], llvm.DIGlobalVariableExpression{
				Name:        g.RelString(nil),
				LinkageName: info.linkName,
				File:        c.getDIFile(pos.Filename),
				Line:        pos.Line,
				Type:        c.getDIType(typ),
				LocalToUnit: false,
				Expr:        c.dibuilder.CreateExpression(nil),
				AlignInBits: alignInBits,
			})
			llvmGlobal.AddMetadata(0, diglobal)
		}
	}
	return llvmGlobal
}

// getGlobalInfo returns some information about a specific global.
func (c *compilerContext) getGlobalInfo(g *ssa.Global) globalInfo {
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
		case "//go:align":
			align, err := strconv.Atoi(parts[1])
			if err == nil {
				info.align = align
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
