package cgo

// This file parses a fragment of C with libclang and stores the result for AST
// modification. It does not touch the AST itself.

import (
	"crypto/sha512"
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"tinygo.org/x/go-llvm"
)

/*
#include <clang-c/Index.h> // If this fails, libclang headers aren't available. Please take a look here: https://tinygo.org/docs/guides/build/
#include <llvm/Config/llvm-config.h>
#include <stdlib.h>
#include <stdint.h>

// This struct should be ABI-compatible on all platforms (uintptr_t has the same
// alignment etc. as void*) but does not include void* pointers that are not
// always real pointers.
// The Go garbage collector assumes that all non-nil pointer-typed integers are
// actually pointers. This is not always true, as data[1] often contains 0x1,
// which is clearly not a valid pointer. Usually the GC won't catch this issue,
// but occasionally it will leading to a crash with a vague error message.
typedef struct {
	enum CXCursorKind kind;
	int xdata;
	uintptr_t data[3];
} GoCXCursor;

// Forwarding functions. They are implemented in libclang_stubs.c and forward to
// the real functions without doing anything else, thus they are entirely
// compatible with the versions without tinygo_ prefix. The only difference is
// the CXCursor type, which has been replaced with GoCXCursor.
GoCXCursor tinygo_clang_getTranslationUnitCursor(CXTranslationUnit tu);
unsigned tinygo_clang_visitChildren(GoCXCursor parent, CXCursorVisitor visitor, CXClientData client_data);
CXString tinygo_clang_getCursorSpelling(GoCXCursor c);
enum CXCursorKind tinygo_clang_getCursorKind(GoCXCursor c);
CXType tinygo_clang_getCursorType(GoCXCursor c);
GoCXCursor tinygo_clang_getTypeDeclaration(CXType t);
CXType tinygo_clang_getTypedefDeclUnderlyingType(GoCXCursor c);
CXType tinygo_clang_getCursorResultType(GoCXCursor c);
int tinygo_clang_Cursor_getNumArguments(GoCXCursor c);
GoCXCursor tinygo_clang_Cursor_getArgument(GoCXCursor c, unsigned i);
CXSourceLocation tinygo_clang_getCursorLocation(GoCXCursor c);
CXSourceRange tinygo_clang_getCursorExtent(GoCXCursor c);
CXTranslationUnit tinygo_clang_Cursor_getTranslationUnit(GoCXCursor c);
long long tinygo_clang_getEnumConstantDeclValue(GoCXCursor c);
CXType tinygo_clang_getEnumDeclIntegerType(GoCXCursor c);
unsigned tinygo_clang_Cursor_isBitField(GoCXCursor c);

int tinygo_clang_globals_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
int tinygo_clang_struct_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
int tinygo_clang_enum_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
void tinygo_clang_inclusion_visitor(CXFile included_file, CXSourceLocation *inclusion_stack, unsigned include_len, CXClientData client_data);
*/
import "C"

// storedRefs stores references to types, used for clang_visitChildren.
var storedRefs refMap

var diagnosticSeverity = [...]string{
	C.CXDiagnostic_Ignored: "ignored",
	C.CXDiagnostic_Note:    "note",
	C.CXDiagnostic_Warning: "warning",
	C.CXDiagnostic_Error:   "error",
	C.CXDiagnostic_Fatal:   "fatal",
}

// Alias so that cgo.go (which doesn't import Clang related stuff and is in
// theory decoupled from Clang) can also use this type.
type clangCursor = C.GoCXCursor

func init() {
	// Check that we haven't messed up LLVM versioning.
	// This can happen when llvm_config_*.go files in either this or the
	// tinygo.org/x/go-llvm packages is incorrect. It should not ever happen
	// with byollvm.
	if C.LLVM_VERSION_STRING != llvm.Version {
		panic("incorrect build: using LLVM version " + llvm.Version + " in the tinygo.org/x/llvm package, and version " + C.LLVM_VERSION_STRING + " in the ./cgo package")
	}
}

