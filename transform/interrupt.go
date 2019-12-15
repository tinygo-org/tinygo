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

	create := mod.NamedFunction("runtime/interrupt.New")
	if create.IsNil() {
		// No interrupt handlers to create.
		return nil
	}

	ctx := mod.Context()
	nullptr := llvm.ConstNull(llvm.PointerType(ctx.Int8Type(), 0))
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	var dibuilder *llvm.DIBuilder

	// Create a function type with the signature of an interrupt handler.
	fnType := llvm.FunctionType(ctx.VoidType(), nil, false)

	for _, call := range getUses(create) {
		if call.IsACallInst().IsNil() {
			errs = append(errs, errorAt(call, "expected a call to runtime/interrupt.New"))
			continue
		}
		if call.OperandsCount() != 6 {
			// 6 params:
			// * 1 for the IRQ number
			// * 2 for the function
			// * 2 extra: context and parentHandle
			// * one more for the called value?
			errs = append(errs, errorAt(call, fmt.Sprintf("unexpected interrupt.New signature, expected 6 operands, got %d", call.OperandsCount())))
			continue
		}
		num := call.Operand(0)
		if num.IsAConstant().IsNil() {
			errs = append(errs, errorAt(call, "non-constant interrupt number"))
			continue
		}
		name := handlerNames[num.SExtValue()]
		if name == "" {
			errs = append(errs, errorAt(call, fmt.Sprintf("cannot find interrupt name for number %d", num.SExtValue())))
			continue
		}

		// Create the func value.
		handlerContext := call.Operand(1)
		handlerFuncPtr := call.Operand(2)
		if isFunctionLocal(handlerContext) || isFunctionLocal(handlerFuncPtr) {
			errs = append(errs, errorAt(call, "func value must be constant"))
			continue
		}
		if !handlerFuncPtr.IsAConstantExpr().IsNil() && handlerFuncPtr.Opcode() == llvm.PtrToInt {
			// This is a ptrtoint: the IR was created for func lowering using a
			// switch statement.
			global := handlerFuncPtr.Operand(0)
			if global.IsAGlobalValue().IsNil() {
				errs = append(errs, errorAt(call, "internal error: expected a global for func lowering"))
				continue
			}
			initializer := global.Initializer()
			if initializer.Type() != mod.GetTypeByName("runtime.funcValueWithSignature") {
				errs = append(errs, errorAt(call, "internal error: func lowering global has unexpected type"))
				continue
			}
			ptrtoint := llvm.ConstExtractValue(initializer, []uint32{0})
			if ptrtoint.IsAConstantExpr().IsNil() || ptrtoint.Opcode() != llvm.PtrToInt {
				errs = append(errs, errorAt(call, "internal error: func lowering global has unexpected func ptr type"))
				continue
			}
			handlerFuncPtr = ptrtoint.Operand(0)
		}
		if handlerFuncPtr.Type().TypeKind() != llvm.PointerTypeKind || handlerFuncPtr.Type().ElementType().TypeKind() != llvm.FunctionTypeKind {
			errs = append(errs, errorAt(call, "internal error: unexpected LLVM types in func value"))
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
			errs = append(errs, errorAt(call, name+" redeclared with a different signature"))
			continue
		} else if fn.IsDeclaration() {
			// Interrupt handler was already defined. Check the first
			// instruction (which should be a call) whether this handler would
			// be identical anyway.
			firstInst := fn.FirstBasicBlock().FirstInstruction()
			if !firstInst.IsACallInst().IsNil() && firstInst.OperandsCount() == 4 && firstInst.Operand(0) == num && firstInst.Operand(1) != handlerContext {
				// Already defined and apparently identical, so assume this is
				// fine.
				continue
			}

			errValue := name + " redeclared in this program"
			fnPos := getPosition(fn)
			if fnPos.IsValid() {
				errValue += "\n\tprevious declaration at " + fnPos.String()
			}
			errs = append(errs, errorAt(call, errValue))
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

		// Attach a proper debug location to this function.
		// Use the location of the call instruction, because as far as the user
		// is concerned that is where the interrupt is defined. This is
		// especially relevant for multiple definition errors, where you'd
		// really want the previous interrupt.New call location to be used for
		// easy debugging.
		loc := call.InstructionDebugLoc()
		if !loc.IsNil() {
			if dibuilder == nil {
				dibuilder = llvm.NewDIBuilder(mod)
				defer func() {
					dibuilder.Finalize()
					dibuilder.Destroy()
				}()
				// Must create a new compile unit for some reason.
				dibuilder.CreateCompileUnit(llvm.DICompileUnit{
					File:      "<interrupt-lowering>",
					Language:  0xb, // DW_LANG_C99 (0xc, off-by-one?)
					Producer:  "TinyGo",
					Optimized: true,
				})
			}

			// Attach debug info to the function.
			file := loc.LocationScope().ScopeFile()
			diFnType := dibuilder.CreateSubroutineType(llvm.DISubroutineType{
				File: file,
			})
			difunc := dibuilder.CreateFunction(file, llvm.DIFunction{
				Name:         name,
				LinkageName:  name,
				File:         file,
				Line:         int(loc.LocationLine()),
				Type:         diFnType,
				LocalToUnit:  false,
				IsDefinition: true,
				ScopeLine:    0,
				Flags:        llvm.FlagPrototyped,
				Optimized:    true,
			})
			fn.SetSubprogram(difunc)

			// Use the same location inside the thunk.
			builder.SetCurrentDebugLocation(loc.LocationLine(), loc.LocationColumn(), difunc, llvm.Metadata{})

		}

		// Fill the function declaration with the forwarding call.
		builder.CreateCall(handlerFuncPtr, []llvm.Value{num, handlerContext, nullptr}, "")
		builder.CreateRetVoid()

		// Replace the function call.
		interruptValue := llvm.ConstNamedStruct(mod.GetTypeByName("runtime/interrupt.Interrupt"), []llvm.Value{num})
		call.ReplaceAllUsesWith(interruptValue)
		call.EraseFromParentAsInstruction()
	}
	return errs
}
