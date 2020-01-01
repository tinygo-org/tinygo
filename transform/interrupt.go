package transform

import (
	"fmt"
	"strings"

	"tinygo.org/x/go-llvm"
)

func LowerInterruptRegistrations(mod llvm.Module) []error {
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

	ctx := mod.Context()
	nullptr := llvm.ConstNull(llvm.PointerType(ctx.Int8Type(), 0))
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	// Create a function type with the signature of an interrupt handler.
	fnType := llvm.FunctionType(ctx.VoidType(), nil, false)

	handleType := mod.GetTypeByName("runtime/interrupt.handle")
	if handleType.IsNil() {
		// Nothing to do here.
		return errs
	}
	handlePtrType := llvm.PointerType(handleType, 0)
	var handlers []llvm.Value
	for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if global.Type() != handlePtrType {
			continue
		}
		handlers = append(handlers, global)
	}

	for _, global := range handlers {
		initializer := global.Initializer()
		num := llvm.ConstExtractValue(initializer, []uint32{1, 0})
		name := handlerNames[num.SExtValue()]

		if name == "" {
			errs = append(errs, errorAt(global, fmt.Sprintf("cannot find interrupt name for number %d", num.SExtValue())))
			continue
		}

		// Create the func value.
		handlerContext := llvm.ConstExtractValue(initializer, []uint32{0, 0})
		handlerFuncPtr := llvm.ConstExtractValue(initializer, []uint32{0, 1})
		if isFunctionLocal(handlerContext) || isFunctionLocal(handlerFuncPtr) {
			errs = append(errs, errorAt(global, "func value must be constant"))
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
			initializer := global.Initializer()
			if initializer.Type() != mod.GetTypeByName("runtime.funcValueWithSignature") {
				errs = append(errs, errorAt(global, "internal error: func lowering global has unexpected type"))
				continue
			}
			ptrtoint := llvm.ConstExtractValue(initializer, []uint32{0})
			if ptrtoint.IsAConstantExpr().IsNil() || ptrtoint.Opcode() != llvm.PtrToInt {
				errs = append(errs, errorAt(global, "internal error: func lowering global has unexpected func ptr type"))
				continue
			}
			handlerFuncPtr = ptrtoint.Operand(0)
		}
		if handlerFuncPtr.Type().TypeKind() != llvm.PointerTypeKind || handlerFuncPtr.Type().ElementType().TypeKind() != llvm.FunctionTypeKind {
			errs = append(errs, errorAt(global, "internal error: unexpected LLVM types in func value"))
			continue
		}

		// Check for an existing handler, and report it as an error if there is
		// one.
		fn := mod.NamedFunction(name)
		if fn.IsNil() {
			fn = llvm.AddFunction(mod, name, fnType)
		} else if fn.Type().ElementType() != fnType {
			// Don't bother with a precise error message (listing the
			// previsous location) because this should not normally happen
			// anyway.
			errs = append(errs, errorAt(global, name+" redeclared with a different signature"))
			continue
		} else if !fn.IsDeclaration() {
			// Interrupt handler was already defined. Check the first
			// instruction (which should be a call) whether this handler would
			// be identical anyway.
			firstInst := fn.FirstBasicBlock().FirstInstruction()
			if !firstInst.IsACallInst().IsNil() && firstInst.OperandsCount() == 4 && firstInst.CalledValue() == handlerFuncPtr && firstInst.Operand(0) == num && firstInst.Operand(1) == handlerContext {
				// Already defined and apparently identical, so assume this is
				// fine.
				continue
			}

			errValue := name + " redeclared in this program"
			fnPos := getPosition(fn)
			if fnPos.IsValid() {
				errValue += "\n\tprevious declaration at " + fnPos.String()
			}
			errs = append(errs, errorAt(global, errValue))
			continue
		}

		// Create the wrapper function.
		fn.SetUnnamedAddr(true)
		entryBlock := ctx.AddBasicBlock(fn, "entry")
		builder.SetInsertPointAtEnd(entryBlock)

		// Set the 'interrupt' flag if needed on this platform.
		if strings.HasPrefix(mod.Target(), "avr") {
			fn.SetFunctionCallConv(85) // CallingConv::AVR_SIGNAL
		}

		// Fill the function declaration with the forwarding call.
		builder.CreateCall(handlerFuncPtr, []llvm.Value{num, handlerContext, nullptr}, "")
		builder.CreateRetVoid()

		for _, user := range getUses(global) {
			if user.IsAConstantExpr().IsNil() || user.Opcode() != llvm.PtrToInt {
				errs = append(errs, errorAt(global, "internal error: expected a ptrtoint"))
				continue
			}
			user.ReplaceAllUsesWith(num)
		}
		global.EraseFromParentAsGlobal()
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
