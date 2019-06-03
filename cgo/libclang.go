package cgo

// This file parses a fragment of C with libclang and stores the result for AST
// modification. It does not touch the AST itself.

import (
	"fmt"
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
CXSourceRange tinygo_clang_getCursorExtent(GoCXCursor c);
CXTranslationUnit tinygo_clang_Cursor_getTranslationUnit(GoCXCursor c);
long long tinygo_clang_getEnumConstantDeclValue(GoCXCursor c);
CXType tinygo_clang_getEnumDeclIntegerType(GoCXCursor c);
unsigned tinygo_clang_Cursor_isBitField(GoCXCursor c);

int tinygo_clang_globals_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
int tinygo_clang_struct_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
int tinygo_clang_enum_visitor(GoCXCursor c, GoCXCursor parent, CXClientData client_data);
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

func (p *cgoPackage) parseFragment(fragment string, cflags []string, posFilename string, posLine int) {
	index := C.clang_createIndex(0, 0)
	defer C.clang_disposeIndex(index)

	// pretend to be a .c file
	filenameC := C.CString(posFilename + "!cgo.c")
	defer C.free(unsafe.Pointer(filenameC))

	// fix up error locations
	fragment = fmt.Sprintf("# %d %#v\n", posLine+1, posFilename) + fragment

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

	if numDiagnostics := int(C.clang_getNumDiagnostics(unit)); numDiagnostics != 0 {
		addDiagnostic := func(diagnostic C.CXDiagnostic) {
			spelling := getString(C.clang_getDiagnosticSpelling(diagnostic))
			severity := diagnosticSeverity[C.clang_getDiagnosticSeverity(diagnostic)]
			location := C.clang_getDiagnosticLocation(diagnostic)
			pos := p.getClangLocationPosition(location, unit)
			p.addError(pos, severity+": "+spelling)
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
		return
	}

	ref := storedRefs.Put(p)
	defer storedRefs.Remove(ref)
	cursor := C.tinygo_clang_getTranslationUnitCursor(unit)
	C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_globals_visitor), C.CXClientData(ref))
}

