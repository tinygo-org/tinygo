package transform

import (
	"fmt"
	"regexp"
	"strings"
	"tinygo.org/x/go-llvm"
)

type idMiner struct {
	complitReg *regexp.Regexp
	v          llvm.Value
}

func makeIdMiner() idMiner {
	return idMiner{complitReg: regexp.MustCompile(`^complit\d*$`)}
}

func (this *idMiner) get(v llvm.Value) string {
	if !v.IsACallInst().IsNil() {
		getAttr := func(key string) (value string, found bool) {
			attr := v.GetCallSiteStringAttribute(-1, key)
			if attr.IsNil() {
				return "", false
			}
			value = attr.GetStringValue()
			return value, len(value) > 0
		}
		typeAttr, hasType := getAttr("tinygo-alloc-type")
		nameAttr, hasName := getAttr("tinygo-alloc-name")
		typeAttr, _ = strings.CutPrefix(typeAttr, "*")
		if nameAttr == "entry" && v.Name() == "makeslice.buf" {
			nameAttr = "slice"
		}
		if hasType && hasName {
			if nameAttr == "complit" {
				return typeAttr
			}
			if nameAttr == "varargs" {
				return fmt.Sprint(nameAttr, " ", typeAttr)
			}
			if nameAttr == "makeslice" || nameAttr == "slicelit" {
				return fmt.Sprint("slice ", typeAttr)
			}
		}
		if hasName {
			if nameAttr == "complit" {
				return "composite literal"
			}
			if nameAttr == "makeslice" || nameAttr == "slicelit" {
				return "slice"
			}
			return nameAttr
		}
		if hasType {
			return typeAttr
		}
	}
	// Not all allocations are attributed with "tinygo-alloc-..." falling back
	// to LLVM IR mining.
	this.v = v
	id := v.Name()
	if len(id) == 0 {
		id = this.lookForAllocatedType()
	}
	var handled bool
	id, handled = this.handleCompositLiteral(id)
	if handled {
		return id
	}
	if id == "FloatType" {
		// When tv.AllocatedType() gives "FloatType" it often means packing something into `interface`.
		id, handled = this.lookForInterfaceWrapping(id)
		if handled {
			return id
		}
	}
	if len(id) == 0 {
		id = "unidentified"
	}
	return id
}

func (this *idMiner) lookForAllocatedType() string {
	if !this.v.IsACallInst().IsNil() {
		return this.v.AllocatedType().String()
	}
	return ""
}

func (this *idMiner) handleCompositLiteral(id string) (newId string, handled bool) {
	if this.complitReg.MatchString(id) {
		// Handle case like: c := scaleVector3(&vector3{4, 5, 6}, 0.5)
		/*
		 * 1. Find the next "call", where %complit is passed as parameter.
		 * 2. Produce ID: "Arg 0 of main.scaleVector3() call" // escapes at ...
		 */
		n := this.v
		for _ = range 64 {
			n = llvm.NextInstruction(n)
			if n.IsNil() {
				break
			}
			if call := n.IsACallInst(); !call.IsNil() {
				args := getArgs(call)
				for idx, arg := range args {
					if arg.arg.Name() == id {
						return getId(call, idx, &arg), true
					}
				}
			}
		}
	}
	return id, false
}

// TODO: make a testcase where a function which have multiple escaping interface and not interface values,
// is called and check that the arg indicies displayed correctly
func (this *idMiner) lookForInterfaceWrapping(id string) (newId string, handled bool) {
	/* --- The costumError -> error, struct -> interface ---
	     * pattern:
	     * %reg = alloc
	     * store reg/var -> ptr %reg	// write to struct, array or slice
	     * call with a (*, ptr nonnull @"reflect/types.type:...", ptr %reg arg, *) arg pair
		 * - Error message: "Arg x of fnName() call with type array:3:basic:int32" + " escapes at ..."
	     * or
	     * insertvalue %runtime._interface with ptr %reg param and ptr @"reflect/types.type:named:main.theError" (name of struct)
		 * - @"reflect/types.type:array:3:basic:int32": [3]int32
		 * - Error message: "array:3:basic:int32" + " escapes at ..."
		 *
	*/
	ptrToAllocated := this.v
	n := this.v
	for _ = range 64 {
		n = llvm.NextInstruction(n)
		if n.IsNil() {
			break
		}
		if !n.IsACallInst().IsNil() {
			args := getArgs(n)
			for idx, arg := range args {
				if len(arg.descTypeName) > 0 && arg.arg == ptrToAllocated {
					return getId(n, idx, &arg), true
				}
			}
		} else if !n.IsAInsertValueInst().IsNil() { // interface wrapping
			insertedValue := n.Operand(1)
			if insertedValue == ptrToAllocated {
				typeName, found := getTypeNameFromInsertValue(n)
				if found {
					return typeName, true
				}
			}
		}
	}
	// %10 = call align 4 dereferenceable(16) ptr @runtime.alloc(i32 16, ptr null, ptr undef) #1, !dbg !193
	// As a fallback produce a message something like: "16 bytes escapes..."
	if !this.v.IsACallInst().IsNil() {
		if cv := this.v.CalledValue(); cv.Name() == "runtime.alloc" && this.v.OperandsCount() > 0 {
			allocSizeArg := this.v.Operand(0)
			allocSize := allocSizeArg.ZExtValue()
			return fmt.Sprint(allocSize, " bytes"), true
		}
	}
	return id, false
}

