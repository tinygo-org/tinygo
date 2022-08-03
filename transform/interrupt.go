package transform

import (
	"fmt"
	"strings"

	"tinygo.org/x/go-llvm"
)

// LowerInterrupts creates interrupt handlers for the interrupts created by
// runtime/interrupt.New.
//
// The operation is as follows. The compiler creates the following during IR
// generation:
//   - calls to runtime/interrupt.callHandlers with an interrupt number.
//   - runtime/interrupt.handle objects that store the (constant) interrupt ID and
//     interrupt handler func value.
//
// This pass then replaces those callHandlers calls with calls to the actual
// interrupt handlers. If there are no interrupt handlers for the given call,
// the interrupt handler is removed. For hardware vectoring, that means that the
// entire function is removed. For software vectoring, that means that the call
// is replaced with an 'unreachable' instruction.
// This might seem like it causes extra overhead, but in fact inlining and const
// propagation will eliminate most if not all of that.
func LowerInterrupts(mod llvm.Module) []error {
	var errs []error

	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	// Collect a map of interrupt handle objects. The fact that they still
	// exist in the IR indicates that they could not be optimized away,
	// therefore we need to make real interrupt handlers for them.
	handleMap := map[int64][]llvm.Value{}
	handleType := mod.GetTypeByName("runtime/interrupt.handle")
	if !handleType.IsNil() {
		handlePtrType := llvm.PointerType(handleType, 0)
		for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
			if global.Type() != handlePtrType {
				continue
			}

			// Get the interrupt number from the initializer
			initializer := global.Initializer()
			num := llvm.ConstExtractValue(initializer, []uint32{2, 0}).SExtValue()
			pkg := packageFromInterruptHandle(global)

			handles, exists := handleMap[num]

			// If there is an existing interrupt handler, ensure it is in the same package
			// as the new one.  This is to prevent any assumptions in code that a single
			// compiler pass can see all packages to chain interrupt handlers. When packages are
			// compiled to separate object files, the linker should spot the duplicate symbols
			// for the wrapper function, failing the build.
			if exists && packageFromInterruptHandle(handles[0]) != pkg {
				errs = append(errs, errorAt(global,
					fmt.Sprintf("handlers for interrupt %d in multiple packages: %s and %s",
						num, pkg, packageFromInterruptHandle(handles[0]))))
				continue
			}

			handleMap[num] = append(handles, global)
		}
	}

	// Discover interrupts. The runtime/interrupt.callHandlers call is a
	// compiler intrinsic that is replaced with the handlers for the given
	// function.
	for _, call := range getUses(mod.NamedFunction("runtime/interrupt.callHandlers")) {
		if call.IsACallInst().IsNil() {
			errs = append(errs, errorAt(call, "expected a call to runtime/interrupt.callHandlers?"))
			continue
		}

		num := call.Operand(0)
		if num.IsAConstantInt().IsNil() {
			errs = append(errs, errorAt(call, "non-constant interrupt number?"))
			call.InstructionParent().Parent().Dump()
			continue
		}

		handlers := handleMap[num.SExtValue()]
		if len(handlers) != 0 {
			// This interrupt has at least one handler.
			// Replace the callHandlers call with (possibly multiple) calls to
			// these handlers.
			builder.SetInsertPointBefore(call)
			for _, handler := range handlers {
				initializer := handler.Initializer()
				context := llvm.ConstExtractValue(initializer, []uint32{0})
				funcPtr := llvm.ConstExtractValue(initializer, []uint32{1}).Operand(0)
				builder.CreateCall(funcPtr, []llvm.Value{
					num,
					context,
				}, "")
			}
			call.EraseFromParentAsInstruction()
		} else {
			// No handlers. Remove the call.
			fn := call.InstructionParent().Parent()
			if fn.Linkage() == llvm.ExternalLinkage {
				// Hardware vectoring. Remove the function entirely (redirecting
				// it to the default handler).
				fn.ReplaceAllUsesWith(llvm.Undef(fn.Type()))
				fn.EraseFromParentAsFunction()
			} else {
				// Software vectoring. Erase the instruction and replace it with
				// 'unreachable'.
				builder.SetInsertPointBefore(call)
				builder.CreateUnreachable()
				// Erase all instructions that follow the unreachable
				// instruction (which is a block terminator).
				inst := call
				for !inst.IsNil() {
					next := llvm.NextInstruction(inst)
					inst.EraseFromParentAsInstruction()
					inst = next
				}
			}
		}
	}

	// Replace all ptrtoint uses of the interrupt handler globals with the real
	// interrupt ID.
	// This can now be safely done after interrupts have been lowered, doing it
	// earlier may result in this interrupt handler being optimized away
	// entirely (which is not what we want).
	for num, handlers := range handleMap {
		for _, handler := range handlers {
			for _, user := range getUses(handler) {
				if user.IsAConstantExpr().IsNil() || user.Opcode() != llvm.PtrToInt {
					errs = append(errs, errorAt(handler, "internal error: expected a ptrtoint"))
					continue
				}
				user.ReplaceAllUsesWith(llvm.ConstInt(user.Type(), uint64(num), true))
			}

			// The runtime/interrput.handle struct can finally be removed.
			// It would probably be eliminated anyway by a globaldce pass but it's
			// better to do it now to be sure.
			handler.EraseFromParentAsGlobal()
		}
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
