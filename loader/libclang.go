package loader

// This file parses a fragment of C with libclang and stores the result for AST
// modification. It does not touch the AST itself.

import (
	"go/ast"
	"go/scanner"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"
)

/*
#include <clang-c/Index.h> // if this fails, install libclang-8-dev
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
CXTranslationUnit tinygo_clang_Cursor_getTranslationUnit(GoCXCursor c);

int tinygo_clang_globals_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
int tinygo_clang_struct_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
*/
import "C"

// refMap stores references to types, used for clang_visitChildren.
var refMap RefMap

var diagnosticSeverity = [...]string{
	C.CXDiagnostic_Ignored: "ignored",
	C.CXDiagnostic_Note:    "note",
	C.CXDiagnostic_Warning: "warning",
	C.CXDiagnostic_Error:   "error",
	C.CXDiagnostic_Fatal:   "fatal",
}

func (info *fileInfo) parseFragment(fragment string, cflags []string, posFilename string, posLine int) []error {
	index := C.clang_createIndex(0, 0)
	defer C.clang_disposeIndex(index)

	filenameC := C.CString(posFilename + "!cgo.c")
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
		C.CXTranslationUnit_None,
		&unit)
	if errCode != 0 {
		panic("loader: failed to parse source with libclang")
	}
	defer C.clang_disposeTranslationUnit(unit)

	if numDiagnostics := int(C.clang_getNumDiagnostics(unit)); numDiagnostics != 0 {
		errs := []error{}
		addDiagnostic := func(diagnostic C.CXDiagnostic) {
			spelling := getString(C.clang_getDiagnosticSpelling(diagnostic))
			severity := diagnosticSeverity[C.clang_getDiagnosticSeverity(diagnostic)]
			location := C.clang_getDiagnosticLocation(diagnostic)
			var file C.CXFile
			var line C.unsigned
			var column C.unsigned
			var offset C.unsigned
			C.clang_getExpansionLocation(location, &file, &line, &column, &offset)
			filename := getString(C.clang_getFileName(file))
			if filename == posFilename+"!cgo.c" {
				// Adjust errors from the `import "C"` snippet.
				// Note: doesn't adjust filenames inside the error message
				// itself.
				filename = posFilename
				line += C.uint(posLine)
				offset = 0 // hard to calculate
			} else if filepath.IsAbs(filename) {
				// Relative paths for readability, like other Go parser errors.
				relpath, err := filepath.Rel(info.Program.Dir, filename)
				if err == nil {
					filename = relpath
				}
			}
			errs = append(errs, &scanner.Error{
				Pos: token.Position{
					Filename: filename,
					Offset:   int(offset),
					Line:     int(line),
					Column:   int(column),
				},
				Msg: severity + ": " + spelling,
			})
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
		return errs
	}

	ref := refMap.Put(info)
	defer refMap.Remove(ref)
	cursor := C.tinygo_clang_getTranslationUnitCursor(unit)
	C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_globals_visitor), C.CXClientData(ref))

	return nil
}