func (f *cgoFile) readNames(fragment string, cflags []string, filename string, callback func(map[string]clangCursor)) {
	index := C.clang_createIndex(0, 0)
	defer C.clang_disposeIndex(index)

	// pretend to be a .c file
	filenameC := C.CString(filename + "!cgo.c")
	defer C.free(unsafe.Pointer(filenameC))

	fragmentC := C.CString(fragment)
	defer C.free(unsafe.Pointer(fragmentC))

	unsavedFile := C.struct_CXUnsavedFile{
		Filename: filenameC,
		Length:   C.ulong(len(fragment)),
		Contents: fragmentC,
	}

	// convert Go slice of strings to C array of strings.
	cmdargsC := C.malloc(C.size_t(len(cflags)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	defer C.free(cmdargsC)
	cmdargs := (*[1 << 16]*C.char)(cmdargsC)
	for i, cflag := range cflags {
		s := C.CString(cflag)
		cmdargs[i] = s
		defer C.free(unsafe.Pointer(s))
	}

	var unit C.CXTranslationUnit
	errCode := C.clang_parseTranslationUnit2(
		index,
		filenameC,
		(**C.char)(cmdargsC), C.int(len(cflags)), // command line args
		&unsavedFile, 1, // unsaved files
		C.CXTranslationUnit_DetailedPreprocessingRecord,
		&unit)
	if errCode != 0 {
		// This is probably a bug in the usage of libclang.
		panic("cgo: failed to parse source with libclang")
	}
	defer C.clang_disposeTranslationUnit(unit)

	// Report parser and type errors.
	if numDiagnostics := int(C.clang_getNumDiagnostics(unit)); numDiagnostics != 0 {
		addDiagnostic := func(diagnostic C.CXDiagnostic) {
			spelling := getString(C.clang_getDiagnosticSpelling(diagnostic))
			severity := diagnosticSeverity[C.clang_getDiagnosticSeverity(diagnostic)]
			location := C.clang_getDiagnosticLocation(diagnostic)
			pos := f.getClangLocationPosition(location, unit)
			f.addError(pos, severity+": "+spelling)
		}
		for i := 0; i < numDiagnostics; i++ {
			diagnostic := C.clang_getDiagnostic(unit, C.uint(i))
			addDiagnostic(diagnostic)

			// Child diagnostics (like notes on redefinitions).
			diagnostics := C.clang_getChildDiagnostics(diagnostic)
			for j := 0; j < int(C.clang_getNumDiagnosticsInSet(diagnostics)); j++ {
				addDiagnostic(C.clang_getDiagnosticInSet(diagnostics, C.uint(j)))
			}
		}
	}

	// Extract information required by CGo.
	ref := storedRefs.Put(f)
	defer storedRefs.Remove(ref)
	cursor := C.tinygo_clang_getTranslationUnitCursor(unit)
	C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_globals_visitor), C.CXClientData(ref))

	// Determine files read during CGo processing, for caching.
	inclusionCallback := func(includedFile C.CXFile) {
		// Get full file path.
		path := getString(C.clang_getFileName(includedFile))

		// Get contents of file (that should be in-memory).
		size := C.size_t(0)
		rawData := C.clang_getFileContents(unit, includedFile, &size)
		if rawData == nil {
			// Sanity check. This should (hopefully) never trigger.
			panic("libclang: file contents was not loaded")
		}
		data := (*[1 << 24]byte)(unsafe.Pointer(rawData))[:size]

		// Hash the contents if it isn't hashed yet.
		if _, ok := f.visitedFiles[path]; !ok {
			// already stored
			sum := sha512.Sum512_224(data)
			f.visitedFiles[path] = sum[:]
		}
	}
	inclusionCallbackRef := storedRefs.Put(inclusionCallback)
	defer storedRefs.Remove(inclusionCallbackRef)
	C.clang_getInclusions(unit, C.CXInclusionVisitor(C.tinygo_clang_inclusion_visitor), C.CXClientData(inclusionCallbackRef))

	// Do all the C AST operations inside a callback. This makes sure that
	// libclang related memory is only freed after it is not necessary anymore.
	callback(f.names)
}

