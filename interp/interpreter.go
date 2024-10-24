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

	// Parameters are considered a kind of local values.
	for i, param := range params {
		locals[i] = param
	}

	// Track what blocks have run instructions at runtime.
	// This is used to prevent unrolling.
	var runtimeBlocks map[int]struct{}

	// Start with the first basic block and the first instruction.
	// Branch instructions may modify both bb and instIndex when branching.
	bb := fn.blocks[0]
	currentBB := 0
	lastBB := -1 // last basic block is undefined, only defined after a branch
	var operands []value
	startRTInsts := len(mem.instructions)
	for instIndex := 0; instIndex < len(bb.instructions); instIndex++ {
		if instIndex == 0 {
			// This is the start of a new basic block.
			if len(mem.instructions) != startRTInsts {
				if _, ok := runtimeBlocks[lastBB]; ok {
					// This loop has been unrolled.
					// Avoid doing this, as it can result in a large amount of extra machine code.
					// This currently uses the branch from the last block, as there is no available information to give a better location.
					lastBBInsts := fn.blocks[lastBB].instructions
					return nil, mem, r.errorAt(lastBBInsts[len(lastBBInsts)-1], errLoopUnrolled)
				}

				// Flag the last block as having run stuff at runtime.
				if runtimeBlocks == nil {
					runtimeBlocks = make(map[int]struct{})
				}
				runtimeBlocks[lastBB] = struct{}{}

				// Reset the block-start runtime instructions counter.
				startRTInsts = len(mem.instructions)
			}

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
					index, ok := fn.locals[v.value]
					if !ok {
						// This is a localValue that is not local to the
						// function. An example would be an inline assembly call
						// operand.
						isRuntimeInst = true
						break
					}
					localVal := locals[index]
					if localVal == nil {
						// Trying to read a function-local value before it is
						// set.
						return nil, mem, r.errorAt(inst, errors.New("interp: local not defined"))
					} else {
						operands = append(operands, localVal)
						if _, ok := localVal.(localValue); ok {
							// The function-local value is still just a
							// localValue (which can't be interpreted at compile
							// time). Not sure whether this ever happens in
							// practice.
							isRuntimeInst = true
							break
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
			if time.Since(r.start) > r.timeout {
				// Running for more than the allowed timeout; This shouldn't happen, but it does.
				// See github.com/tinygo-org/tinygo/issues/2124
				return nil, mem, r.errorAt(fn.blocks[0].instructions[0], fmt.Errorf("interp: running for more than %s, timing out (executed calls: %d)", r.timeout, r.callsExecuted))
			}

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
				switch operands[0].Uint(r) {
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
			value := operands[0].Uint(r)
			targetLabel := operands[1].Uint(r) // default label
			// Do a lazy switch by iterating over all cases.
			for i := 2; i < len(operands); i += 2 {
				if value == operands[i].Uint(r) {
					targetLabel = operands[i+1].Uint(r)
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
			switch operands[0].Uint(r) {
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
			case strings.HasPrefix(callFn.name, "runtime.print") || callFn.name == "runtime._panic" || callFn.name == "runtime.hashmapGet" || callFn.name == "runtime.hashmapInterfaceHash" ||
				callFn.name == "os.runtime_args" || callFn.name == "syscall.runtime_envs" ||
				callFn.name == "internal/task.start" || callFn.name == "internal/task.Current" ||
				callFn.name == "time.startTimer" || callFn.name == "time.stopTimer" || callFn.name == "time.resetTimer":
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
				//   * internal/task.start, internal/task.Current: start and read shcheduler state,
				//     which is modified elsewhere.
				//   * Timer functions access runtime internal state which may
				//     not be initialized.
				err := r.runAtRuntime(fn, inst, locals, &mem, indent)
				if err != nil {
					return nil, mem, err
				}
			case callFn.name == "internal/task.Pause":
				// Task scheduling isn't possible at compile time.
				return nil, mem, r.errorAt(inst, errUnsupportedRuntimeInst)
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
				size := operands[1].Uint(r)

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
				dstLen := operands[3].Uint(r)
				srcLen := operands[4].Uint(r)
				elemSize := operands[5].Uint(r)
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
					if mem.hasExternalStore(src) || mem.hasExternalLoadOrStore(dst) {
						// These are the same checks as there are on llvm.Load
						// and llvm.Store in the interpreter. Copying is
						// essentially loading from the source array and storing
						// to the destination array, hence why we need to do the
						// same checks here.
						// This fixes the following bug:
						// https://github.com/tinygo-org/tinygo/issues/3890
						err := r.runAtRuntime(fn, inst, locals, &mem, indent)
						if err != nil {
							return nil, mem, err
						}
						continue
					}
					nBytes := uint32(n * elemSize)
					dstObj := mem.getWritable(dst.index())
					dstBuf := dstObj.buffer.asRawValue(r)
					srcBuf := mem.get(src.index()).buffer.asRawValue(r)
					copy(dstBuf.buf[dst.offset():dst.offset()+nBytes], srcBuf.buf[src.offset():])
					dstObj.buffer = dstBuf
					mem.put(dst.index(), dstObj)
				}
				locals[inst.localIndex] = makeLiteralInt(n, inst.llvmInst.Type().IntTypeWidth())
			case strings.HasPrefix(callFn.name, "llvm.memcpy.p0") || strings.HasPrefix(callFn.name, "llvm.memmove.p0"):
				// Copy a block of memory from one pointer to another.
				dst, err := operands[1].asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				src, err := operands[2].asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				nBytes := uint32(operands[3].Uint(r))
				dstObj := mem.getWritable(dst.index())
				dstBuf := dstObj.buffer.asRawValue(r)
				if mem.get(src.index()).buffer == nil {
					// Looks like the source buffer is not defined.
					// This can happen with //extern or //go:embed.
					return nil, mem, r.errorAt(inst, errUnsupportedRuntimeInst)
				}
				srcBuf := mem.get(src.index()).buffer.asRawValue(r)
				copy(dstBuf.buf[dst.offset():dst.offset()+nBytes], srcBuf.buf[src.offset():])
				dstObj.buffer = dstBuf
				mem.put(dst.index(), dstObj)
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
				actualType, err := operands[1].toLLVMValue(inst.llvmInst.Operand(0).Type(), &mem)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				if !actualType.IsAConstantInt().IsNil() && actualType.ZExtValue() == 0 {
					locals[inst.localIndex] = literalValue{uint8(0)}
					break
				}
				// Strip pointer casts (bitcast, getelementptr).
				for !actualType.IsAConstantExpr().IsNil() {
					opcode := actualType.Opcode()
					if opcode != llvm.GetElementPtr && opcode != llvm.BitCast {
						break
					}
					actualType = actualType.Operand(0)
				}
				if strings.TrimPrefix(actualType.Name(), "reflect/types.type:") == strings.TrimPrefix(assertedType.Name(), "reflect/types.typeid:") {
					locals[inst.localIndex] = literalValue{uint8(1)}
				} else {
					locals[inst.localIndex] = literalValue{uint8(0)}
				}
			case callFn.name == "__tinygo_interp_raise_test_error":
				// Special function that will trigger an error.
				// This is used to test error reporting.
				return nil, mem, r.errorAt(inst, errors.New("test error"))
			case strings.HasSuffix(callFn.name, ".$typeassert"):
				if r.debug {
					fmt.Fprintln(os.Stderr, indent+"interface assert:", operands[1:])
				}

				// Load various values for the interface implements check below.
				typecodePtr, err := operands[1].asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				// typecodePtr always point to the numMethod field in the type
				// description struct. The methodSet, when present, comes right
				// before the numMethod field (the compiler doesn't generate
				// method sets for concrete types without methods).
				// Considering that the compiler doesn't emit interface type
				// asserts for interfaces with no methods (as the always succeed)
				// then if the offset is zero, this assert must always fail.
				if typecodePtr.offset() == 0 {
					locals[inst.localIndex] = literalValue{uint8(0)}
					break
				}
				typecodePtrOffset, err := typecodePtr.addOffset(-int64(r.pointerSize))
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				methodSetPtr, err := mem.load(typecodePtrOffset, r.pointerSize).asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				methodSet := mem.get(methodSetPtr.index()).llvmGlobal.Initializer()
				numMethods := int(r.builder.CreateExtractValue(methodSet, 0, "").ZExtValue())
				llvmFn := inst.llvmInst.CalledValue()
				methodSetAttr := llvmFn.GetStringAttributeAtIndex(-1, "tinygo-methods")
				methodSetString := methodSetAttr.GetStringValue()

				// Make a set of all the methods on the concrete type, for
				// easier checking in the next step.
				concreteTypeMethods := map[string]struct{}{}
				for i := 0; i < numMethods; i++ {
					methodInfo := r.builder.CreateExtractValue(methodSet, 1, "")
					name := r.builder.CreateExtractValue(methodInfo, i, "").Name()
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

				// Load the type code and method set of the interface value.
				typecodePtr, err := operands[len(operands)-2].asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				typecodePtrOffset, err := typecodePtr.addOffset(-int64(r.pointerSize))
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				methodSetPtr, err := mem.load(typecodePtrOffset, r.pointerSize).asPointer(r)
				if err != nil {
					return nil, mem, r.errorAt(inst, err)
				}
				methodSet := mem.get(methodSetPtr.index()).llvmGlobal.Initializer()

				// We don't need to load the interface method set.

				// Load the signature of the to-be-called function.
				llvmFn := inst.llvmInst.CalledValue()
				invokeAttr := llvmFn.GetStringAttributeAtIndex(-1, "tinygo-invoke")
				invokeName := invokeAttr.GetStringValue()
				signature := r.mod.NamedGlobal(invokeName)

				// Iterate through all methods, looking for the one method that
				// should be returned.
				numMethods := int(r.builder.CreateExtractValue(methodSet, 0, "").ZExtValue())
				var method llvm.Value
				for i := 0; i < numMethods; i++ {
					methodSignatureAgg := r.builder.CreateExtractValue(methodSet, 1, "")
					methodSignature := r.builder.CreateExtractValue(methodSignatureAgg, i, "")
					if methodSignature == signature {
						methodAgg := r.builder.CreateExtractValue(methodSet, 2, "")
						method = r.builder.CreateExtractValue(methodAgg, i, "")
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
					for i, v := range operands[1:] {
						argStrings[i] = v.String()
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
						Inst: inst.llvmInst.String(),
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
			if inst.llvmInst.IsVolatile() || inst.llvmInst.Ordering() != llvm.AtomicOrderingNotAtomic || mem.hasExternalStore(ptr) {
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
			if inst.llvmInst.IsVolatile() || inst.llvmInst.Ordering() != llvm.AtomicOrderingNotAtomic || mem.hasExternalLoadOrStore(ptr) {
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
				llvmType:   inst.llvmInst.AllocatedType(),
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
			var offset int64
			for i := 1; i < len(operands); i += 2 {
				index := operands[i].Int(r)
				elementSize := operands[i+1].Int(r)
				if elementSize < 0 {
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
				ptrValue := operands[0].Uint(r) + uint64(offset)
				locals[inst.localIndex] = makeLiteralInt(ptrValue, int(operands[0].len(r)*8))
				continue
			}
			ptr, err = ptr.addOffset(int64(offset))
			if err != nil {
				return nil, mem, r.errorAt(inst, err)
			}
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
			lhs := operands[0]
			rhs := operands[1]
			result := r.interpretICmp(lhs, rhs, predicate)
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
				lhs = math.Float64frombits(operands[0].Uint(r))
				rhs = math.Float64frombits(operands[1].Uint(r))
			case 4:
				lhs = float64(math.Float32frombits(uint32(operands[0].Uint(r))))
				rhs = float64(math.Float32frombits(uint32(operands[1].Uint(r))))
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
				if inst.opcode == llvm.Add {
					// This likely means this is part of a
					// unsafe.Pointer(uintptr(ptr) + offset) pattern.
					lhsPtr, err = lhsPtr.addOffset(int64(rhs.Uint(r)))
					if err != nil {
						return nil, mem, r.errorAt(inst, err)
					}
					locals[inst.localIndex] = lhsPtr
				} else if inst.opcode == llvm.Xor && rhs.Uint(r) == 0 {
					// Special workaround for strings.noescape, see
					// src/strings/builder.go in the Go source tree. This is
					// the identity operator, so we can return the input.
					locals[inst.localIndex] = lhs
				} else if inst.opcode == llvm.And && rhs.Uint(r) < 8 {
					// This is probably part of a pattern to get the lower bits
					// of a pointer for pointer tagging, like this:
					//     uintptr(unsafe.Pointer(t)) & 0b11
					// We can actually support this easily by ANDing with the
					// pointer offset.
					result := uint64(lhsPtr.offset()) & rhs.Uint(r)
					locals[inst.localIndex] = makeLiteralInt(result, int(lhs.len(r)*8))
				} else {
					// Catch-all for weird operations that should just be done
					// at runtime.
					err := r.runAtRuntime(fn, inst, locals, &mem, indent)
					if err != nil {
						return nil, mem, err
					}
				}
				continue
			}
			var result uint64
			switch inst.opcode {
			case llvm.Add:
				result = lhs.Uint(r) + rhs.Uint(r)
			case llvm.Sub:
				result = lhs.Uint(r) - rhs.Uint(r)
			case llvm.Mul:
				result = lhs.Uint(r) * rhs.Uint(r)
			case llvm.UDiv:
				result = lhs.Uint(r) / rhs.Uint(r)
			case llvm.SDiv:
				result = uint64(lhs.Int(r) / rhs.Int(r))
			case llvm.URem:
				result = lhs.Uint(r) % rhs.Uint(r)
			case llvm.SRem:
				result = uint64(lhs.Int(r) % rhs.Int(r))
			case llvm.Shl:
				result = lhs.Uint(r) << rhs.Uint(r)
			case llvm.LShr:
				result = lhs.Uint(r) >> rhs.Uint(r)
			case llvm.AShr:
				result = uint64(lhs.Int(r) >> rhs.Uint(r))
			case llvm.And:
				result = lhs.Uint(r) & rhs.Uint(r)
			case llvm.Or:
				result = lhs.Uint(r) | rhs.Uint(r)
			case llvm.Xor:
				result = lhs.Uint(r) ^ rhs.Uint(r)
			default:
				panic("unreachable")
			}
			locals[inst.localIndex] = makeLiteralInt(result, int(lhs.len(r)*8))
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
				value = uint64(operands[0].Int(r))
			} else {
				value = operands[0].Uint(r)
			}
			bitwidth := operands[1].Uint(r)
			if r.debug {
				fmt.Fprintln(os.Stderr, indent+instructionNameMap[inst.opcode]+":", value, bitwidth)
			}
			locals[inst.localIndex] = makeLiteralInt(value, int(bitwidth))
		case llvm.SIToFP, llvm.UIToFP:
			var value float64
			switch inst.opcode {
			case llvm.SIToFP:
				value = float64(operands[0].Int(r))
			case llvm.UIToFP:
				value = float64(operands[0].Uint(r))
			}
			bitwidth := operands[1].Uint(r)
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

// Interpret an icmp instruction. Doesn't have side effects, only returns the
// output of the comparison.
func (r *runner) interpretICmp(lhs, rhs value, predicate llvm.IntPredicate) bool {
	switch predicate {
	case llvm.IntEQ, llvm.IntNE:
		var result bool
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
		return result
	case llvm.IntUGT:
		return lhs.Uint(r) > rhs.Uint(r)
	case llvm.IntUGE:
		return lhs.Uint(r) >= rhs.Uint(r)
	case llvm.IntULT:
		return lhs.Uint(r) < rhs.Uint(r)
	case llvm.IntULE:
		return lhs.Uint(r) <= rhs.Uint(r)
	case llvm.IntSGT:
		return lhs.Int(r) > rhs.Int(r)
	case llvm.IntSGE:
		return lhs.Int(r) >= rhs.Int(r)
	case llvm.IntSLT:
		return lhs.Int(r) < rhs.Int(r)
	case llvm.IntSLE:
		return lhs.Int(r) <= rhs.Int(r)
	default:
		// _should_ be unreachable, until LLVM adds new icmp operands (unlikely)
		panic("interp: unsupported icmp")
	}
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
				err := mem.markExternalStore(arg)
				if err != nil {
					return r.errorAt(inst, err)
				}
			}
		}
		result = r.builder.CreateCall(inst.llvmInst.CalledFunctionType(), llvmFn, args, inst.name)
	case llvm.Load:
		err := mem.markExternalLoad(operands[0])
		if err != nil {
			return r.errorAt(inst, err)
		}
		result = r.builder.CreateLoad(inst.llvmInst.Type(), operands[0], inst.name)
		if inst.llvmInst.IsVolatile() {
			result.SetVolatile(true)
		}
		if ordering := inst.llvmInst.Ordering(); ordering != llvm.AtomicOrderingNotAtomic {
			result.SetOrdering(ordering)
		}
	case llvm.Store:
		err := mem.markExternalStore(operands[1])
		if err != nil {
			return r.errorAt(inst, err)
		}
		result = r.builder.CreateStore(operands[0], operands[1])
		if inst.llvmInst.IsVolatile() {
			result.SetVolatile(true)
		}
		if ordering := inst.llvmInst.Ordering(); ordering != llvm.AtomicOrderingNotAtomic {
			result.SetOrdering(ordering)
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
			mem.instructions = append(mem.instructions, agg)
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
			mem.instructions = append(mem.instructions, agg)
		}
		result = operands[1]
		for i := len(indices) - 1; i >= 0; i-- {
			agg := aggregates[i]
			result = r.builder.CreateInsertValue(agg, result, int(indices[i]), inst.name+".insertvalue"+strconv.Itoa(i))
			if i != 0 { // don't add last result to mem.instructions as it will be done at the end already
				mem.instructions = append(mem.instructions, result)
			}
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
