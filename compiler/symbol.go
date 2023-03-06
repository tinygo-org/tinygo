package compiler

// This file manages symbols, that is, functions and globals. It reads their
// pragmas, determines the link name, etc.

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// functionInfo contains some information about a function or method. In
// particular, it contains information obtained from pragmas.
//
// The linkName value contains a valid link name, even if //go:linkname is not
// present.
type functionInfo struct {
	module     string     // go:wasm-module
	importName string     // go:linkname, go:export - The name the developer assigns
	linkName   string     // go:linkname, go:export - The name that we map for the particular module -> importName
	section    string     // go:section - object file section name
	exported   bool       // go:export, CGo
	interrupt  bool       // go:interrupt
	nobounds   bool       // go:nobounds
	variadic   bool       // go:variadic (CGo only)
	inline     inlineType // go:inline
}

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

// getFunction returns the LLVM function for the given *ssa.Function, creating
// it if needed. It can later be filled with compilerContext.createFunction().
func (c *compilerContext) getFunction(fn *ssa.Function) (llvm.Type, llvm.Value) {
	info := c.getFunctionInfo(fn)
	llvmFn := c.mod.NamedFunction(info.linkName)
	if !llvmFn.IsNil() {
		return llvmFn.GlobalValueType(), llvmFn
	}

	var retType llvm.Type
	if fn.Signature.Results() == nil {
		retType = c.ctx.VoidType()
	} else if fn.Signature.Results().Len() == 1 {
		retType = c.getLLVMType(fn.Signature.Results().At(0).Type())
	} else {
		results := make([]llvm.Type, 0, fn.Signature.Results().Len())
		for i := 0; i < fn.Signature.Results().Len(); i++ {
			results = append(results, c.getLLVMType(fn.Signature.Results().At(i).Type()))
		}
		retType = c.ctx.StructType(results, false)
	}

	var paramInfos []paramInfo
	for _, param := range getParams(fn.Signature) {
		paramType := c.getLLVMType(param.Type())
		paramFragmentInfos := c.expandFormalParamType(paramType, param.Name(), param.Type())
		paramInfos = append(paramInfos, paramFragmentInfos...)
	}

	// Add an extra parameter as the function context. This context is used in
	// closures and bound methods, but should be optimized away when not used.
	if !info.exported {
		paramInfos = append(paramInfos, paramInfo{llvmType: c.i8ptrType, name: "context", elemSize: 0})
	}

	var paramTypes []llvm.Type
	for _, info := range paramInfos {
		paramTypes = append(paramTypes, info.llvmType)
	}

	fnType := llvm.FunctionType(retType, paramTypes, info.variadic)
	llvmFn = llvm.AddFunction(c.mod, info.linkName, fnType)
	if strings.HasPrefix(c.Triple, "wasm") {
		// C functions without prototypes like this:
		//   void foo();
		// are actually variadic functions. However, it appears that it has been
		// decided in WebAssembly that such prototype-less functions are not
		// allowed in WebAssembly.
		// In C, this can only happen when there are zero parameters, hence this
		// check here. For more information:
		// https://reviews.llvm.org/D48443
		// https://github.com/WebAssembly/tool-conventions/issues/16
		if info.variadic && len(fn.Params) == 0 {
			attr := c.ctx.CreateStringAttribute("no-prototype", "")
			llvmFn.AddFunctionAttr(attr)
		}
	}
	c.addStandardDeclaredAttributes(llvmFn)

	dereferenceableOrNullKind := llvm.AttributeKindID("dereferenceable_or_null")
	for i, info := range paramInfos {
		if info.elemSize != 0 {
			dereferenceableOrNull := c.ctx.CreateEnumAttribute(dereferenceableOrNullKind, info.elemSize)
			llvmFn.AddAttributeAtIndex(i+1, dereferenceableOrNull)
		}
	}

	// Set a number of function or parameter attributes, depending on the
	// function. These functions are runtime functions that are known to have
	// certain attributes that might not be inferred by the compiler.
	switch info.linkName {
	case "abort":
		// On *nix systems, the "abort" functuion in libc is used to handle fatal panics.
		// Mark it as noreturn so LLVM can optimize away code.
		llvmFn.AddFunctionAttr(c.ctx.CreateEnumAttribute(llvm.AttributeKindID("noreturn"), 0))
	case "runtime.alloc":
		// Tell the optimizer that runtime.alloc is an allocator, meaning that it
		// returns values that are never null and never alias to an existing value.
		for _, attrName := range []string{"noalias", "nonnull"} {
			llvmFn.AddAttributeAtIndex(0, c.ctx.CreateEnumAttribute(llvm.AttributeKindID(attrName), 0))
		}
	case "runtime.sliceAppend":
		// Appending a slice will only read the to-be-appended slice, it won't
		// be modified.
		llvmFn.AddAttributeAtIndex(2, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("nocapture"), 0))
		llvmFn.AddAttributeAtIndex(2, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("readonly"), 0))
	case "runtime.sliceCopy":
		// Copying a slice won't capture any of the parameters.
		llvmFn.AddAttributeAtIndex(1, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("writeonly"), 0))
		llvmFn.AddAttributeAtIndex(1, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("nocapture"), 0))
		llvmFn.AddAttributeAtIndex(2, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("readonly"), 0))
		llvmFn.AddAttributeAtIndex(2, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("nocapture"), 0))
	case "runtime.trackPointer":
		// This function is necessary for tracking pointers on the stack in a
		// portable way (see gc_stack_portable.go). Indicate to the optimizer
		// that the only thing we'll do is read the pointer.
		llvmFn.AddAttributeAtIndex(1, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("nocapture"), 0))
		llvmFn.AddAttributeAtIndex(1, c.ctx.CreateEnumAttribute(llvm.AttributeKindID("readonly"), 0))
	case "__mulsi3", "__divmodsi4", "__udivmodsi4":
		if strings.Split(c.Triple, "-")[0] == "avr" {
			// These functions are compiler-rt/libgcc functions that are
			// currently implemented in Go. Assembly versions should appear in
			// LLVM 16 hopefully. Until then, they need to be made available to
			// the linker and the best way to do that is llvm.compiler.used.
			// I considered adding a pragma for this, but the LLVM language
			// reference explicitly says that this feature should not be exposed
			// to source languages:
			// > This is a rare construct that should only be used in rare
			// > circumstances, and should not be exposed to source languages.
			llvmutil.AppendToGlobal(c.mod, "llvm.compiler.used", llvmFn)
		}
	}

	// External/exported functions may not retain pointer values.
	// https://golang.org/cmd/cgo/#hdr-Passing_pointers
	if info.exported {
		if c.archFamily() == "wasm32" {
			// We need to add the wasm-import-module and the wasm-import-name
			// attributes.
			module := info.module
			if module == "" {
				module = "env"
			}
			llvmFn.AddFunctionAttr(c.ctx.CreateStringAttribute("wasm-import-module", module))

			name := info.importName
			if name == "" {
				name = info.linkName
			}
			llvmFn.AddFunctionAttr(c.ctx.CreateStringAttribute("wasm-import-name", name))
		}
		nocaptureKind := llvm.AttributeKindID("nocapture")
		nocapture := c.ctx.CreateEnumAttribute(nocaptureKind, 0)
		for i, typ := range paramTypes {
			if typ.TypeKind() == llvm.PointerTypeKind {
				llvmFn.AddAttributeAtIndex(i+1, nocapture)
			}
		}
	}

	// Synthetic functions are functions that do not appear in the source code,
	// they are artificially constructed. Usually they are wrapper functions
	// that are not referenced anywhere except in a SSA call instruction so
	// should be created right away.
	// The exception is the package initializer, which does appear in the
	// *ssa.Package members and so shouldn't be created here.
	if fn.Synthetic != "" && fn.Synthetic != "package initializer" && fn.Synthetic != "generic function" {
		irbuilder := c.ctx.NewBuilder()
		b := newBuilder(c, irbuilder, fn)
		b.createFunction()
		irbuilder.Dispose()
		llvmFn.SetLinkage(llvm.LinkOnceODRLinkage)
		llvmFn.SetUnnamedAddr(true)
	}

	return fnType, llvmFn
}