type arg struct {
	descTypeName string	// It is set when arg is an interface arg,
	// this is the type name which is wrapped by the interface.
	name string // sometimes the name is available
	arg  llvm.Value
}

func getId(callInst llvm.Value, idx int, a *arg) string {
	ta := len(a.descTypeName) > 0
	na := len(a.name) > 0
	fn := callInst.CalledValue().Name()
	if ta && na {
		return fmt.Sprintf("Arg %d (%s %s) of %s()", idx, a.name, a.descTypeName, fn)
	} else if ta {
		return fmt.Sprintf("Arg %d (type: %s) of %s()", idx, a.descTypeName, fn)
	} else if na {
		return fmt.Sprintf("Arg %d (name: %s) of %s()", idx, a.name, fn)
	}
	return fmt.Sprintf("Arg %d of %s()", idx, fn)
}

func getArgs(callInst llvm.Value) []arg {
	typeSuffix := ".typecode"
	valueSuffix := ".value"
	res := make([]arg, 0, 4)
	opCnt := callInst.OperandsCount()
	idx := 0
	for idx < opCnt {
		argV := callInst.Operand(idx)
		placed := false
		idx++
		var nextArgV llvm.Value
		if idx < opCnt {
			nextArgV = callInst.Operand(idx)
			if argV.Type().TypeKind() == llvm.PointerTypeKind && nextArgV.Type().TypeKind() == llvm.PointerTypeKind {
				// Check pointer pair
				ptrName := argV.Name()
				nextPtrName := nextArgV.Name()
				if len(ptrName) > 0 {
					typeName, found := removeReflexPrefix(ptrName)
					if found {
						res = append(res, arg{descTypeName: typeName, arg: nextArgV})
						placed = true
						idx++
					}
				}
				// The easier case
				if !placed {
					argName, found := strings.CutSuffix(ptrName, typeSuffix)
					if found {
						_, valueFound := strings.CutSuffix(nextPtrName, valueSuffix)
						if valueFound {
							res = append(res,
								arg{name: argName, arg: nextArgV})
							placed = true
							idx++
						}
					}
				}
				// Step back on the instruction list to look for 'extractvalue'
				if !placed {
					p := callInst
					p0Found := false
					p1Found := false
					for _ = range 64 {
						p = llvm.PrevInstruction(p)
						if p.IsNil() {
							break
						}
						if !p.IsAExtractValueInst().IsNil() {
							operand := p.Operand(0)
							tp := operand.Type()
							tpKind := tp.TypeKind()
							if tpKind == llvm.StructTypeKind && tp.StructName() == "runtime._interface" {
								p0Found = p0Found || argV == p
								p1Found = p1Found || nextArgV == p
							}
							if p0Found && p1Found {
								res = append(res,
									arg{descTypeName: "unknown", arg: nextArgV})
								placed = true
								idx++
								break
							}
						}
					}
				}
			}
		}
		if !placed {
			res = append(res, arg{arg: argV})
		}
	}
	return res
}

func removeReflexPrefix(typeName string) (newTypeName string, removed bool) {
	prefixs := []string{"reflect/types.type:named:", "reflect/types.type:"}
	for _, prefix := range prefixs {
		newTypeName, removed = strings.CutPrefix(typeName, prefix)
		if removed {
			return
		}
	}
	return
}

func getTypeNameFromInsertValue(insertValueInst llvm.Value) (typeName string, found bool) {
	baseAggregate := insertValueInst.Operand(0)

	// Since it's an inline constant structure, let's look at its elements.
	// The type descriptor is the first element of this constant struct.
	if !baseAggregate.IsAConstantStruct().IsNil() || !baseAggregate.IsAConstantAggregateZero().IsNil() {
		// Extract the first element (the GEP expression)
		gepExpr := baseAggregate.Operand(0)

		// Drill down through the GEP/Casts to find the Global Variable symbol
		for !gepExpr.IsAConstantExpr().IsNil() {
			// In a GEP or bitcast expression, Operand(0) is the source pointer
			gepExpr = gepExpr.Operand(0)
		}

		// If we successfully drilled down to the Global Variable, grab its name
		if !gepExpr.IsAGlobalVariable().IsNil() {
			typeName, found = removeReflexPrefix(gepExpr.Name())
		}
	}
	return
}

func dumpIR(hdr string, instCnt int, fstInst llvm.Value) {
	fmt.Println("---", hdr, "---")
	n := fstInst
	fmt.Println(n.String())
	for _ = range instCnt {
		n = llvm.NextInstruction(n)
		if n.IsNil() {
			break
		}
		fmt.Println(n.String())
	}
}
