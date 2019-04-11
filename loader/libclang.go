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

int tinygo_clang_globals_visitor(CXCursor c, CXCursor parent, CXClientData client_data);
int tinygo_clang_struct_visitor(CXCursor c, CXCursor parent, CXClientData client_data);
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
	cursor := C.clang_getTranslationUnitCursor(unit)
	C.clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_globals_visitor), C.CXClientData(ref))

	return nil
}

//export tinygo_clang_globals_visitor
func tinygo_clang_globals_visitor(c, parent C.CXCursor, client_data C.CXClientData) C.int {
	info := refMap.Get(unsafe.Pointer(client_data)).(*fileInfo)
	kind := C.clang_getCursorKind(c)
	switch kind {
	case C.CXCursor_FunctionDecl:
		name := getString(C.clang_getCursorSpelling(c))
		cursorType := C.clang_getCursorType(c)
		if C.clang_isFunctionTypeVariadic(cursorType) != 0 {
			return C.CXChildVisit_Continue // not supported
		}
		numArgs := int(C.clang_Cursor_getNumArguments(c))
		fn := &functionInfo{}
		info.functions[name] = fn
		for i := 0; i < numArgs; i++ {
			arg := C.clang_Cursor_getArgument(c, C.uint(i))
			argName := getString(C.clang_getCursorSpelling(arg))
			argType := C.clang_getArgType(cursorType, C.uint(i))
			if argName == "" {
				argName = "$" + strconv.Itoa(i)
			}
			fn.args = append(fn.args, paramInfo{
				name:     argName,
				typeExpr: info.makeASTType(argType),
			})
		}
		resultType := C.clang_getCursorResultType(c)
		if resultType.kind != C.CXType_Void {
			fn.results = &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: info.makeASTType(resultType),
					},
				},
			}
		}
	case C.CXCursor_TypedefDecl:
		typedefType := C.clang_getCursorType(c)
		name := getString(C.clang_getTypedefName(typedefType))
		underlyingType := C.clang_getTypedefDeclUnderlyingType(c)
		expr := info.makeASTType(underlyingType)
		if strings.HasPrefix(name, "_Cgo_") {
			expr := expr.(*ast.Ident)
			typeSize := C.clang_Type_getSizeOf(underlyingType)
			switch expr.Name {
			// TODO: plain char (may be signed or unsigned)
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
		name := getString(C.clang_getCursorSpelling(c))
		cursorType := C.clang_getCursorType(c)
		info.globals[name] = &globalInfo{
			typeExpr: info.makeASTType(cursorType),
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

// makeASTType return the ast.Expr for the given libclang type. In other words,
// it converts a libclang type to a type in the Go AST.
func (info *fileInfo) makeASTType(typ C.CXType) ast.Expr {
	var typeName string
	switch typ.kind {
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
		return &ast.StarExpr{
			Star: info.importCPos,
			X:    info.makeASTType(C.clang_getPointeeType(typ)),
		}
	case C.CXType_ConstantArray:
		return &ast.ArrayType{
			Lbrack: info.importCPos,
			Len: &ast.BasicLit{
				ValuePos: info.importCPos,
				Kind:     token.INT,
				Value:    strconv.FormatInt(int64(C.clang_getArraySize(typ)), 10),
			},
			Elt: info.makeASTType(C.clang_getElementType(typ)),
		}
	case C.CXType_FunctionProto:
		// Be compatible with gc, which uses the *[0]byte type for function
		// pointer types.
		// Return type [0]byte because this is a function type, not a pointer to
		// this function type.
		return &ast.ArrayType{
			Lbrack: info.importCPos,
			Len: &ast.BasicLit{
				ValuePos: info.importCPos,
				Kind:     token.INT,
				Value:    "0",
			},
			Elt: &ast.Ident{
				NamePos: info.importCPos,
				Name:    "byte",
			},
		}
	case C.CXType_Typedef:
		typedefName := getString(C.clang_getTypedefName(typ))
		return &ast.Ident{
			NamePos: info.importCPos,
			Name:    "C." + typedefName,
		}
	case C.CXType_Elaborated:
		underlying := C.clang_Type_getNamedType(typ)
		return info.makeASTType(underlying)
	case C.CXType_Record:
		cursor := C.clang_getTypeDeclaration(typ)
		switch C.clang_getCursorKind(cursor) {
		case C.CXCursor_StructDecl:
			fieldList := &ast.FieldList{
				Opening: info.importCPos,
				Closing: info.importCPos,
			}
			ref := refMap.Put(struct {
				fieldList *ast.FieldList
				info      *fileInfo
			}{fieldList, info})
			defer refMap.Remove(ref)
			C.clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_struct_visitor), C.CXClientData(uintptr(ref)))
			return &ast.StructType{
				Struct: info.importCPos,
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
		NamePos: info.importCPos,
		Name:    typeName,
	}
}

//export tinygo_clang_struct_visitor
func tinygo_clang_struct_visitor(c, parent C.CXCursor, client_data C.CXClientData) C.int {
	passed := refMap.Get(unsafe.Pointer(client_data)).(struct {
		fieldList *ast.FieldList
		info      *fileInfo
	})
	fieldList := passed.fieldList
	info := passed.info
	if C.clang_getCursorKind(c) != C.CXCursor_FieldDecl {
		panic("expected field inside cursor")
	}
	name := getString(C.clang_getCursorSpelling(c))
	typ := C.clang_getCursorType(c)
	field := &ast.Field{
		Type: info.makeASTType(typ),
	}
	field.Names = []*ast.Ident{
		&ast.Ident{
			NamePos: info.importCPos,
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