//export tinygo_clang_globals_visitor
func tinygo_clang_globals_visitor(c, parent C.GoCXCursor, client_data C.CXClientData) C.int {
	p := storedRefs.Get(unsafe.Pointer(client_data)).(*cgoPackage)
	kind := C.tinygo_clang_getCursorKind(c)
	pos := p.getCursorPosition(c)
	switch kind {
	case C.CXCursor_FunctionDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		if _, required := p.missingSymbols[name]; !required {
			return C.CXChildVisit_Continue
		}
		cursorType := C.tinygo_clang_getCursorType(c)
		if C.clang_isFunctionTypeVariadic(cursorType) != 0 {
			return C.CXChildVisit_Continue // not supported
		}
		numArgs := int(C.tinygo_clang_Cursor_getNumArguments(c))
		fn := &functionInfo{
			pos: pos,
		}
		p.functions[name] = fn
		for i := 0; i < numArgs; i++ {
			arg := C.tinygo_clang_Cursor_getArgument(c, C.uint(i))
			argName := getString(C.tinygo_clang_getCursorSpelling(arg))
			argType := C.clang_getArgType(cursorType, C.uint(i))
			if argName == "" {
				argName = "$" + strconv.Itoa(i)
			}
			fn.args = append(fn.args, paramInfo{
				name:     argName,
				typeExpr: p.makeASTType(argType, pos),
			})
		}
		resultType := C.tinygo_clang_getCursorResultType(c)
		if resultType.kind != C.CXType_Void {
			fn.results = &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: p.makeASTType(resultType, pos),
					},
				},
			}
		}
	case C.CXCursor_StructDecl:
		typ := C.tinygo_clang_getCursorType(c)
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		if _, required := p.missingSymbols["struct_"+name]; !required {
			return C.CXChildVisit_Continue
		}
		p.makeASTType(typ, pos)
	case C.CXCursor_TypedefDecl:
		typedefType := C.tinygo_clang_getCursorType(c)
		name := getString(C.clang_getTypedefName(typedefType))
		if _, required := p.missingSymbols[name]; !required {
			return C.CXChildVisit_Continue
		}
		p.makeASTType(typedefType, pos)
	case C.CXCursor_VarDecl:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		if _, required := p.missingSymbols[name]; !required {
			return C.CXChildVisit_Continue
		}
		cursorType := C.tinygo_clang_getCursorType(c)
		p.globals[name] = globalInfo{
			typeExpr: p.makeASTType(cursorType, pos),
			pos:      pos,
		}
	case C.CXCursor_MacroDefinition:
		name := getString(C.tinygo_clang_getCursorSpelling(c))
		if _, required := p.missingSymbols[name]; !required {
			return C.CXChildVisit_Continue
		}
		sourceRange := C.tinygo_clang_getCursorExtent(c)
		start := C.clang_getRangeStart(sourceRange)
		end := C.clang_getRangeEnd(sourceRange)
		var file, endFile C.CXFile
		var startOffset, endOffset C.unsigned
		C.clang_getExpansionLocation(start, &file, nil, nil, &startOffset)
		if file == nil {
			p.addError(pos, "internal error: could not find file where macro is defined")
			break
		}
		C.clang_getExpansionLocation(end, &endFile, nil, nil, &endOffset)
		if file != endFile {
			p.addError(pos, "internal error: expected start and end location of a macro to be in the same file")
			break
		}
		if startOffset > endOffset {
			p.addError(pos, "internal error: start offset of macro is after end offset")
			break
		}

		// read file contents and extract the relevant byte range
		tu := C.tinygo_clang_Cursor_getTranslationUnit(c)
		var size C.size_t
		sourcePtr := C.clang_getFileContents(tu, file, &size)
		if endOffset >= C.uint(size) {
			p.addError(pos, "internal error: end offset of macro lies after end of file")
			break
		}
		source := string(((*[1 << 28]byte)(unsafe.Pointer(sourcePtr)))[startOffset:endOffset:endOffset])
		if !strings.HasPrefix(source, name) {
			p.addError(pos, fmt.Sprintf("internal error: expected macro value to start with %#v, got %#v", name, source))
			break
		}
		value := strings.TrimSpace(source[len(name):])
		for len(value) != 0 && value[0] == '(' && value[len(value)-1] == ')' {
			value = strings.TrimSpace(value[1 : len(value)-1])
		}
		if len(value) == 0 {
			// Pretend it doesn't exist at all.
			return C.CXChildVisit_Continue
		}
		// For information about integer literals:
		// https://en.cppreference.com/w/cpp/language/integer_literal
		if value[0] == '"' {
			// string constant
			p.constants[name] = constantInfo{&ast.BasicLit{pos, token.STRING, value}, pos}
			return C.CXChildVisit_Continue
		}
		if value[0] == '\'' {
			// char constant
			p.constants[name] = constantInfo{&ast.BasicLit{pos, token.CHAR, value}, pos}
			return C.CXChildVisit_Continue
		}
		// assume it's a number (int or float)
		value = strings.Replace(value, "'", "", -1) // remove ' chars
		value = strings.TrimRight(value, "lu")      // remove llu suffixes etc.
		// find the first non-number
		nonnum := byte(0)
		for i := 0; i < len(value); i++ {
			if value[i] < '0' || value[i] > '9' {
				nonnum = value[i]
				break
			}
		}
		// determine number type based on the first non-number
		switch nonnum {
		case 0:
			// no non-number found, must be an integer
			p.constants[name] = constantInfo{&ast.BasicLit{pos, token.INT, value}, pos}
		case 'x', 'X':
			// hex integer constant
			// TODO: may also be a floating point number per C++17.
			p.constants[name] = constantInfo{&ast.BasicLit{pos, token.INT, value}, pos}
		case '.', 'e':
			// float constant
			value = strings.TrimRight(value, "fFlL")
			p.constants[name] = constantInfo{&ast.BasicLit{pos, token.FLOAT, value}, pos}
		default:
			// unknown type, ignore
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

// addError is a utility function to add an error to the list of errors.
func (p *cgoPackage) addError(pos token.Pos, msg string) {
	position := p.fset.PositionFor(pos, true)
	if filepath.IsAbs(position.Filename) {
		// Relative paths for readability, like other Go parser errors.
		relpath, err := filepath.Rel(p.dir, position.Filename)
		if err == nil {
			position.Filename = relpath
		}
	}
	p.errors = append(p.errors, scanner.Error{
		Pos: position,
		Msg: msg,
	})
}

// makeASTType return the ast.Expr for the given libclang type. In other words,
// it converts a libclang type to a type in the Go AST.
func (p *cgoPackage) makeASTType(typ C.CXType, pos token.Pos) ast.Expr {
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
			X:    p.makeASTType(pointeeType, pos),
		}
	case C.CXType_ConstantArray:
		return &ast.ArrayType{
			Lbrack: pos,
			Len: &ast.BasicLit{
				ValuePos: pos,
				Kind:     token.INT,
				Value:    strconv.FormatInt(int64(C.clang_getArraySize(typ)), 10),
			},
			Elt: p.makeASTType(C.clang_getElementType(typ), pos),
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
		if _, ok := p.typedefs[name]; !ok {
			p.typedefs[name] = nil // don't recurse
			c := C.tinygo_clang_getTypeDeclaration(typ)
			underlyingType := C.tinygo_clang_getTypedefDeclUnderlyingType(c)
			expr := p.makeASTType(underlyingType, pos)
			if strings.HasPrefix(name, "_Cgo_") {
				expr := expr.(*ast.Ident)
				typeSize := C.clang_Type_getSizeOf(underlyingType)
				switch expr.Name {
				case "C.char":
					if typeSize != 1 {
						// This happens for some very special purpose architectures
						// (DSPs etc.) that are not currently targeted.
						// https://www.embecosm.com/2017/04/18/non-8-bit-char-support-in-clang-and-llvm/
						p.addError(pos, fmt.Sprintf("unknown char width: %d", typeSize))
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
			p.typedefs[name] = &typedefInfo{
				typeExpr: expr,
				pos:      pos,
			}
		}
		return &ast.Ident{
			NamePos: pos,
			Name:    "C." + name,
		}
	case C.CXType_Elaborated:
		underlying := C.clang_Type_getNamedType(typ)
		switch underlying.kind {
		case C.CXType_Record:
			return p.makeASTType(underlying, pos)
		case C.CXType_Enum:
			return p.makeASTType(underlying, pos)
		default:
			typeKindSpelling := getString(C.clang_getTypeKindSpelling(underlying.kind))
			p.addError(pos, fmt.Sprintf("unknown elaborated type (libclang type kind %s)", typeKindSpelling))
			typeName = "<unknown>"
		}
	case C.CXType_Record:
		cursor := C.tinygo_clang_getTypeDeclaration(typ)
		name := getString(C.tinygo_clang_getCursorSpelling(cursor))
		if name == "" {
			// Anonymous record, probably inside a typedef.
			typeExpr, bitfieldList := p.makeASTRecordType(cursor, pos)
			if bitfieldList != nil {
				// This struct has bitfields, so we have to declare it as a
				// named type (for bitfield getters/setters to work).
				p.anonStructNum++
				cgoName := "struct_" + strconv.Itoa(p.anonStructNum)
				p.elaboratedTypes[cgoName] = &elaboratedTypeInfo{
					typeExpr:  typeExpr,
					pos:       pos,
					bitfields: bitfieldList,
				}
				return &ast.Ident{
					NamePos: pos,
					Name:    "C." + cgoName,
				}
			}
			return typeExpr
		} else {
			var cgoName string
			switch C.tinygo_clang_getCursorKind(cursor) {
			case C.CXCursor_StructDecl:
				cgoName = "struct_" + name
			case C.CXCursor_UnionDecl:
				cgoName = "union_" + name
			default:
				// makeASTRecordType will create an appropriate error.
				cgoName = "record_" + name
			}
			if _, ok := p.elaboratedTypes[cgoName]; !ok {
				p.elaboratedTypes[cgoName] = nil // predeclare (to avoid endless recursion)
				typeExpr, bitfieldList := p.makeASTRecordType(cursor, pos)
				p.elaboratedTypes[cgoName] = &elaboratedTypeInfo{
					typeExpr:  typeExpr,
					pos:       pos,
					bitfields: bitfieldList,
				}
			}
			return &ast.Ident{
				NamePos: pos,
				Name:    "C." + cgoName,
			}
		}
	case C.CXType_Enum:
		cursor := C.tinygo_clang_getTypeDeclaration(typ)
		name := getString(C.tinygo_clang_getCursorSpelling(cursor))
		underlying := C.tinygo_clang_getEnumDeclIntegerType(cursor)
		if name == "" {
			// anonymous enum
			ref := storedRefs.Put(p)
			defer storedRefs.Remove(ref)
			C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_enum_visitor), C.CXClientData(ref))
			return p.makeASTType(underlying, pos)
		} else {
			// named enum
			if _, ok := p.enums[name]; !ok {
				ref := storedRefs.Put(p)
				defer storedRefs.Remove(ref)
				C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_enum_visitor), C.CXClientData(ref))
				p.enums[name] = enumInfo{
					typeExpr: p.makeASTType(underlying, pos),
					pos:      pos,
				}
			}
			return &ast.Ident{
				NamePos: pos,
				Name:    "C.enum_" + name,
			}
		}
	}
	if typeName == "" {
		// Report this as an error.
		typeSpelling := getString(C.clang_getTypeSpelling(typ))
		typeKindSpelling := getString(C.clang_getTypeKindSpelling(typ.kind))
		p.addError(pos, fmt.Sprintf("unknown C type: %v (libclang type kind %s)", typeSpelling, typeKindSpelling))
		typeName = "C.<unknown>"
	}
	return &ast.Ident{
		NamePos: pos,
		Name:    typeName,
	}
}