// getFunctionInfo returns information about a function that is not directly
// present in *ssa.Function, such as the link name and whether it should be
// exported.
func (c *compilerContext) getFunctionInfo(f *ssa.Function) functionInfo {
	info := functionInfo{
		// Pick the default linkName.
		linkName: f.RelString(nil),
	}
	// Check for //go: pragmas, which may change the link name (among others).
	info.parsePragmas(f)
	return info
}

// parsePragmas is used by getFunctionInfo to parse function pragmas such as
// //export or //go:noinline.
func (info *functionInfo) parsePragmas(f *ssa.Function) {
	if f.Syntax() == nil {
		return
	}
	if decl, ok := f.Syntax().(*ast.FuncDecl); ok && decl.Doc != nil {

		// Our importName for a wasm module (if we are compiling to wasm), or llvm link name
		var importName string

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

				importName = parts[1]
				info.exported = true
			case "//go:interrupt":
				if hasUnsafeImport(f.Pkg.Pkg) {
					info.interrupt = true
				}
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
			case "//go:section":
				if len(parts) == 2 && hasUnsafeImport(f.Pkg.Pkg) {
					info.section = parts[1]
				}
			case "//go:nobounds":
				// Skip bounds checking in this function. Useful for some
				// runtime functions.
				// This is somewhat dangerous and thus only imported in packages
				// that import unsafe.
				if hasUnsafeImport(f.Pkg.Pkg) {
					info.nobounds = true
				}
			case "//go:variadic":
				// The //go:variadic pragma is emitted by the CGo preprocessing
				// pass for C variadic functions. This includes both explicit
				// (with ...) and implicit (no parameters in signature)
				// functions.
				if strings.HasPrefix(f.Name(), "C.") {
					// This prefix cannot naturally be created, it must have
					// been created as a result of CGo preprocessing.
					info.variadic = true
				}
			}
		}

		// Set the importName for our exported function if we have one
		if importName != "" {
			if info.module == "" {
				info.linkName = importName
			} else {
				// WebAssembly import
				info.importName = importName
			}
		}

	}
}

