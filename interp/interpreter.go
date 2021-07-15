package interp

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/go-llvm"
)

func (r *runner) run(fn *function, params []value, parentMem *memoryView, indent string) (value, memoryView, *Error) {
	mem := memoryView{r: r, parent: parentMem}
	locals := make([]value, len(fn.locals))
	r.callsExecuted++

	if time.Since(r.start) > time.Minute {
		// Running for more than a minute. This should never happen.
		return nil, mem, r.errorAt(fn.blocks[0].instructions[0], fmt.Errorf("interp: running for more than a minute, timing out (executed calls: %d)", r.callsExecuted))
	}

	// Parameters are considered a kind of local values.
	for i, param := range params {
		locals[i] = param
	}

	// Start with the first basic block and the first instruction.
	// Branch instructions may modify both bb and instIndex when branching.
	bb := fn.blocks[0]
	currentBB := 0
	lastBB := -1 // last basic block is undefined, only defined after a branch
	var operands []value
	for instIndex := 0; instIndex < len(bb.instructions); instIndex++ {
		if instIndex == 0 {
			// This is the start of a new basic block.
			// There may be PHI nodes that need to be resolved. Resolve all PHI
			// nodes before continuing with regular instructions.
			// PHI nodes need to be treated specially because they can have a
			// mutual dependency:
			//   for.loop:
			//     %a = phi i8 [ 1, %entry ], [ %b, %for.loop ]
			//     %b = phi i8 [ 3, %entry ], [ %a, %for.loop ]
			// If these PHI nodes are processed like a regular instruction, %a
			// and %b are both 3 on the second iteration of the loop because %b
			// loads the value of %a from the second iteration, while it should
			// load the value from the previous iteration. The correct behavior
			// is that these two values swap each others place on each
			// iteration.
			var phiValues []value
			var phiIndices []int
			for _, inst := range bb.phiNodes {
				var result value
				for i := 0; i < len(inst.operands); i += 2 {
					if int(inst.operands[i].(literalValue).value.(uint32)) == lastBB {
						incoming := inst.operands[i+1]
						if local, ok := incoming.(localValue); ok {
							result = locals[fn.locals[local.value]]
						} else {
							result = incoming
						}
						break
					}
				}
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"phi", inst.operands, "->", result)
				}
				if result == nil {
					panic("could not find PHI input")
				}
				phiValues = append(phiValues, result)
				phiIndices = append(phiIndices, inst.localIndex)
			}
			for i, value := range phiValues {
				locals[phiIndices[i]] = value
			}
		}

		inst := bb.instructions[instIndex]
		operands = operands[:0]
		isRuntimeInst := false
		if inst.opcode != llvm.PHI {
			for _, v := range inst.operands {
				if v, ok := v.(localValue); ok {
					if localVal := locals[fn.locals[v.value]]; localVal == nil {
						return nil, mem, r.errorAt(inst, errors.New("interp: local not defined"))
					} else {
						operands = append(operands, localVal)
						if _, ok := localVal.(localValue); ok {
							isRuntimeInst = true
						}
						continue
					}
				}
				operands = append(operands, v)
			}
		}
		if isRuntimeInst {
			err := r.runAtRuntime(fn, inst, locals, &mem, indent)
			if err != nil {
				return nil, mem, err
			}
			continue
		}
		switch inst.opcode {
		case llvm.Ret:
			if len(operands) != 0 {
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"ret", operands[0])
				}
				// Return instruction has a value to return.
				return operands[0], mem, nil
			}
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"ret")
			}
			// Return instruction doesn't return anything, it's just 'ret void'.
			return nil, mem, nil
		case llvm.Br:
			switch len(operands) {
			case 1:
				// Unconditional branch: [nextBB]
				lastBB = currentBB
				currentBB = int(operands[0].(literalValue).value.(uint32))
				bb = fn.blocks[currentBB]
				instIndex = -1 // start at 0 the next cycle
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"br", operands, "->", currentBB)
				}
			case 3:
				// Conditional branch: [cond, thenBB, elseBB]
				lastBB = currentBB
				switch operands[0].Uint() {
				case 1: // true -> thenBB
					currentBB = int(operands[1].(literalValue).value.(uint32))
				case 0: // false -> elseBB
					currentBB = int(operands[2].(literalValue).value.(uint32))
				default:
					panic("bool should be 0 or 1")
				}
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"br", operands, "->", currentBB)
				}
				bb = fn.blocks[currentBB]
				instIndex = -1 // start at 0 the next cycle
			default:
				panic("unknown operands length")
			}
		case llvm.Switch:
			// Switch statement: [value, defaultLabel, case0, label0, case1, label1, ...]
			value := operands[0].Uint()
			targetLabel := operands[1].Uint() // default label
			// Do a lazy switch by iterating over all cases.
			for i := 2; i < len(operands); i += 2 {
				if value == operands[i].Uint() {
					targetLabel = operands[i+1].Uint()
					break
				}
			}
			lastBB = currentBB
			currentBB = int(targetLabel)
			bb = fn.blocks[currentBB]
			instIndex = -1 // start at 0 the next cycle
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"switch", operands, "->", currentBB)
			}
		case llvm.Select:
			// Select is much like a ternary operator: it picks a result from
			// the second and third operand based on the boolean first operand.
			var result value
			switch operands[0].Uint() {
			case 1:
				result = operands[1]
			case 0:
				result = operands[2]
			default:
				panic("boolean must be 0 or 1")
			}
			locals[inst.localIndex] = result
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"select", operands, "->", result)
			}
		case llvm.Call:
			// A call instruction can either be a regular call or a runtime intrinsic.
			fnPtr, err := operands[0].asPointer(r)
			if err != nil {
				return nil, mem, r.errorAt(inst, err)
			}
			callFn := r.getFunction(fnPtr.llvmValue(&mem))
			switch {
			case callFn.name == "runtime.trackPointer":
				// Allocas and such are created as globals, so don't need a
				// runtime.trackPointer.
				// Unless the object is allocated at runtime for example, in
				// which case this call won't even get to this point but will
				// already be emitted in initAll.
				continue
			case strings.HasPrefix(callFn.name, "runtime.print") || callFn.name == "runtime._panic" || callFn.name == "runtime.hashmapGet" || callFn.name == "os.runtime_args":
				// These functions should be run at runtime. Specifically:
				//   * Print and panic functions are best emitted directly without
				//     interpreting them, otherwise we get a ton of putchar (etc.)
				//     calls.
				//   * runtime.hashmapGet tries to access the map value directly.
				//     This is not possible as the map value is treated as a special
				//     kind of object in this package.
				//   * os.runtime_args reads globals that are initialized outside
				//     the view of the interp package so it always needs to be run
				//     at runtime.
				err := r.runAtRuntime(fn, inst, locals, &mem, indent)
				if err != nil {
					return nil, mem, err
				}
			case callFn.name == "runtime.nanotime" && r.pkgName == "time":
				// The time package contains a call to runtime.nanotime.
				// This appears to be to work around a limitation in Windows
				// Server 2008:
				//   > Monotonic times are reported as offsets from startNano.
				//   > We initialize startNano to runtimeNano() - 1 so that on systems where
				//   > monotonic time resolution is fairly low (e.g. Windows 2008
				//   > which appears to have a default resolution of 15ms),
				//   > we avoid ever reporting a monotonic time of 0.
				//   > (Callers may want to use 0 as "time not set".)
				// Simply let runtime.nanotime return 0 in this case, which
				// should be fine and avoids a call to runtime.nanotime. It
				// means that monotonic time in the time package is counted from
				// time.Time{}.Sub(1), which should be fine.
				locals[inst.localIndex] = literalValue{uint64(0)}
			case callFn.name == "runtime.alloc":
				// Allocate heap memory. At compile time, this is instead done
				// by creating a global variable.

				// Get the requested memory size to be allocated.
				size := operands[1].Uint()

				// Get the object layout, if it is available.
				llvmLayoutType := r.getLLVMTypeFromLayout(operands[2])

				// Create the object.
				alloc := object{
					globalName:     r.pkgName + "$alloc",
					llvmLayoutType: llvmLayoutType,
					buffer:         newRawValue(uint32(size)),
					size:           uint32(size),
				}
				index := len(r.objects)
				r.objects = append(r.objects, alloc)

				// And create a pointer to this object, for working with it (so
				// that stores to it copy it, etc).
				ptr := newPointerValue(r, index, 0)
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"runtime.alloc:", size, "->", ptr)
				}
				locals[inst.localIndex] = ptr
			case callFn.name == "runtime.sliceCopy":
				// sliceCopy implements the built-in copy function for slices.
				// It is implemented here so that it can be used even if the
				// runtime implementation is not available. Doing it this way
				// may also be faster.
				// Code:
				// func sliceCopy(dst, src unsafe.Pointer, dstLen, srcLen uintptr, elemSize uintptr) int {
				//     n := srcLen
				//     if n > dstLen {
				//         n = dstLen
				//     }
				//     memmove(dst, src, n*elemSize)
				//     return int(n)
				// }
				dstLen := operands[3].Uint()
				srcLen := operands[4].Uint()
				elemSize := operands[5].Uint()
				n := srcLen
				if n > dstLen {
					n = dstLen
				}
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"copy:", operands[1], operands[2], n)
				}
				if n != 0 {
					// Only try to copy bytes when there are any bytes to copy.
					// This is not just an optimization. If one of the slices
					// (or both) are nil, the asPointer method call will fail
					// even though copying a nil slice is allowed.
					dst, err := operands[1].asPointer(r)
					if err != nil {
						return nil, mem, r.errorAt(inst, err)
					}
					src, err := operands[2].asPointer(r)
					if err != nil {
						return nil, mem, r.errorAt(inst, err)
					}
					nBytes := uint32(n * elemSize)
					dstObj := mem.getWritable(dst.index())
					dstBuf := dstObj.buffer.asRawValue(r)
					srcBuf := mem.get(src.index()).buffer.asRawValue(r)
					copy(dstBuf.buf[dst.offset():dst.offset()+nBytes], srcBuf.buf[src.offset():])
					dstObj.buffer = dstBuf
					mem.put(dst.index(), dstObj)
				}
				switch inst.llvmInst.Type().IntTypeWidth() {
				case 16:
					locals[inst.localIndex] = literalValue{uint16(n)}
				case 32:
					locals[inst.localIndex] = literalValue{uint32(n)}
				case 64:
					locals[inst.localIndex] = literalValue{uint64(n)}
				default:
					panic("unknown integer type width")
				}
			case strings.HasPrefix(callFn.name, "llvm.memcpy.p0i8.p0i8.") || strings.HasPrefix(callFn.name, "llvm.memmove.p0i8.p0i8."):
				// Copy a block of memory from one pointer to another.
				dst, err := operands[1].asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				src, err := operands[2].asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				nBytes := uint32(operands[3].Uint())
				dstObj := mem.getWritable(dst.index())
				dstBuf := dstObj.buffer.asRawValue(r)
				srcBuf := mem.get(src.index()).buffer.asRawValue(r)
				copy(dstBuf.buf[dst.offset():dst.offset()+nBytes], srcBuf.buf[src.offset():])
				dstObj.buffer = dstBuf
				mem.put(dst.index(), dstObj)
			case callFn.name == "(reflect.rawType).elem":
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"call (reflect.rawType).elem:", operands[1:])
				}
				// Extract the type code global from the first parameter.
				typecodeIDPtrToInt, err := operands[1].toLLVMValue(inst.llvmInst.Operand(0).Type(), &mem)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				typecodeID := typecodeIDPtrToInt.Operand(0)

				// Get the type class.
				// See also: getClassAndValueFromTypeCode in transform/reflect.go.
				typecodeName := typecodeID.Name()
				const prefix = "reflect/types.type:"
				if !strings.HasPrefix(typecodeName, prefix) {
					panic("unexpected typecode name: " + typecodeName)
				}
				id := typecodeName[len(prefix):]
				class := id[:strings.IndexByte(id, ':')]
				value := id[len(class)+1:]
				if class == "named" {
					// Get the underlying type.
					class = value[:strings.IndexByte(value, ':')]
					value = value[len(class)+1:]
				}

				// Elem() is only valid for certain type classes.
				switch class {
				case "chan", "pointer", "slice", "array":
					elementType := llvm.ConstExtractValue(typecodeID.Initializer(), []uint32{0})
					uintptrType := r.mod.Context().IntType(int(mem.r.pointerSize) * 8)
					locals[inst.localIndex] = r.getValue(llvm.ConstPtrToInt(elementType, uintptrType))
				default:
					return nil, mem, r.errorAt(inst, fmt.Errorf("(reflect.Type).Elem() called on %s type", class))
				}
			case callFn.name == "runtime.typeAssert":
				// This function must be implemented manually as it is normally
				// implemented by the interface lowering pass.
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"typeassert:", operands[1:])
				}
				assertedType, err := operands[2].toLLVMValue(inst.llvmInst.Operand(1).Type(), &mem)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				actualTypePtrToInt, err := operands[1].toLLVMValue(inst.llvmInst.Operand(0).Type(), &mem)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				actualType := actualTypePtrToInt.Operand(0)
				if strings.TrimPrefix(actualType.Name(), "reflect/types.type:") == strings.TrimPrefix(assertedType.Name(), "reflect/types.typeid:") {
					locals[inst.localIndex] = literalValue{uint8(1)}
				} else {
					locals[inst.localIndex] = literalValue{uint8(0)}
				}
			case strings.HasSuffix(callFn.name, ".$typeassert"):
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"interface assert:", operands[1:])
				}

				// Load various values for the interface implements check below.
				typecodePtr, err := operands[1].asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				methodSetPtr, err := mem.load(typecodePtr.addOffset(r.pointerSize*2), r.pointerSize).asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				methodSet := mem.get(methodSetPtr.index()).llvmGlobal.Initializer()
				llvmFn := inst.llvmInst.CalledValue()
				methodSetAttr := llvmFn.GetStringAttributeAtIndex(-1, "tinygo-methods")
				methodSetString := methodSetAttr.GetStringValue()

				// Make a set of all the methods on the concrete type, for
				// easier checking in the next step.
				concreteTypeMethods := map[string]struct{}{}
				for i := 0; i < methodSet.Type().ArrayLength(); i++ {
					methodInfo := llvm.ConstExtractValue(methodSet, []uint32{uint32(i)})
					name := llvm.ConstExtractValue(methodInfo, []uint32{0}).Name()
					concreteTypeMethods[name] = struct{}{}
				}

				// Check whether all interface methods are also in the list
				// of defined methods calculated above. This is the interface
				// assert itself.
				assertOk := uint8(1) // i1 true
				for _, name := range strings.Split(methodSetString, "; ") {
					if _, ok := concreteTypeMethods[name]; !ok {
						// There is a method on the interface that is not
						// implemented by the type. The assertion will fail.
						assertOk = 0 // i1 false
						break
					}
				}
				// If assertOk is still 1, the assertion succeeded.
				locals[inst.localIndex] = literalValue{assertOk}
			case strings.HasSuffix(callFn.name, "$invoke"):
				// This thunk is the interface method dispatcher: it is called
				// with all regular parameters and a type code. It will then
				// call the concrete method for it.
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"invoke method:", operands[1:])
				}

				// Load the type code of the interface value.
				typecodeIDBitCast, err := operands[len(operands)-3].toLLVMValue(inst.llvmInst.Operand(len(operands)-4).Type(), &mem)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				typecodeID := typecodeIDBitCast.Operand(0).Initializer()

				// Load the method set, which is part of the typecodeID object.
				methodSet := llvm.ConstExtractValue(typecodeID, []uint32{2}).Operand(0).Initializer()

				// We don't need to load the interface method set.

				// Load the signature of the to-be-called function.
				llvmFn := inst.llvmInst.CalledValue()
				invokeAttr := llvmFn.GetStringAttributeAtIndex(-1, "tinygo-invoke")
				invokeName := invokeAttr.GetStringValue()
				signature := r.mod.NamedGlobal(invokeName)

				// Iterate through all methods, looking for the one method that
				// should be returned.
				numMethods := methodSet.Type().ArrayLength()
				var method llvm.Value
				for i := 0; i < numMethods; i++ {
					methodSignature := llvm.ConstExtractValue(methodSet, []uint32{uint32(i), 0})
					if methodSignature == signature {
						method = llvm.ConstExtractValue(methodSet, []uint32{uint32(i), 1}).Operand(0)
					}
				}
				if method.IsNil() {
					return nil, mem, r.errorAt(inst, errors.New("could not find method: "+invokeName))
				}

				// Change the to-be-called function to the underlying method to
				// be called and fall through to the default case.
				callFn = r.getFunction(method)
				fallthrough
			default:
				if len(callFn.blocks) == 0 {
					// Call to a function declaration without a definition
					// available.
					err := r.runAtRuntime(fn, inst, locals, &mem, indent)
					if err != nil {
						return nil, mem, err
					}
					continue
				}
				// Call a function with a definition available. Run it as usual,
				// possibly trying to recover from it if it failed to execute.
				if r.debug {
					argStrings := make([]string, len(operands)-1)
					for i := range argStrings {
						argStrings[i] = operands[i+1].String()
					}
					fmt.Fprintln(os.Stderr, indent+"call:", callFn.name+"("+strings.Join(argStrings, ", ")+")")
				}
				retval, callMem, callErr := r.run(callFn, operands[1:], &mem, indent+"    ")
				if callErr != nil {
					if isRecoverableError(callErr.Err) {
						// This error can be recovered by doing the call at
						// runtime instead of at compile time. But we need to
						// revert any changes made by the call first.
						if r.debug {
							fmt.Fprintln(os.Stderr, indent+"!! revert because of error:", callErr.Err)
						}
						callMem.revert()
						err := r.runAtRuntime(fn, inst, locals, &mem, indent)
						if err != nil {
							return nil, mem, err
						}
						continue
					}
					// Add to the traceback, so that error handling code can see
					// how this function got called.
					callErr.Traceback = append(callErr.Traceback, ErrorLine{
						Pos:  getPosition(inst.llvmInst),
						Inst: inst.llvmInst,
					})
					return nil, mem, callErr
				}
				locals[inst.localIndex] = retval
				mem.extend(callMem)
			}
		case llvm.Load:
			// Load instruction, loading some data from the topmost memory view.
			ptr, err := operands[0].asPointer(r)
			if err != nil {
				return nil, mem, r.errorAt(inst, err)
			}
			size := operands[1].(literalValue).value.(uint64)
			if mem.hasExternalStore(ptr) {
				// If there could be an external store (for example, because a
				// pointer to the object was passed to a function that could not
				// be interpreted at compile time) then the load must be done at
				// runtime.
				err := r.runAtRuntime(fn, inst, locals, &mem, indent)
				if err != nil {
					return nil, mem, err
				}
				continue
			}
			result := mem.load(ptr, uint32(size))
			if result == nil {
				err := r.runAtRuntime(fn, inst, locals, &mem, indent)
				if err != nil {
					return nil, mem, err
				}
				continue
			}
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"load:", ptr, "->", result)
			}
			locals[inst.localIndex] = result
		case llvm.Store:
			// Store instruction. Create a new object in the memory view and
			// store to that, to make it possible to roll back this store.
			ptr, err := operands[1].asPointer(r)
			if err != nil {
				return nil, mem, r.errorAt(inst, err)
			}
			if mem.hasExternalLoadOrStore(ptr) {
				err := r.runAtRuntime(fn, inst, locals, &mem, indent)
				if err != nil {
					return nil, mem, err
				}
				continue
			}
			val := operands[0]
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"store:", val, ptr)
			}
			ok := mem.store(val, ptr)
			if !ok {
				// Could not store the value, do it at runtime.
				err := r.runAtRuntime(fn, inst, locals, &mem, indent)
				if err != nil {
					return nil, mem, err
				}
			}
		case llvm.Alloca:
			// Alloca normally allocates some stack memory. In the interpreter,
			// it allocates a global instead.
			// This can likely be optimized, as all it really needs is an alloca
			// in the initAll function and creating a global is wasteful for
			// this purpose.

			// Create the new object.
			size := operands[0].(literalValue).value.(uint64)
			alloca := object{
				llvmType:   inst.llvmInst.Type(),
				globalName: r.pkgName + "$alloca",
				buffer:     newRawValue(uint32(size)),
				size:       uint32(size),
			}
			index := len(r.objects)
			r.objects = append(r.objects, alloca)

			// Create a pointer to this object (an alloca produces a pointer).
			ptr := newPointerValue(r, index, 0)
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"alloca:", operands, "->", ptr)
			}
			locals[inst.localIndex] = ptr
		case llvm.GetElementPtr:
			// GetElementPtr does pointer arithmetic, changing the offset of the
			// pointer into the underlying object.
			var offset uint64
			for i := 2; i < len(operands); i += 2 {
				index := operands[i].Uint()
				elementSize := operands[i+1].Uint()
				if int64(elementSize) < 0 {
					// This is a struct field.
					offset += index
				} else {
					// This is a normal GEP, probably an array index.
					offset += elementSize * index
				}
			}
			ptr, err := operands[0].asPointer(r)
			if err != nil {
				if err != errIntegerAsPointer {
					return nil, mem, r.errorAt(inst, err)
				}
				// GEP on fixed pointer value (for example, memory-mapped I/O).
				ptrValue := operands[0].Uint() + offset
				switch operands[0].len(r) {
				case 8:
					locals[inst.localIndex] = literalValue{uint64(ptrValue)}
				case 4:
					locals[inst.localIndex] = literalValue{uint32(ptrValue)}
				case 2:
					locals[inst.localIndex] = literalValue{uint16(ptrValue)}
				default:
					panic("pointer operand is not of a known pointer size")
				}
				continue
			}
			ptr = ptr.addOffset(uint32(offset))
			locals[inst.localIndex] = ptr
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"gep:", operands, "->", ptr)
			}
		case llvm.BitCast, llvm.IntToPtr, llvm.PtrToInt:
			// Various bitcast-like instructions that all keep the same bits
			// while changing the LLVM type.
			// Because interp doesn't preserve the type, these operations are
			// identity operations.
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+instructionNameMap[inst.opcode]+":", operands[0])
			}
			locals[inst.localIndex] = operands[0]
		case llvm.ExtractValue:
			agg := operands[0].asRawValue(r)
			offset := operands[1].(literalValue).value.(uint64)
			size := operands[2].(literalValue).value.(uint64)
			elt := rawValue{
				buf: agg.buf[offset : offset+size],
			}
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"extractvalue:", operands, "->", elt)
			}
			locals[inst.localIndex] = elt
		case llvm.InsertValue:
			agg := operands[0].asRawValue(r)
			elt := operands[1].asRawValue(r)
			offset := int(operands[2].(literalValue).value.(uint64))
			newagg := newRawValue(uint32(len(agg.buf)))
			copy(newagg.buf, agg.buf)
			copy(newagg.buf[offset:], elt.buf)
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"insertvalue:", operands, "->", newagg)
			}
			locals[inst.localIndex] = newagg
		case llvm.ICmp:
			predicate := llvm.IntPredicate(operands[2].(literalValue).value.(uint8))
			var result bool
			lhs := operands[0]
			rhs := operands[1]
			switch predicate {
			case llvm.IntEQ, llvm.IntNE:
				lhsPointer, lhsErr := lhs.asPointer(r)
				rhsPointer, rhsErr := rhs.asPointer(r)
				if (lhsErr == nil) != (rhsErr == nil) {
					// Fast path: only one is a pointer, so they can't be equal.
					result = false
				} else if lhsErr == nil {
					// Both must be nil, so both are pointers.
					// Compare them directly.
					result = lhsPointer.equal(rhsPointer)
				} else {
					// Fall back to generic comparison.
					result = lhs.asRawValue(r).equal(rhs.asRawValue(r))
				}
				if predicate == llvm.IntNE {
					result = !result
				}
			case llvm.IntUGT:
				result = lhs.Uint() > rhs.Uint()
			case llvm.IntUGE:
				result = lhs.Uint() >= rhs.Uint()
			case llvm.IntULT:
				result = lhs.Uint() < rhs.Uint()
			case llvm.IntULE:
				result = lhs.Uint() <= rhs.Uint()
			case llvm.IntSGT:
				result = lhs.Int() > rhs.Int()
			case llvm.IntSGE:
				result = lhs.Int() >= rhs.Int()
			case llvm.IntSLT:
				result = lhs.Int() < rhs.Int()
			case llvm.IntSLE:
				result = lhs.Int() <= rhs.Int()
			default:
				return nil, mem, r.errorAt(inst, errors.New("interp: unsupported icmp"))
			}
			if result {
				locals[inst.localIndex] = literalValue{uint8(1)}
			} else {
				locals[inst.localIndex] = literalValue{uint8(0)}
			}
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"icmp:", operands[0], intPredicateString(predicate), operands[1], "->", result)
			}
		case llvm.FCmp:
			predicate := llvm.FloatPredicate(operands[2].(literalValue).value.(uint8))
			var result bool
			var lhs, rhs float64
			switch operands[0].len(r) {
			case 8:
				lhs = math.Float64frombits(operands[0].Uint())
				rhs = math.Float64frombits(operands[1].Uint())
			case 4:
				lhs = float64(math.Float32frombits(uint32(operands[0].Uint())))
				rhs = float64(math.Float32frombits(uint32(operands[1].Uint())))
			default:
				panic("unknown float type")
			}
			switch predicate {
			case llvm.FloatOEQ:
				result = lhs == rhs
			case llvm.FloatUNE:
				result = lhs != rhs
			case llvm.FloatOGT:
				result = lhs > rhs
			case llvm.FloatOGE:
				result = lhs >= rhs
			case llvm.FloatOLT:
				result = lhs < rhs
			case llvm.FloatOLE:
				result = lhs <= rhs
			default:
				return nil, mem, r.errorAt(inst, errors.New("interp: unsupported fcmp"))
			}
			if result {
				locals[inst.localIndex] = literalValue{uint8(1)}
			} else {
				locals[inst.localIndex] = literalValue{uint8(0)}
			}
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+"fcmp:", operands[0], predicate, operands[1], "->", result)
			}
		case llvm.Add, llvm.Sub, llvm.Mul, llvm.UDiv, llvm.SDiv, llvm.URem, llvm.SRem, llvm.Shl, llvm.LShr, llvm.AShr, llvm.And, llvm.Or, llvm.Xor:
			// Integer binary operations.
			lhs := operands[0]
			rhs := operands[1]
			lhsPtr, err := lhs.asPointer(r)
			if err == nil {
				// The lhs is a pointer. This sometimes happens for particular
				// pointer tricks.
				switch inst.opcode {
				case llvm.Add:
					// This likely means this is part of a
					// unsafe.Pointer(uintptr(ptr) + offset) pattern.
					lhsPtr = lhsPtr.addOffset(uint32(rhs.Uint()))
					locals[inst.localIndex] = lhsPtr
					continue
				case llvm.Xor:
					if rhs.Uint() == 0 {
						// Special workaround for strings.noescape, see
						// src/strings/builder.go in the Go source tree. This is
						// the identity operator, so we can return the input.
						locals[inst.localIndex] = lhs
						continue
					}
				default:
					// Catch-all for weird operations that should just be done
					// at runtime.
					err := r.runAtRuntime(fn, inst, locals, &mem, indent)
					if err != nil {
						return nil, mem, err
					}
					continue
				}
			}
			var result uint64
			switch inst.opcode {
			case llvm.Add:
				result = lhs.Uint() + rhs.Uint()
			case llvm.Sub:
				result = lhs.Uint() - rhs.Uint()
			case llvm.Mul:
				result = lhs.Uint() * rhs.Uint()
			case llvm.UDiv:
				result = lhs.Uint() / rhs.Uint()
			case llvm.SDiv:
				result = uint64(lhs.Int() / rhs.Int())
			case llvm.URem:
				result = lhs.Uint() % rhs.Uint()
			case llvm.SRem:
				result = uint64(lhs.Int() % rhs.Int())
			case llvm.Shl:
				result = lhs.Uint() << rhs.Uint()
			case llvm.LShr:
				result = lhs.Uint() >> rhs.Uint()
			case llvm.AShr:
				result = uint64(lhs.Int() >> rhs.Uint())
			case llvm.And:
				result = lhs.Uint() & rhs.Uint()
			case llvm.Or:
				result = lhs.Uint() | rhs.Uint()
			case llvm.Xor:
				result = lhs.Uint() ^ rhs.Uint()
			default:
				panic("unreachable")
			}
			switch lhs.len(r) {
			case 8:
				locals[inst.localIndex] = literalValue{result}
			case 4:
				locals[inst.localIndex] = literalValue{uint32(result)}
			case 2:
				locals[inst.localIndex] = literalValue{uint16(result)}
			case 1:
				locals[inst.localIndex] = literalValue{uint8(result)}
			default:
				panic("unknown integer size")
			}
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+instructionNameMap[inst.opcode]+":", lhs, rhs, "->", result)
			}
		case llvm.SExt, llvm.ZExt, llvm.Trunc:
			// Change the size of an integer to a larger or smaller bit width.
			// We make use of the fact that the Uint() function already
			// zero-extends the value and that Int() already sign-extends the
			// value, so we only need to truncate it to the appropriate bit
			// width. This means we can implement sext, zext and trunc in the
			// same way, by first {zero,sign}extending all the way up to uint64
			// and then truncating it as necessary.
			var value uint64
			if inst.opcode == llvm.SExt {
				value = uint64(operands[0].Int())
			} else {
				value = operands[0].Uint()
			}
			bitwidth := operands[1].Uint()
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+instructionNameMap[inst.opcode]+":", value, bitwidth)
			}
			switch bitwidth {
			case 64:
				locals[inst.localIndex] = literalValue{value}
			case 32:
				locals[inst.localIndex] = literalValue{uint32(value)}
			case 16:
				locals[inst.localIndex] = literalValue{uint16(value)}
			case 8:
				locals[inst.localIndex] = literalValue{uint8(value)}
			default:
				panic("unknown integer size in sext/zext/trunc")
			}
		case llvm.SIToFP, llvm.UIToFP:
			var value float64
			switch inst.opcode {
			case llvm.SIToFP:
				value = float64(operands[0].Int())
			case llvm.UIToFP:
				value = float64(operands[0].Uint())
			}
			bitwidth := operands[1].Uint()
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+instructionNameMap[inst.opcode]+":", value, bitwidth)
			}
			switch bitwidth {
			case 64:
				locals[inst.localIndex] = literalValue{math.Float64bits(value)}
			case 32:
				locals[inst.localIndex] = literalValue{math.Float32bits(float32(value))}
			default:
				panic("unknown integer size in sitofp/uitofp")
			}
		default:
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+inst.String())
			}
			return nil, mem, r.errorAt(inst, errUnsupportedInst)
		}
	}
	return nil, mem, r.errorAt(bb.instructions[len(bb.instructions)-1], errors.New("interp: reached end of basic block without terminator"))
}