// Convert the AST node under the given Clang cursor to a Go AST node and return
// it.
func (f *cgoFile) createASTNode(name string, c clangCursor) (ast.Node, *elaboratedTypeInfo) {
	kind := C.tinygo_clang_getCursorKind(c)
	pos := f.getCursorPosition(c)
	switch kind {
	case C.CXCursor_FunctionDecl:
		cursorType := C.tinygo_clang_getCursorType(c)
		numArgs := int(C.tinygo_clang_Cursor_getNumArguments(c))
		obj := &ast.Object{
			Kind: ast.Fun,
			Name: "C." + name,
		}
		args := make([]*ast.Field, numArgs)
		decl := &ast.FuncDecl{
			Doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{
						Slash: pos - 1,
						Text:  "//export " + name,
					},
				},
			},
			Name: &ast.Ident{
				NamePos: pos,
				Name:    "C." + name,
				Obj:     obj,
			},
			Type: &ast.FuncType{
				Func: pos,
				Params: &ast.FieldList{
					Opening: pos,
					List:    args,
					Closing: pos,
				},
			},
		}
		if C.clang_isFunctionTypeVariadic(cursorType) != 0 {
			decl.Doc.List = append(decl.Doc.List, &ast.Comment{
				Slash: pos - 1,
				Text:  "//go:variadic",
			})
		}
		for i := 0; i < numArgs; i++ {
			arg := C.tinygo_clang_Cursor_getArgument(c, C.uint(i))
			argName := getString(C.tinygo_clang_getCursorSpelling(arg))
			argType := C.clang_getArgType(cursorType, C.uint(i))
			if argName == "" {
				argName = "$" + strconv.Itoa(i)
			}
			args[i] = &ast.Field{
				Names: []*ast.Ident{
					{
						NamePos: pos,
						Name:    argName,
						Obj: &ast.Object{
							Kind: ast.Var,
							Name: argName,
							Decl: decl,
						},
					},
				},
				Type: f.makeDecayingASTType(argType, pos),
			}
		}
		resultType := C.tinygo_clang_getCursorResultType(c)
		if resultType.kind != C.CXType_Void {
			decl.Type.Results = &ast.FieldList{
				List: []*ast.Field{
					{
						Type: f.makeASTType(resultType, pos),
					},
				},
			}
		}
		obj.Decl = decl
		return decl, nil
	case C.CXCursor_StructDecl, C.CXCursor_UnionDecl:
		typ := f.makeASTRecordType(c, pos)
		typeName := "C." + name
		typeExpr := typ.typeExpr
		if typ.unionSize != 0 {
			// Convert to a single-field struct type.
			typeExpr = f.makeUnionField(typ)
		}
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
			Type: typeExpr,
		}
		obj.Decl = typeSpec
		return typeSpec, typ
	case C.CXCursor_TypedefDecl:
		typeName := "C." + name
		underlyingType := C.tinygo_clang_getTypedefDeclUnderlyingType(c)
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: typeName,
		}
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: pos,
				Name:    typeName,
				Obj:     obj,
			},
			Type: f.makeASTType(underlyingType, pos),
		}
		if underlyingType.kind != C.CXType_Enum {
			typeSpec.Assign = pos
		}
		obj.Decl = typeSpec
		return typeSpec, nil
	case C.CXCursor_VarDecl:
		cursorType := C.tinygo_clang_getCursorType(c)
		typeExpr := f.makeASTType(cursorType, pos)
		gen := &ast.GenDecl{
			TokPos: pos,
			Tok:    token.VAR,
			Lparen: token.NoPos,
			Rparen: token.NoPos,
			Doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{
						Slash: pos - 1,
						Text:  "//go:extern " + name,
					},
				},
			},
		}
		obj := &ast.Object{
			Kind: ast.Var,
			Name: "C." + name,
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{{
				NamePos: pos,
				Name:    "C." + name,
				Obj:     obj,
			}},
			Type: typeExpr,
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
		return gen, nil
	case C.CXCursor_MacroDefinition:
		sourceRange := C.tinygo_clang_getCursorExtent(c)
		start := C.clang_getRangeStart(sourceRange)
		end := C.clang_getRangeEnd(sourceRange)
		var file, endFile C.CXFile
		var startOffset, endOffset C.unsigned
		C.clang_getExpansionLocation(start, &file, nil, nil, &startOffset)
		if file == nil {
			f.addError(pos, "internal error: could not find file where macro is defined")
			return nil, nil
		}
		C.clang_getExpansionLocation(end, &endFile, nil, nil, &endOffset)
		if file != endFile {
			f.addError(pos, "internal error: expected start and end location of a macro to be in the same file")
			return nil, nil
		}
		if startOffset > endOffset {
			f.addError(pos, "internal error: start offset of macro is after end offset")
			return nil, nil
		}

		// read file contents and extract the relevant byte range
		tu := C.tinygo_clang_Cursor_getTranslationUnit(c)
		var size C.size_t
		sourcePtr := C.clang_getFileContents(tu, file, &size)
		if endOffset >= C.uint(size) {
			f.addError(pos, "internal error: end offset of macro lies after end of file")
			return nil, nil
		}
		source := string(((*[1 << 28]byte)(unsafe.Pointer(sourcePtr)))[startOffset:endOffset:endOffset])
		if !strings.HasPrefix(source, name) {
			f.addError(pos, fmt.Sprintf("internal error: expected macro value to start with %#v, got %#v", name, source))
			return nil, nil
		}
		value := source[len(name):]
		// Try to convert this #define into a Go constant expression.
		expr, scannerError := parseConst(pos+token.Pos(len(name)), f.fset, value)
		if scannerError != nil {
			f.errors = append(f.errors, *scannerError)
			return nil, nil
		}

		gen := &ast.GenDecl{
			TokPos: token.NoPos,
			Tok:    token.CONST,
			Lparen: token.NoPos,
			Rparen: token.NoPos,
		}
		obj := &ast.Object{
			Kind: ast.Con,
			Name: "C." + name,
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{{
				NamePos: pos,
				Name:    "C." + name,
				Obj:     obj,
			}},
			Values: []ast.Expr{expr},
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
		return gen, nil
	case C.CXCursor_EnumDecl:
		obj := &ast.Object{
			Kind: ast.Typ,
			Name: "C." + name,
		}
		underlying := C.tinygo_clang_getEnumDeclIntegerType(c)
		// TODO: gc's CGo implementation uses types such as `uint32` for enums
		// instead of types such as C.int, which are used here.
		typeSpec := &ast.TypeSpec{
			Name: &ast.Ident{
				NamePos: pos,
				Name:    "C." + name,
				Obj:     obj,
			},
			Assign: pos,
			Type:   f.makeASTType(underlying, pos),
		}
		obj.Decl = typeSpec
		return typeSpec, nil
	case C.CXCursor_EnumConstantDecl:
		value := C.tinygo_clang_getEnumConstantDeclValue(c)
		expr := &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.INT,
			Value:    strconv.FormatInt(int64(value), 10),
		}
		gen := &ast.GenDecl{
			TokPos: token.NoPos,
			Tok:    token.CONST,
			Lparen: token.NoPos,
			Rparen: token.NoPos,
		}
		obj := &ast.Object{
			Kind: ast.Con,
			Name: "C." + name,
		}
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{{
				NamePos: pos,
				Name:    "C." + name,
				Obj:     obj,
			}},
			Values: []ast.Expr{expr},
		}
		obj.Decl = valueSpec
		gen.Specs = append(gen.Specs, valueSpec)
		return gen, nil
	default:
		f.addError(pos, fmt.Sprintf("internal error: unknown cursor type: %d", kind))
		return nil, nil
	}
}