// getParams returns the function parameters, including the receiver at the
// start. This is an alternative to the Params member of *ssa.Function, which is
// not yet populated when the package has not yet been built.
func getParams(sig *types.Signature) []*types.Var {
	params := []*types.Var{}
	if sig.Recv() != nil {
		params = append(params, sig.Recv())
	}
	for i := 0; i < sig.Params().Len(); i++ {
		params = append(params, sig.Params().At(i))
	}
	return params
}

// addStandardDeclaredAttributes adds attributes that are set for any function,
// whether declared or defined.
func (c *compilerContext) addStandardDeclaredAttributes(llvmFn llvm.Value) {
	if c.SizeLevel >= 1 {
		// Set the "optsize" attribute to make slightly smaller binaries at the
		// cost of minimal performance loss (-Os in Clang).
		kind := llvm.AttributeKindID("optsize")
		attr := c.ctx.CreateEnumAttribute(kind, 0)
		llvmFn.AddFunctionAttr(attr)
	}
	if c.SizeLevel >= 2 {
		// Set the "minsize" attribute to reduce code size even further,
		// regardless of performance loss (-Oz in Clang).
		kind := llvm.AttributeKindID("minsize")
		attr := c.ctx.CreateEnumAttribute(kind, 0)
		llvmFn.AddFunctionAttr(attr)
	}
	if c.CPU != "" {
		llvmFn.AddFunctionAttr(c.ctx.CreateStringAttribute("target-cpu", c.CPU))
	}
	if c.Features != "" {
		llvmFn.AddFunctionAttr(c.ctx.CreateStringAttribute("target-features", c.Features))
	}
}