//export tinygo_clang_globals_visitor
func tinygo_clang_globals_visitor(c, parent C.GoCXCursor, client_data C.CXClientData) C.int {
	info := refMap.Get(unsafe.Pointer(client_data)).(*fileInfo)
	kind := C.tinygo_clang_getCursorKind(c)
	pos := info.getCursorPosition(c)
	switch kind {
	case C.CXCursor_FunctionDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		cursorType := C.tinygo_clang_getCursorType(c)
		if C.clang_isFunctionTypeVariadic(cursorType) != 0 {
			return C.CXChildVisit_Continue // not supported
		}
		numArgs := int(C.tinygo_clang_Cursor_getNumArguments(c))
		fn := &functionInfo{}
		info.functions[name] = fn
		for i := 0; i < numArgs; i++ {
			arg := C.tinygo_clang_Cursor_getArgument(c, C.uint(i))
			argName := getString(C.tinygo_clang_getCursorSpelling(arg))
			argType := C.clang_getArgType(cursorType, C.uint(i))
			if argName == "" {
				argName = "$" + strconv.Itoa(i)
			}
			fn.args = append(fn.args, paramInfo{
				name:     argName,
				typeExpr: info.makeASTType(argType, pos),
			})
		}
		resultType := C.tinygo_clang_getCursorResultType(c)
		if resultType.kind != C.CXType_Void {
			fn.results = &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: info.makeASTType(resultType, pos),
					},
				},
			}
		}
	case C.CXCursor_TypedefDecl:
		typedefType := C.tinygo_clang_getCursorType(c)
		name := getString(C.clang_getTypedefName(typedefType))
		underlyingType := C.tinygo_clang_getTypedefDeclUnderlyingType(c)
		expr := info.makeASTType(underlyingType, pos)
		if strings.HasPrefix(name, "_Cgo_") {
			expr := expr.(*ast.Ident)
			typeSize := C.clang_Type_getSizeOf(underlyingType)
			switch expr.Name {
			case "C.char":
				if typeSize != 1 {
					// This happens for some very special purpose architectures
					// (DSPs etc.) that are not currently targeted.
					// https://www.embecosm.com/2017/04/18/non-8-bit-char-support-in-clang-and-llvm/
					panic("unknown char width")
				}
				switch underlyingType.kind {
				case C.CXType_Char_S:
					expr.Name = "int8"
				case C.CXType_Char_U:
					expr.Name = "uint8"
				}
			case "C.schar", "C.short", "C.int", "C.long", "C.longlong":
				switch typeSize {
				case 1:
					expr.Name = "int8"
				case 2:
					expr.Name = "int16"
				case 4:
					expr.Name = "int32"
				case 8:
					expr.Name = "int64"
				}
			case "C.uchar", "C.ushort", "C.uint", "C.ulong", "C.ulonglong":
				switch typeSize {
				case 1:
					expr.Name = "uint8"
				case 2:
					expr.Name = "uint16"
				case 4:
					expr.Name = "uint32"
				case 8:
					expr.Name = "uint64"
				}
			}
		}
		info.typedefs[name] = &typedefInfo{
			typeExpr: expr,
		}
	case C.CXCursor_VarDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		cursorType := C.tinygo_clang_getCursorType(c)
		info.globals[name] = &globalInfo{
			typeExpr: info.makeASTType(cursorType, pos),
		}
	}
	return C.CXChildVisit_Continue
}

func getString(clangString C.CXString) (s string) {
	rawString := C.clang_getCString(clangString)
	s = C.GoString(rawString)
	C.clang_disposeString(clangString)
	return
}

// getCursorPosition returns a usable token.Pos from a libclang cursor. If the
// file for this cursor has not been seen before, it is read from libclang
// (which already has the file in memory) and added to the token.FileSet.
func (info *fileInfo) getCursorPosition(cursor C.GoCXCursor) token.Pos {
	location := C.tinygo_clang_getCursorLocation(cursor)
	var file C.CXFile
	var line C.unsigned
	var column C.unsigned
	var offset C.unsigned
	C.clang_getExpansionLocation(location, &file, &line, &column, &offset)
	if line == 0 {
		// Invalid token.
		return token.NoPos
	}
	filename := getString(C.clang_getFileName(file))
	if _, ok := info.tokenFiles[filename]; !ok {
		// File has not been seen before in this package, add line information
		// now by reading the file from libclang.
		tu := C.tinygo_clang_Cursor_getTranslationUnit(cursor)
		var size C.size_t
		sourcePtr := C.clang_getFileContents(tu, file, &size)
		source := ((*[1 << 28]byte)(unsafe.Pointer(sourcePtr)))[:size:size]
		lines := []int{0}
		for i := 0; i < len(source)-1; i++ {
			if source[i] == '\n' {
				lines = append(lines, i+1)
			}
		}
		f := info.fset.AddFile(filename, -1, int(size))
		f.SetLines(lines)
		info.tokenFiles[filename] = f
	}
	return info.tokenFiles[filename].Pos(int(offset))
}