func getString(clangString C.CXString) (s string) {
	rawString := C.clang_getCString(clangString)
	s = C.GoString(rawString)
	C.clang_disposeString(clangString)
	return
}

//export tinygo_clang_globals_visitor
func tinygo_clang_globals_visitor(c, parent C.GoCXCursor, client_data C.CXClientData) C.int {
	f := storedRefs.Get(unsafe.Pointer(client_data)).(*cgoFile)
	switch C.tinygo_clang_getCursorKind(c) {
	case C.CXCursor_FunctionDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		f.names[name] = c
	case C.CXCursor_StructDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		if name != "" {
			f.names["struct_"+name] = c
		}
	case C.CXCursor_UnionDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		if name != "" {
			f.names["union_"+name] = c
		}
	case C.CXCursor_TypedefDecl:
		typedefType := C.tinygo_clang_getCursorType(c)
		name := getString(C.clang_getTypedefName(typedefType))
		f.names[name] = c
	case C.CXCursor_VarDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		f.names[name] = c
	case C.CXCursor_MacroDefinition:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		f.names[name] = c
	case C.CXCursor_EnumDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		if name != "" {
			// Named enum, which can be referenced from Go using C.enum_foo.
			f.names["enum_"+name] = c
		}
		// The enum fields are in global scope, so recurse to visit them.
		return C.CXChildVisit_Recurse
	case C.CXCursor_EnumConstantDecl:
		// We arrive here because of the "Recurse" above.
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		f.names[name] = c
	}
	return C.CXChildVisit_Continue
}

// getCursorPosition returns a usable token.Pos from a libclang cursor.
func (p *cgoPackage) getCursorPosition(cursor C.GoCXCursor) token.Pos {
	return p.getClangLocationPosition(C.tinygo_clang_getCursorLocation(cursor), C.tinygo_clang_Cursor_getTranslationUnit(cursor))
}