// makeASTRecordType parses a C record (struct or union) and translates it into
// a Go struct type. Unions are implemented by setting the first field to a
// zero-lengt "C union" field, which cannot be written in Go directly.
func (p *cgoPackage) makeASTRecordType(cursor C.GoCXCursor, pos token.Pos) (*ast.StructType, []bitfieldInfo) {
	fieldList := &ast.FieldList{
		Opening: pos,
		Closing: pos,
	}
	var bitfieldList []bitfieldInfo
	inBitfield := false
	bitfieldNum := 0
	ref := storedRefs.Put(struct {
		fieldList    *ast.FieldList
		pkg          *cgoPackage
		inBitfield   *bool
		bitfieldNum  *int
		bitfieldList *[]bitfieldInfo
	}{fieldList, p, &inBitfield, &bitfieldNum, &bitfieldList})
	defer storedRefs.Remove(ref)
	C.tinygo_clang_visitChildren(cursor, C.CXCursorVisitor(C.tinygo_clang_struct_visitor), C.CXClientData(ref))
	switch C.tinygo_clang_getCursorKind(cursor) {
	case C.CXCursor_StructDecl:
		return &ast.StructType{
			Struct: pos,
			Fields: fieldList,
		}, bitfieldList
	case C.CXCursor_UnionDecl:
		if bitfieldList != nil {
			// This is valid C... but please don't do this.
			p.addError(pos, "bitfield in a union is not supported")
		}
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
		}, bitfieldList
	default:
		cursorKind := C.tinygo_clang_getCursorKind(cursor)
		cursorKindSpelling := getString(C.clang_getCursorKindSpelling(cursorKind))
		p.addError(pos, fmt.Sprintf("expected StructDecl or UnionDecl, not %s", cursorKindSpelling))
		return &ast.StructType{
			Struct: pos,
		}, nil
	}
}