// addStandardDefinedAttributes adds the set of attributes that are added to
// every function defined by TinyGo (even thunks/wrappers), possibly depending
// on the architecture. It does not set attributes only set for declared
// functions, use addStandardDeclaredAttributes for this.
func (c *compilerContext) addStandardDefinedAttributes(llvmFn llvm.Value) {
	// TinyGo does not currently raise exceptions, so set the 'nounwind' flag.
	// This behavior matches Clang when compiling C source files.
	// It reduces binary size on Linux a little bit on non-x86_64 targets by
	// eliminating exception tables for these functions.
	llvmFn.AddFunctionAttr(c.ctx.CreateEnumAttribute(llvm.AttributeKindID("nounwind"), 0))
	if strings.Split(c.Triple, "-")[0] == "x86_64" {
		// Required by the ABI.
		if llvmutil.Major() < 15 {
			// Needed for LLVM 14 support.
			llvmFn.AddFunctionAttr(c.ctx.CreateEnumAttribute(llvm.AttributeKindID("uwtable"), 0))
		} else {
			// The uwtable has two possible values: sync (1) or async (2). We
			// use sync because we currently don't use async unwind tables.
			// For details, see: https://llvm.org/docs/LangRef.html#function-attributes
			llvmFn.AddFunctionAttr(c.ctx.CreateEnumAttribute(llvm.AttributeKindID("uwtable"), 1))
		}
	}
}

// addStandardAttribute adds all attributes added to defined functions.
func (c *compilerContext) addStandardAttributes(llvmFn llvm.Value) {
	c.addStandardDeclaredAttributes(llvmFn)
	c.addStandardDefinedAttributes(llvmFn)
}

// globalInfo contains some information about a specific global. By default,
// linkName is equal to .RelString(nil) on a global and extern is false, but for
// some symbols this is different (due to //go:extern for example).
type globalInfo struct {
	linkName string // go:extern
	extern   bool   // go:extern
	align    int    // go:align
	section  string // go:section
}

// loadASTComments loads comments on globals from the AST, for use later in the
// program. In particular, they are required for //go:extern pragmas on globals.
func (c *compilerContext) loadASTComments(pkg *loader.Package) {
	for _, file := range pkg.Files {
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
								id := pkg.Pkg.Path() + "." + name.Name
								c.astComments[id] = decl.Doc
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

		// Set alignment from the //go:align comment.
		alignment := c.targetData.ABITypeAlignment(llvmType)
		if info.align > alignment {
			alignment = info.align
		}
		if alignment <= 0 || alignment&(alignment-1) != 0 {
			// Check for power-of-two (or 0).
			// See: https://stackoverflow.com/a/108360
			c.addError(g.Pos(), "global variable alignment must be a positive power of two")
		} else {
			// Set the alignment only when it is a power of two.
			llvmGlobal.SetAlignment(alignment)
		}

		if c.Debug && !info.extern {
			// Add debug info.
			pos := c.program.Fset.Position(g.Pos())
			diglobal := c.dibuilder.CreateGlobalVariableExpression(c.difiles[pos.Filename], llvm.DIGlobalVariableExpression{
				Name:        g.RelString(nil),
				LinkageName: info.linkName,
				File:        c.getDIFile(pos.Filename),
				Line:        pos.Line,
				Type:        c.getDIType(typ),
				LocalToUnit: false,
				Expr:        c.dibuilder.CreateExpression(nil),
				AlignInBits: uint32(alignment) * 8,
			})
			llvmGlobal.AddMetadata(0, diglobal)
		}
	}
	return llvmGlobal
}

// getGlobalInfo returns some information about a specific global.
func (c *compilerContext) getGlobalInfo(g *ssa.Global) globalInfo {
	info := globalInfo{
		// Pick the default linkName.
		linkName: g.RelString(nil),
	}
	// Check for //go: pragmas, which may change the link name (among others).
	doc := c.astComments[info.linkName]
	if doc != nil {
		info.parsePragmas(doc)
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
		case "//go:section":
			if len(parts) == 2 {
				info.section = parts[1]
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