// getClangLocationPosition returns a usable token.Pos based on a libclang
// location and translation unit. If the file for this cursor has not been seen
// before, it is read from libclang (which already has the file in memory) and
// added to the token.FileSet.
func (p *cgoPackage) getClangLocationPosition(location C.CXSourceLocation, tu C.CXTranslationUnit) token.Pos {
	var file C.CXFile
	var line C.unsigned
	var column C.unsigned
	var offset C.unsigned
	C.clang_getExpansionLocation(location, &file, &line, &column, &offset)
	if line == 0 || file == nil {
		// Invalid token.
		return token.NoPos
	}
	filename := getString(C.clang_getFileName(file))
	if _, ok := p.tokenFiles[filename]; !ok {
		// File has not been seen before in this package, add line information
		// now by reading the file from libclang.
		var size C.size_t
		sourcePtr := C.clang_getFileContents(tu, file, &size)
		source := ((*[1 << 28]byte)(unsafe.Pointer(sourcePtr)))[:size:size]
		lines := []int{0}
		for i := 0; i < len(source)-1; i++ {
			if source[i] == '\n' {
				lines = append(lines, i+1)
			}
		}
		f := p.fset.AddFile(filename, -1, int(size))
		f.SetLines(lines)
		p.tokenFiles[filename] = f
	}
	positionFile := p.tokenFiles[filename]

	// Check for alternative line/column information (set with a line directive).
	var filename2String C.CXString
	var line2 C.unsigned
	var column2 C.unsigned
	C.clang_getPresumedLocation(location, &filename2String, &line2, &column2)
	filename2 := getString(filename2String)
	if filename2 != filename || line2 != line || column2 != column {
		// The location was changed with a preprocessor directive.
		// TODO: this only works for locations that are added in order. Adding
		// line/column info to a file that already has line/column info after
		// the given offset is ignored.
		positionFile.AddLineColumnInfo(int(offset), filename2, int(line2), int(column2))
	}

	return positionFile.Pos(int(offset))
}

// addError is a utility function to add an error to the list of errors. It will
// convert the token position to a line/column position first, and call
// addErrorAt.
func (p *cgoPackage) addError(pos token.Pos, msg string) {
	p.addErrorAt(p.fset.PositionFor(pos, true), msg)
}

// addErrorAfter is like addError, but adds the text `after` to the source
// location.
func (p *cgoPackage) addErrorAfter(pos token.Pos, after, msg string) {
	position := p.fset.PositionFor(pos, true)
	lines := strings.Split(after, "\n")
	if len(lines) != 1 {
		// Adjust lines.
		// For why we can't just do pos+token.Pos(len(after)), see:
		// https://github.com/golang/go/issues/35803
		position.Line += len(lines) - 1
		position.Column = len(lines[len(lines)-1]) + 1
	} else {
		position.Column += len(after)
	}
	p.addErrorAt(position, msg)
}

// addErrorAt is a utility function to add an error to the list of errors.
func (p *cgoPackage) addErrorAt(position token.Position, msg string) {
	if filepath.IsAbs(position.Filename) {
		// Relative paths for readability, like other Go parser errors.
		relpath, err := filepath.Rel(p.currentDir, position.Filename)
		if err == nil {
			position.Filename = relpath
		}
	}
	p.errors = append(p.errors, scanner.Error{
		Pos: position,
		Msg: msg,
	})
}

// makeDecayingASTType does the same as makeASTType but takes care of decaying
// types (arrays in function parameters, etc). It is otherwise identical to
// makeASTType.
func (f *cgoFile) makeDecayingASTType(typ C.CXType, pos token.Pos) ast.Expr {
	// Strip typedefs, if any.
	underlyingType := typ
	if underlyingType.kind == C.CXType_Typedef {
		c := C.tinygo_clang_getTypeDeclaration(typ)
		underlyingType = C.tinygo_clang_getTypedefDeclUnderlyingType(c)
		// TODO: support a chain of typedefs. At the moment, it seems to get
		// stuck in an endless loop when trying to get to the most underlying
		// type.
	}
	// Check for decaying type. An example would be an array type in a
	// parameter. This declaration:
	//   void foo(char buf[6]);
	// is the same as this one:
	//   void foo(char *buf);
	// But this one:
	//   void bar(char buf[6][4]);
	// equals this:
	//   void bar(char *buf[4]);
	// so not all array dimensions should be stripped, just the first one.
	// TODO: there are more kinds of decaying types.
	if underlyingType.kind == C.CXType_ConstantArray {
		// Apply type decaying.
		pointeeType := C.clang_getElementType(underlyingType)
		return &ast.StarExpr{
			Star: pos,
			X:    f.makeASTType(pointeeType, pos),
		}
	}
	return f.makeASTType(typ, pos)
}

