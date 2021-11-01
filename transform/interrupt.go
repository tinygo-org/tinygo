package transform

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

// LowerInterrupts creates interrupt handlers for the interrupts created by
// runtime/interrupt.New.
//
// The operation is as follows. The compiler creates the following during IR
// generation:
//     * calls to runtime/interrupt.Register that map interrupt IDs to ISR names.
//     * runtime/interrupt.handle objects that store the (constant) interrupt ID and
//       interrupt handler func value.
//
// This pass then creates the specially named interrupt handler names that
// simply call the registered handlers. This might seem like it causes extra
// overhead, but in fact inlining and const propagation will eliminate most if
// not all of that.
func LowerInterrupts(mod llvm.Module, config *compileopts.Config) []error {
	var errs []error

	// Discover interrupts. The runtime/interrupt.Register call is a compiler
	// intrinsic that maps interrupt numbers to handler names.
	handlerNames := map[int64]string{}
	for _, call := range getUses(mod.NamedFunction("runtime/interrupt.Register")) {
		if call.IsACallInst().IsNil() {
			errs = append(errs, errorAt(call, "expected a call to runtime/interrupt.Register?"))
			continue
		}

		num := call.Operand(0)
		if num.IsAConstant().IsNil() {
			errs = append(errs, errorAt(call, "non-constant interrupt number?"))
			continue
		}

		// extract the interrupt name
		nameStrGEP := call.Operand(1)
		if nameStrGEP.IsAConstantExpr().IsNil() || nameStrGEP.Opcode() != llvm.GetElementPtr {
			errs = append(errs, errorAt(call, "expected a string operand?"))
			continue
		}
		nameStrPtr := nameStrGEP.Operand(0) // note: assuming it's a GEP to the first byte
		nameStrLen := call.Operand(2)
		if nameStrPtr.IsAGlobalValue().IsNil() || !nameStrPtr.IsGlobalConstant() || nameStrLen.IsAConstant().IsNil() {
			errs = append(errs, errorAt(call, "non-constant interrupt name?"))
			continue
		}

		// keep track of this name
		name := string(getGlobalBytes(nameStrPtr)[:nameStrLen.SExtValue()])
		handlerNames[num.SExtValue()] = name

		// remove this pseudo-call
		call.ReplaceAllUsesWith(llvm.ConstNull(call.Type()))
		call.EraseFromParentAsInstruction()
	}

	hasSoftwareVectoring := hasUses(mod.NamedFunction("runtime.callInterruptHandler"))
	softwareVector := make(map[int64]llvm.Value)

	ctx := mod.Context()
	i8ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	nullptr := llvm.ConstNull(i8ptrType)
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	// Create a function type with the signature of an interrupt handler.
	fnType := llvm.FunctionType(ctx.VoidType(), nil, false)

	// Collect a map of interrupt handle objects. The fact that they still
	// exist in the IR indicates that they could not be optimized away,
	// therefore we need to make real interrupt handlers for them.
	handleMap := map[int][]llvm.Value{}
	handleType := mod.GetTypeByName("runtime/interrupt.handle")
	if !handleType.IsNil() {
		handlePtrType := llvm.PointerType(handleType, 0)
		for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
			if global.Type() != handlePtrType {
				continue
			}

			// Get the interrupt number from the initializer
			initializer := global.Initializer()
			num := int(llvm.ConstExtractValue(initializer, []uint32{1, 0}).SExtValue())
			pkg := packageFromInterruptHandle(global)

			handles, exists := handleMap[num]

			// If there is an existing interrupt handler, ensure it is in the same package
			// as the new one.  This is to prevent any assumptions in code that a single
			// compiler pass can see all packages to chain interrupt handlers. When packages are
			// compiled to separate object files, the linker should spot the duplicate symbols
			// for the wrapper function, failing the build.
			if exists && packageFromInterruptHandle(handles[0]) != pkg {
				errs = append(errs, errorAt(global,
					fmt.Sprintf("handlers for interrupt %d (%s) in multiple packages: %s and %s",
						num, handlerNames[int64(num)], pkg, packageFromInterruptHandle(handles[0]))))
				continue
			}

			handleMap[num] = append(handles, global)
		}
	}

	// Output interrupts in numerical order for reproducible builds (Go map
	// intentionally randomizes iteration order of maps).
	interrupts := make([]int, 0, len(handleMap))
	for k := range handleMap {
		interrupts = append(interrupts, k)
	}
	sort.Ints(interrupts)

	// Iterate over all handle objects, replacing their ptrtoint uses with a
	// real interrupt ID and creating an interrupt handler for them.
	for _, interrupt := range interrupts {
		handles := handleMap[interrupt]

		// There is always at least one handler for each interrupt number.  We
		// arbitrarily take the first handler to attach any errors to.
		first := handles[0]

		initializer := first.Initializer()
		num := llvm.ConstExtractValue(initializer, []uint32{1, 0})
		name := handlerNames[num.SExtValue()]

		isSoftwareVectored := false
		if name == "" {
			// No function name was defined for this interrupt number, which
			// probably means one of two things:
			//   * runtime/interrupt.Register wasn't called to give the interrupt
			//     number a function name (such as on Cortex-M).
			//   * We're using software vectoring instead of hardware vectoring,
			//     which means the name of the handler doesn't matter (it will
			//     probably be inlined anyway).
			if hasSoftwareVectoring {
				isSoftwareVectored = true
				if name == "" {
					// Name doesn't matter, so pick something unique.
					name = "runtime/interrupt.interruptHandler" + strconv.FormatInt(num.SExtValue(), 10)
				}
			} else {
				errs = append(errs, errorAt(first, fmt.Sprintf("cannot find interrupt name for number %d", num.SExtValue())))
				continue
			}
		}

		// Check for an existing interrupt handler, and report it as an error if
		// there is one.
		fn := mod.NamedFunction(name)
		if fn.IsNil() {
			fn = llvm.AddFunction(mod, name, fnType)
		} else if fn.Type().ElementType() != fnType {
			// Don't bother with a precise error message (listing the previsous
			// location) because this should not normally happen anyway.
			errs = append(errs, errorAt(first, name+" redeclared with a different signature"))
			continue
		} else if !fn.IsDeclaration() {
			errValue := name + " redeclared in this program"
			fnPos := getPosition(fn)
			if fnPos.IsValid() {
				errValue += "\n\tprevious declaration at " + fnPos.String()
			}
			errs = append(errs, errorAt(first, errValue))
			continue
		}

		// Create the wrapper function which is the actual interrupt handler
		// that is inserted in the interrupt vector.
		fn.SetUnnamedAddr(true)
		AddStandardAttributes(fn, config)
		fn.SetSection(".text." + name)
		if isSoftwareVectored {
			fn.SetLinkage(llvm.InternalLinkage)
			softwareVector[num.SExtValue()] = fn
		}
		entryBlock := ctx.AddBasicBlock(fn, "entry")
		builder.SetInsertPointAtEnd(entryBlock)

		// Set the 'interrupt' flag if needed on this platform.
		if strings.HasPrefix(mod.Target(), "avr") {
			// This special calling convention is needed on AVR to save and
			// restore all clobbered registers, instead of just the ones that
			// would need to be saved/restored in a normal function call.
			// Note that the AVR_INTERRUPT calling convention would enable
			// interrupts right at the beginning of the handler, potentially
			// leading to lots of nested interrupts and a stack overflow.
			fn.SetFunctionCallConv(85) // CallingConv::AVR_SIGNAL
		}

		// For each handle (i.e. each call to interrupt.New), check the usage,
		// output a call to the actual handler function and clean-up the handle
		// that is no longer needed.
		for _, handler := range handles {
			// Extract the func value.
			initializer := handler.Initializer()
			handlerContext := llvm.ConstExtractValue(initializer, []uint32{0, 0})
			handlerFuncPtr := llvm.ConstExtractValue(initializer, []uint32{0, 1})
			if !handlerContext.IsConstant() || !handlerFuncPtr.IsConstant() {
				// This should have been checked already in the compiler.
				errs = append(errs, errorAt(handler, "func value must be constant"))
				continue
			}
			if !handlerFuncPtr.IsAConstantExpr().IsNil() && handlerFuncPtr.Opcode() == llvm.PtrToInt {
				// This is a ptrtoint: the IR was created for func lowering using a
				// switch statement.
				global := handlerFuncPtr.Operand(0)
				if global.IsAGlobalValue().IsNil() {
					errs = append(errs, errorAt(global, "internal error: expected a global for func lowering"))
					continue
				}
				if !strings.HasSuffix(global.Name(), "$withSignature") {
					errs = append(errs, errorAt(global, "internal error: func lowering global has unexpected name: "+global.Name()))
					continue
				}
				initializer := global.Initializer()
				ptrtoint := llvm.ConstExtractValue(initializer, []uint32{0})
				if ptrtoint.IsAConstantExpr().IsNil() || ptrtoint.Opcode() != llvm.PtrToInt {
					errs = append(errs, errorAt(global, "internal error: func lowering global has unexpected func ptr type"))
					continue
				}
				handlerFuncPtr = ptrtoint.Operand(0)
			}
			if handlerFuncPtr.Type().TypeKind() != llvm.PointerTypeKind || handlerFuncPtr.Type().ElementType().TypeKind() != llvm.FunctionTypeKind {
				errs = append(errs, errorAt(handler, "internal error: unexpected LLVM types in func value"))
				continue
			}

			// Fill the function declaration with the forwarding call.
			// In practice, the called function will often be inlined which avoids
			// the extra indirection.
			handlerFuncPtrType := llvm.PointerType(llvm.FunctionType(ctx.VoidType(), []llvm.Type{num.Type(), i8ptrType, i8ptrType}, false), handlerFuncPtr.Type().PointerAddressSpace())
			handlerFuncPtr = llvm.ConstBitCast(handlerFuncPtr, handlerFuncPtrType)
			builder.CreateCall(handlerFuncPtr, []llvm.Value{num, handlerContext, nullptr}, "")

			// Replace all ptrtoint uses of the global with the interrupt constant.
			// That can only now be safely done after the interrupt handler has been
			// created, doing it before the interrupt handler is created might
			// result in this interrupt handler being optimized away entirely.
			for _, user := range getUses(handler) {
				if user.IsAConstantExpr().IsNil() || user.Opcode() != llvm.PtrToInt {
					errs = append(errs, errorAt(handler, "internal error: expected a ptrtoint"))
					continue
				}
				user.ReplaceAllUsesWith(num)
			}

			// The runtime/interrput.handle struct can finally be removed.
			// It would probably be eliminated anyway by a globaldce pass but it's
			// better to do it now to be sure.
			handler.EraseFromParentAsGlobal()
		}

		// The wrapper function has no return value
		builder.CreateRetVoid()
	}
	// Create a dispatcher function that calls the appropriate interrupt handler
	// for each interrupt ID. This is used in the case of software vectoring.
	// The function looks like this:
	//     func callInterruptHandler(id int) {
	//         switch id {
	//         case IRQ_UART:
	//             interrupt.interruptHandler3()
	//         case IRQ_FOO:
	//             interrupt.interruptHandler7()
	//         default:
	//             // do nothing
	//     }
	if hasSoftwareVectoring {
		// Create a sorted list of interrupt vector IDs.
		ids := make([]int64, 0, len(softwareVector))
		for id := range softwareVector {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

		// Start creating the function body with the big switch.
		dispatcher := mod.NamedFunction("runtime.callInterruptHandler")
		entryBlock := ctx.AddBasicBlock(dispatcher, "entry")
		defaultBlock := ctx.AddBasicBlock(dispatcher, "default")
		builder.SetInsertPointAtEnd(entryBlock)
		interruptID := dispatcher.Param(0)
		sw := builder.CreateSwitch(interruptID, defaultBlock, len(ids))

		// Create a switch case for each interrupt ID that calls the appropriate
		// handler.
		for _, id := range ids {
			block := ctx.AddBasicBlock(dispatcher, "interrupt"+strconv.FormatInt(id, 10))
			builder.SetInsertPointAtEnd(block)
			builder.CreateCall(softwareVector[id], nil, "")
			builder.CreateRetVoid()
			sw.AddCase(llvm.ConstInt(interruptID.Type(), uint64(id), true), block)
		}

		// Create a default case that just returns.
		// Perhaps it is better to call some default interrupt handler here that
		// logs an error?
		builder.SetInsertPointAtEnd(defaultBlock)
		builder.CreateRetVoid()

		// Make sure the dispatcher is optimized.
		// Without this, it will probably not get inlined.
		dispatcher.SetLinkage(llvm.InternalLinkage)
		dispatcher.SetUnnamedAddr(true)
	}

	// Remove now-useless runtime/interrupt.use calls. These are used for some
	// platforms like AVR that do not need to enable interrupts to use them, so
	// need another way to keep them alive.
	// After interrupts have been lowered, this call is useless and would cause
	// a linker error so must be removed.
	for _, call := range getUses(mod.NamedFunction("runtime/interrupt.use")) {
		if call.IsACallInst().IsNil() {
			errs = append(errs, errorAt(call, "internal error: expected call to runtime/interrupt.use"))
			continue
		}

		call.EraseFromParentAsInstruction()
	}

	return errs
}

func packageFromInterruptHandle(handle llvm.Value) string {
	return strings.Split(handle.Name(), "$")[0]
}