func (r *runner) runAtRuntime(fn *function, inst instruction, locals []value, mem *memoryView, indent string) *Error {
	numOperands := inst.llvmInst.OperandsCount()
	operands := make([]llvm.Value, numOperands)
	for i := 0; i < numOperands; i++ {
		operand := inst.llvmInst.Operand(i)
		if !operand.IsAInstruction().IsNil() || !operand.IsAArgument().IsNil() {
			var err error
			operand, err = locals[fn.locals[operand]].toLLVMValue(operand.Type(), mem)
			if err != nil {
				return r.errorAt(inst, err)
			}
		}
		operands[i] = operand
	}
	if r.debug {
		fmt.Fprintln(os.Stderr, indent+inst.String())
	}
	var result llvm.Value
	switch inst.opcode {
	case llvm.Call:
		llvmFn := operands[len(operands)-1]
		args := operands[:len(operands)-1]
		for _, arg := range args {
			if arg.Type().TypeKind() == llvm.PointerTypeKind {
				mem.markExternalStore(arg)
			}
		}
		result = r.builder.CreateCall(llvmFn, args, inst.name)
	case llvm.Load:
		mem.markExternalLoad(operands[0])
		result = r.builder.CreateLoad(operands[0], inst.name)
		if inst.llvmInst.IsVolatile() {
			result.SetVolatile(true)
		}
	case llvm.Store:
		mem.markExternalStore(operands[1])
		result = r.builder.CreateStore(operands[0], operands[1])
		if inst.llvmInst.IsVolatile() {
			result.SetVolatile(true)
		}
	case llvm.BitCast:
		result = r.builder.CreateBitCast(operands[0], inst.llvmInst.Type(), inst.name)
	case llvm.ExtractValue:
		indices := inst.llvmInst.Indices()
		// Note: the Go LLVM API doesn't support multiple indices, so simulate
		// this operation with some extra extractvalue instructions. Hopefully
		// this is optimized to a single instruction.
		agg := operands[0]
		for i := 0; i < len(indices)-1; i++ {
			agg = r.builder.CreateExtractValue(agg, int(indices[i]), inst.name+".agg")
		}
		result = r.builder.CreateExtractValue(agg, int(indices[len(indices)-1]), inst.name)
	case llvm.InsertValue:
		indices := inst.llvmInst.Indices()
		// Similar to extractvalue, we're working around a limitation in the Go
		// LLVM API here by splitting the insertvalue into multiple instructions
		// if there is more than one operand.
		agg := operands[0]
		aggregates := []llvm.Value{agg}
		for i := 0; i < len(indices)-1; i++ {
			agg = r.builder.CreateExtractValue(agg, int(indices[i]), inst.name+".agg"+strconv.Itoa(i))
			aggregates = append(aggregates, agg)
		}
		result = operands[1]
		for i := len(indices) - 1; i >= 0; i-- {
			agg := aggregates[i]
			result = r.builder.CreateInsertValue(agg, result, int(indices[i]), inst.name+".insertvalue"+strconv.Itoa(i))
		}

	case llvm.Add:
		result = r.builder.CreateAdd(operands[0], operands[1], inst.name)
	case llvm.Sub:
		result = r.builder.CreateSub(operands[0], operands[1], inst.name)
	case llvm.Mul:
		result = r.builder.CreateMul(operands[0], operands[1], inst.name)
	case llvm.UDiv:
		result = r.builder.CreateUDiv(operands[0], operands[1], inst.name)
	case llvm.SDiv:
		result = r.builder.CreateSDiv(operands[0], operands[1], inst.name)
	case llvm.URem:
		result = r.builder.CreateURem(operands[0], operands[1], inst.name)
	case llvm.SRem:
		result = r.builder.CreateSRem(operands[0], operands[1], inst.name)
	case llvm.ZExt:
		result = r.builder.CreateZExt(operands[0], inst.llvmInst.Type(), inst.name)
	default:
		return r.errorAt(inst, errUnsupportedRuntimeInst)
	}
	locals[inst.localIndex] = localValue{result}
	mem.instructions = append(mem.instructions, result)
	return nil
}

func intPredicateString(predicate llvm.IntPredicate) string {
	switch predicate {
	case llvm.IntEQ:
		return "eq"
	case llvm.IntNE:
		return "ne"
	case llvm.IntUGT:
		return "ugt"
	case llvm.IntUGE:
		return "uge"
	case llvm.IntULT:
		return "ult"
	case llvm.IntULE:
		return "ule"
	case llvm.IntSGT:
		return "sgt"
	case llvm.IntSGE:
		return "sge"
	case llvm.IntSLT:
		return "slt"
	case llvm.IntSLE:
		return "sle"
	default:
		return "cmp?"
	}
}