// makeASTType return the ast.Expr for the given libclang type. In other words,
// it converts a libclang type to a type in the Go AST.
func (f *cgoFile) makeASTType(typ C.CXType, pos token.Pos) ast.Expr {
	var typeName string
	switch typ.kind {
	case C.CXType_Char_S, C.CXType_Char_U:
		typeName = "C.char"
	case C.CXType_SChar:
		typeName = "C.schar"
	case C.CXType_UChar:
		typeName = "C.uchar"
	case C.CXType_Short:
		typeName = "C.short"
	case C.CXType_UShort:
		typeName = "C.ushort"
	case C.CXType_Int:
		typeName = "C.int"
	case C.CXType_UInt:
		typeName = "C.uint"
	case C.CXType_Long:
		typeName = "C.long"
	case C.CXType_ULong:
		typeName = "C.ulong"
	case C.CXType_LongLong:
		typeName = "C.longlong"
	case C.CXType_ULongLong:
		typeName = "C.ulonglong"
	case C.CXType_Bool:
		typeName = "bool"
	case C.CXType_Float, C.CXType_Double, C.CXType_LongDouble:
		switch C.clang_Type_getSizeOf(typ) {
		case 4:
			typeName = "float32"
		case 8:
			typeName = "float64"
		default:
			// Don't do anything, rely on the fallback code to show a somewhat
			// sensible error message like "undeclared name: C.long double".
		}
	case C.CXType_Complex:
		switch C.clang_Type_getSizeOf(typ) {
		case 8:
			typeName = "complex64"
		case 16:
			typeName = "complex128"
		}
	case C.CXType_Pointer:
		pointeeType := C.clang_getPointeeType(typ)
		if pointeeType.kind == C.CXType_Void {
			// void* type is translated to Go as unsafe.Pointer
			return &ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: pos,
					Name:    "unsafe",
				},
				Sel: &ast.Ident{
					NamePos: pos,
					Name:    "Pointer",
				},
			}
		}
		return &ast.StarExpr{
			Star: pos,
			X:    f.makeASTType(pointeeType, pos),
		}
	case C.CXType_ConstantArray:
		return &ast.ArrayType{
			Lbrack: pos,
			Len: &ast.BasicLit{
				ValuePos: pos,
				Kind:     token.INT,
				Value:    strconv.FormatInt(int64(C.clang_getArraySize(typ)), 10),
			},
			Elt: f.makeASTType(C.clang_getElementType(typ), pos),
		}
	case C.CXType_FunctionProto:
		// Be compatible with gc, which uses the *[0]byte type for function
		// pointer types.
		// Return type [0]byte because this is a function type, not a pointer to
		// this function type.
		return &ast.ArrayType{
			Lbrack: pos,
			Len: &ast.BasicLit{
				ValuePos: pos,
				Kind:     token.INT,
				Value:    "0",
			},
			Elt: &ast.Ident{
				NamePos: pos,
				Name:    "byte",
			},
		}
	case C.CXType_Typedef:
		name := getString(C.clang_getTypedefName(typ))
		c := C.tinygo_clang_getTypeDeclaration(typ)
		return &ast.Ident{
			NamePos: pos,
			Name:    f.getASTDeclName(name, c, false),
		}
	case C.CXType_Elaborated:
		underlying := C.clang_Type_getNamedType(typ)
		switch underlying.kind {
		case C.CXType_Record:
			return f.makeASTType(underlying, pos)
		case C.CXType_Enum:
			return f.makeASTType(underlying, pos)
		default:
			typeKindSpelling := getString(C.clang_getTypeKindSpelling(underlying.kind))
			f.addError(pos, fmt.Sprintf("unknown elaborated type (libclang type kind %s)", typeKindSpelling))
			typeName = "<unknown>"
		}
	case C.CXType_Record:
		cursor := C.tinygo_clang_getTypeDeclaration(typ)
		name := getString(C.tinygo_clang_getCursorSpelling(cursor))
		var cgoRecordPrefix string
		switch C.tinygo_clang_getCursorKind(cursor) {
		case C.CXCursor_StructDecl:
			cgoRecordPrefix = "struct_"
		case C.CXCursor_UnionDecl:
			cgoRecordPrefix = "union_"
		default:
			// makeASTRecordType will create an appropriate error.
			cgoRecordPrefix = "record_"
		}
		if name == "" {
			// Anonymous record, probably inside a typedef.
			clangLocation := C.tinygo_clang_getCursorLocation(cursor)
			var file C.CXFile
			var line C.unsigned
			var column C.unsigned
			C.clang_getFileLocation(clangLocation, &file, &line, &column, nil)
			location := token.Position{
				Filename: getString(C.clang_getFileName(file)),
				Line:     int(line),
				Column:   int(column),
			}
			if location.Filename == "" || location.Line == 0 {
				// Not sure when this would happen, but protect from it anyway.
				f.addError(pos, "could not find file/line information")
			}
			name = f.getUnnamedDeclName("_Ctype_"+cgoRecordPrefix+"__", location)
		} else {
			name = cgoRecordPrefix + name
		}
		return &ast.Ident{
			NamePos: pos,
			Name:    f.getASTDeclName(name, cursor, false),
		}
	case C.CXType_Enum:
		cursor := C.tinygo_clang_getTypeDeclaration(typ)
		name := getString(C.tinygo_clang_getCursorSpelling(cursor))
		if name == "" {
			name = f.getUnnamedDeclName("_Ctype_enum___", cursor)
		} else {
			name = "enum_" + name
		}
		return &ast.Ident{
			NamePos: pos,
			Name:    f.getASTDeclName(name, cursor, false),
		}
	}
	if typeName == "" {
		// Report this as an error.
		typeSpelling := getString(C.clang_getTypeSpelling(typ))
		typeKindSpelling := getString(C.clang_getTypeKindSpelling(typ.kind))
		f.addError(pos, fmt.Sprintf("unknown C type: %v (libclang type kind %s)", typeSpelling, typeKindSpelling))
		typeName = "C.<unknown>"
	}
	return &ast.Ident{
		NamePos: pos,
		Name:    typeName,
	}
}

