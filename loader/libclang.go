package loader

// This file parses a fragment of C with libclang and stores the result for AST
// modification. It does not touch the AST itself.

import (
	"errors"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unsafe"
)

/*
#include <clang-c/Index.h> // if this fails, install libclang-7-dev
#include <stdlib.h>

int tinygo_clang_visitor(CXCursor c, CXCursor parent, CXClientData client_data);
*/
import "C"

var globalFileInfo *fileInfo

func (info *fileInfo) parseFragment(fragment string, cflags []string) error {
	index := C.clang_createIndex(0, 1)
	defer C.clang_disposeIndex(index)

	filenameC := C.CString("cgo-fake.c")
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

	if C.clang_getNumDiagnostics(unit) != 0 {
		return errors.New("cgo: libclang cannot parse fragment")
	}

	if globalFileInfo != nil {
		// There is a race condition here but that doesn't really matter as it
		// is a sanity check anyway.
		panic("libclang.go cannot be used concurrently yet")
	}
	globalFileInfo = info
	defer func() {
		globalFileInfo = nil
	}()

	cursor := C.clang_getTranslationUnitCursor(unit)
	C.clang_visitChildren(cursor, (*[0]byte)(unsafe.Pointer(C.tinygo_clang_visitor)), C.CXClientData(uintptr(0)))

	return nil
}

//export tinygo_clang_visitor
func tinygo_clang_visitor(c, parent C.CXCursor, client_data C.CXClientData) C.int {
	info := globalFileInfo
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