// makeASTType return the ast.Expr for the given libclang type. In other words,
// it converts a libclang type to a type in the Go AST.
func (info *fileInfo) makeASTType(typ C.CXType, pos token.Pos) ast.Expr {
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
			X:    info.makeASTType(pointeeType, pos),
		}
	case C.CXType_ConstantArray:
		return &ast.ArrayType{
			Lbrack: pos,
			Len: &ast.BasicLit{
				ValuePos: pos,
				Kind:     token.INT,
				Value:    strconv.FormatInt(int64(C.clang_getArraySize(typ)), 10),
			},
			Elt: info.makeASTType(C.clang_getElementType(typ), pos),
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
		typedefName := getString(C.clang_getTypedefName(typ))
		return &ast.Ident{
			NamePos: pos,
			Name:    "C." + typedefName,
		}
	case C.CXType_Elaborated:
		underlying := C.clang_Type_getNamedType(typ)
		switch underlying.kind {
		case C.CXType_Record:
			cursor := C.tinygo_clang_getTypeDeclaration(typ)
			name := getString(C.tinygo_clang_getCursorSpelling(cursor))
			// It is possible that this is a recursive definition, for example
			// in linked lists (structs contain a pointer to the next element
			// of the same type). If the name exists in info.elaboratedTypes,
			// it is being processed, although it may not be fully defined yet.
			if _, ok := info.elaboratedTypes[name]; !ok {
				info.elaboratedTypes[name] = nil // predeclare (to avoid endless recursion)
				info.elaboratedTypes[name] = info.makeASTType(underlying, info.getCursorPosition(cursor))
			}
			return &ast.Ident{
				NamePos: pos,
				Name:    "C.struct_" + name,
			}
		default:
			panic("unknown elaborated type")
		}
	case C.CXType_Record:
		cursor := C.tinygo_clang_getTypeDeclaration(typ)
		fieldList := &ast.FieldList{
			Opening: pos,
			Closing: pos,
		}
		ref := refMap.Put(struct {
			fieldList *ast.FieldList
			info      *fileInfo
		}{fieldList, info})
		defer refMap.Remove(ref)
		C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_struct_visitor), C.CXClientData(ref))
		switch C.tinygo_clang_getCursorKind(cursor) {
		case C.CXCursor_StructDecl:
			return &ast.StructType{
				Struct: pos,
				Fields: fieldList,
			}
		case C.CXCursor_UnionDecl:
			if len(fieldList.List) > 1 {
				// Insert a special field at the front (of zero width) as a
				// marker that this is struct is actually a union. This is done
				// by giving the field a name that cannot be expressed directly
				// in Go.
				// Other parts of the compiler look at the first element in a
				// struct (of size > 2) to know whether this is a union.
				// Note that we don't have to insert it for single-element
				// unions as they're basically equivalent to a struct.
				unionMarker := &ast.Field{
					Type: &ast.StructType{
						Struct: pos,
					},
				}
				unionMarker.Names = []*ast.Ident{
					&ast.Ident{
						NamePos: pos,
						Name:    "C union",
						Obj: &ast.Object{
							Kind: ast.Var,
							Name: "C union",
							Decl: unionMarker,
						},
					},
				}
				fieldList.List = append([]*ast.Field{unionMarker}, fieldList.List...)
			}
			return &ast.StructType{
				Struct: pos,
				Fields: fieldList,
			}
		}
	}
	if typeName == "" {
		// Fallback, probably incorrect but at least the error points to an odd
		// type name.
		typeName = "C." + getString(C.clang_getTypeSpelling(typ))
	}
	return &ast.Ident{
		NamePos: pos,
		Name:    typeName,
	}
}

//export tinygo_clang_struct_visitor
func tinygo_clang_struct_visitor(c, parent C.GoCXCursor, client_data C.CXClientData) C.int {
	passed := refMap.Get(unsafe.Pointer(client_data)).(struct {
		fieldList *ast.FieldList
		info      *fileInfo
	})
	fieldList := passed.fieldList
	info := passed.info
	if C.tinygo_clang_getCursorKind(c) != C.CXCursor_FieldDecl {
		panic("expected field inside cursor")
	}
	name := getString(C.tinygo_clang_getCursorSpelling(c))
	typ := C.tinygo_clang_getCursorType(c)
	field := &ast.Field{
		Type: info.makeASTType(typ, info.getCursorPosition(c)),
	}
	field.Names = []*ast.Ident{
		&ast.Ident{
			NamePos: info.getCursorPosition(c),
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