// getIntegerType returns an AST node that defines types such as C.int.
func (p *cgoPackage) getIntegerType(name string, cursor clangCursor) *ast.TypeSpec {
	pos := p.getCursorPosition(cursor)

	// Find a Go type that matches the size and signedness of the given C type.
	underlyingType := C.tinygo_clang_getTypedefDeclUnderlyingType(cursor)
	var goName string
	typeSize := C.clang_Type_getSizeOf(underlyingType)
	switch name {
	case "C.char":
		if typeSize != 1 {
			// This happens for some very special purpose architectures
			// (DSPs etc.) that are not currently targeted.
			// https://www.embecosm.com/2017/04/18/non-8-bit-char-support-in-clang-and-llvm/
			p.addError(pos, fmt.Sprintf("unknown char width: %d", typeSize))
		}
		switch underlyingType.kind {
		case C.CXType_Char_S:
			goName = "int8"
		case C.CXType_Char_U:
			goName = "uint8"
		}
	case "C.schar", "C.short", "C.int", "C.long", "C.longlong":
		switch typeSize {
		case 1:
			goName = "int8"
		case 2:
			goName = "int16"
		case 4:
			goName = "int32"
		case 8:
			goName = "int64"
		}
	case "C.uchar", "C.ushort", "C.uint", "C.ulong", "C.ulonglong":
		switch typeSize {
		case 1:
			goName = "uint8"
		case 2:
			goName = "uint16"
		case 4:
			goName = "uint32"
		case 8:
			goName = "uint64"
		}
	}

	if goName == "" { // should not happen
		p.addError(pos, "internal error: did not find Go type for C type "+name)
		goName = "int"
	}

	// Construct an *ast.TypeSpec for this type.
	obj := &ast.Object{
		Kind: ast.Typ,
		Name: name,
	}
	spec := &ast.TypeSpec{
		Name: &ast.Ident{
			NamePos: pos,
			Name:    name,
			Obj:     obj,
		},
		Type: &ast.Ident{
			NamePos: pos,
			Name:    goName,
		},
	}
	obj.Decl = spec
	return spec
}

// makeASTRecordType parses a C record (struct or union) and translates it into
// a Go struct type.
func (f *cgoFile) makeASTRecordType(cursor C.GoCXCursor, pos token.Pos) *elaboratedTypeInfo {
	fieldList := &ast.FieldList{
		Opening: pos,
		Closing: pos,
	}
	var bitfieldList []bitfieldInfo
	inBitfield := false
	bitfieldNum := 0
	ref := storedRefs.Put(struct {
		fieldList    *ast.FieldList
		file         *cgoFile
		inBitfield   *bool
		bitfieldNum  *int
		bitfieldList *[]bitfieldInfo
	}{fieldList, f, &inBitfield, &bitfieldNum, &bitfieldList})
	defer storedRefs.Remove(ref)
	C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_struct_visitor), C.CXClientData(ref))
	renameFieldKeywords(fieldList)
	switch C.tinygo_clang_getCursorKind(cursor) {
	case C.CXCursor_StructDecl:
		return &elaboratedTypeInfo{
			typeExpr: &ast.StructType{
				Struct: pos,
				Fields: fieldList,
			},
			pos:       pos,
			bitfields: bitfieldList,
		}
	case C.CXCursor_UnionDecl:
		typeInfo := &elaboratedTypeInfo{
			typeExpr: &ast.StructType{
				Struct: pos,
				Fields: fieldList,
			},
			pos:       pos,
			bitfields: bitfieldList,
		}
		if len(fieldList.List) <= 1 {
			// Useless union, treat it as a regular struct.
			return typeInfo
		}
		if bitfieldList != nil {
			// This is valid C... but please don't do this.
			f.addError(pos, "bitfield in a union is not supported")
		}
		typ := C.tinygo_clang_getCursorType(cursor)
		alignInBytes := int64(C.clang_Type_getAlignOf(typ))
		sizeInBytes := int64(C.clang_Type_getSizeOf(typ))
		if sizeInBytes == 0 {
			f.addError(pos, "zero-length union is not supported")
		}
		typeInfo.unionSize = sizeInBytes
		typeInfo.unionAlign = alignInBytes
		return typeInfo
	default:
		cursorKind := C.tinygo_clang_getCursorKind(cursor)
		cursorKindSpelling := getString(C.clang_getCursorKindSpelling(cursorKind))
		f.addError(pos, fmt.Sprintf("expected StructDecl or UnionDecl, not %s", cursorKindSpelling))
		return &elaboratedTypeInfo{
			typeExpr: &ast.StructType{
				Struct: pos,
			},
			pos: pos,
		}
	}
}