//export tinygo_clang_struct_visitor
func tinygo_clang_struct_visitor(c, parent C.GoCXCursor, client_data C.CXClientData) C.int {
	passed := storedRefs.Get(unsafe.Pointer(client_data)).(struct {
		fieldList    *ast.FieldList
		pkg          *cgoPackage
		inBitfield   *bool
		bitfieldNum  *int
		bitfieldList *[]bitfieldInfo
	})
	fieldList := passed.fieldList
	p := passed.pkg
	inBitfield := passed.inBitfield
	bitfieldNum := passed.bitfieldNum
	bitfieldList := passed.bitfieldList
	pos := p.getCursorPosition(c)
	if cursorKind := C.tinygo_clang_getCursorKind(c); cursorKind != C.CXCursor_FieldDecl {
		cursorKindSpelling := getString(C.clang_getCursorKindSpelling(cursorKind))
		p.addError(pos, fmt.Sprintf("expected FieldDecl in struct or union, not %s", cursorKindSpelling))
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
		Type: p.makeASTType(typ, p.getCursorPosition(c)),
	}
	offsetof := int64(C.clang_Type_getOffsetOf(C.tinygo_clang_getCursorType(parent), C.CString(name)))
	alignOf := int64(C.clang_Type_getAlignOf(typ) * 8)
	bitfieldOffset := offsetof % alignOf
	if bitfieldOffset != 0 {
		if C.tinygo_clang_Cursor_isBitField(c) != 1 {
			p.addError(pos, "expected a bitfield")
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
		&ast.Ident{
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

//export tinygo_clang_enum_visitor
func tinygo_clang_enum_visitor(c, parent C.GoCXCursor, client_data C.CXClientData) C.int {
	p := storedRefs.Get(unsafe.Pointer(client_data)).(*cgoPackage)
	name := getString(C.tinygo_clang_getCursorSpelling(c))
	pos := p.getCursorPosition(c)
	value := C.tinygo_clang_getEnumConstantDeclValue(c)
	p.constants[name] = constantInfo{
		expr: &ast.BasicLit{pos, token.INT, strconv.FormatInt(int64(value), 10)},
		pos:  pos,
	}
	return C.CXChildVisit_Continue
}