//export tinygo_clang_struct_visitor
func tinygo_clang_struct_visitor(c, parent C.GoCXCursor, client_data C.CXClientData) C.int {
	passed := storedRefs.Get(unsafe.Pointer(client_data)).(struct {
		fieldList    *ast.FieldList
		file         *cgoFile
		inBitfield   *bool
		bitfieldNum  *int
		bitfieldList *[]bitfieldInfo
	})
	fieldList := passed.fieldList
	f := passed.file
	inBitfield := passed.inBitfield
	bitfieldNum := passed.bitfieldNum
	bitfieldList := passed.bitfieldList
	pos := f.getCursorPosition(c)
	switch cursorKind := C.tinygo_clang_getCursorKind(c); cursorKind {
	case C.CXCursor_FieldDecl:
		// Expected. This is a regular field.
	case C.CXCursor_StructDecl, C.CXCursor_UnionDecl:
		// Ignore. The next field will be the struct/union itself.
		return C.CXChildVisit_Continue
	default:
		cursorKindSpelling := getString(C.clang_getCursorKindSpelling(cursorKind))
		f.addError(pos, fmt.Sprintf("expected FieldDecl in struct or union, not %s", cursorKindSpelling))
		return C.CXChildVisit_Continue
	}
	name := getString(C.tinygo_clang_getCursorSpelling(c))
	if name == "" {
		// Assume this is a bitfield of 0 bits.
		// Warning: this is not necessarily true!
		return C.CXChildVisit_Continue
	}
	typ := C.tinygo_clang_getCursorType(c)
	field := &ast.Field{
		Type: f.makeASTType(typ, f.getCursorPosition(c)),
	}
	offsetof := int64(C.clang_Type_getOffsetOf(C.tinygo_clang_getCursorType(parent), C.CString(name)))
	alignOf := int64(C.clang_Type_getAlignOf(typ) * 8)
	bitfieldOffset := offsetof % alignOf
	if bitfieldOffset != 0 {
		if C.tinygo_clang_Cursor_isBitField(c) != 1 {
			f.addError(pos, "expected a bitfield")
			return C.CXChildVisit_Continue
		}
		if !*inBitfield {
			*bitfieldNum++
		}
		bitfieldName := "__bitfield_" + strconv.Itoa(*bitfieldNum)
		prevField := fieldList.List[len(fieldList.List)-1]
		if !*inBitfield {
			// The previous element also was a bitfield, but wasn't noticed
			// then. Add it now.
			*inBitfield = true
			*bitfieldList = append(*bitfieldList, bitfieldInfo{
				field:    prevField,
				name:     prevField.Names[0].Name,
				startBit: 0,
				pos:      prevField.Names[0].NamePos,
			})
			prevField.Names[0].Name = bitfieldName
			prevField.Names[0].Obj.Name = bitfieldName
		}
		prevBitfield := &(*bitfieldList)[len(*bitfieldList)-1]
		prevBitfield.endBit = bitfieldOffset
		*bitfieldList = append(*bitfieldList, bitfieldInfo{
			field:    prevField,
			name:     name,
			startBit: bitfieldOffset,
			pos:      pos,
		})
		return C.CXChildVisit_Continue
	}
	*inBitfield = false
	field.Names = []*ast.Ident{
		{
			NamePos: pos,
			Name:    name,
			Obj: &ast.Object{
				Kind: ast.Var,
				Name: name,
				Decl: field,
			},
		},
	}
	fieldList.List = append(fieldList.List, field)
	return C.CXChildVisit_Continue
}

//export tinygo_clang_inclusion_visitor
func tinygo_clang_inclusion_visitor(includedFile C.CXFile, inclusionStack *C.CXSourceLocation, includeLen C.unsigned, clientData C.CXClientData) {
	callback := storedRefs.Get(unsafe.Pointer(clientData)).(func(C.CXFile))
	callback(includedFile)
}
