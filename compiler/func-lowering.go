package compiler

// This file lowers func values into their final form. This is necessary for
// funcValueSwitch, which needs full program analysis.

import (
	"sort"
	"strconv"

	"tinygo.org/x/go-llvm"
)

// funcSignatureInfo keeps information about a single signature and its uses.
type funcSignatureInfo struct {
	sig                     llvm.Value   // *uint8 to identify the signature
	funcValueWithSignatures []llvm.Value // slice of runtime.funcValueWithSignature
}

// funcWithUses keeps information about a single function used as func value and
// the assigned function ID. More commonly used functions are assigned a lower
// ID.
type funcWithUses struct {
	funcPtr  llvm.Value
	useCount int // how often this function is used in a func value
	id       int // assigned ID
}

// Slice to sort functions by their use counts, or else their name if they're
// used equally often.
type funcWithUsesList []*funcWithUses

func (l funcWithUsesList) Len() int { return len(l) }
func (l funcWithUsesList) Less(i, j int) bool {
	if l[i].useCount != l[j].useCount {
		// return the reverse: we want the highest use counts sorted first
		return l[i].useCount > l[j].useCount
	}
	iName := l[i].funcPtr.Name()
	jName := l[j].funcPtr.Name()
	return iName < jName
}
func (l funcWithUsesList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// LowerFuncValue lowers the runtime.funcValueWithSignature type and
// runtime.getFuncPtr function to their final form.
func (c *Compiler) LowerFuncValues() {
	if c.funcImplementation() != funcValueSwitch {
		return
	}

	// Find all func values used in the program with their signatures.
	funcValueWithSignaturePtr := llvm.PointerType(c.getLLVMRuntimeType("funcValueWithSignature"), 0)
	signatures := map[string]*funcSignatureInfo{}
	for global := c.mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if global.Type() != funcValueWithSignaturePtr {
			continue
		}
		sig := llvm.ConstExtractValue(global.Initializer(), []uint32{1})
		name := sig.Name()
		if info, ok := signatures[name]; ok {
			info.funcValueWithSignatures = append(info.funcValueWithSignatures, global)
		} else {
			signatures[name] = &funcSignatureInfo{
				sig:                     sig,
				funcValueWithSignatures: []llvm.Value{global},
			}
		}
	}

	// Sort the signatures, for deterministic execution.
	names := make([]string, 0, len(signatures))
	for name := range signatures {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		info := signatures[name]
		functions := make(funcWithUsesList, len(info.funcValueWithSignatures))
		for i, use := range info.funcValueWithSignatures {
			var useCount int
			for _, use2 := range getUses(use) {
				useCount += len(getUses(use2))
			}
			functions[i] = &funcWithUses{
				funcPtr:  llvm.ConstExtractValue(use.Initializer(), []uint32{0}).Operand(0),
				useCount: useCount,
			}
		}
		sort.Sort(functions)

		for i, fn := range functions {
			fn.id = i + 1
			for _, ptrtoint := range getUses(fn.funcPtr) {
				if ptrtoint.IsAConstantExpr().IsNil() || ptrtoint.Opcode() != llvm.PtrToInt {
					continue
				}
				for _, funcValueWithSignatureConstant := range getUses(ptrtoint) {
					for _, funcValueWithSignatureGlobal := range getUses(funcValueWithSignatureConstant) {
						for _, use := range getUses(funcValueWithSignatureGlobal) {
							if ptrtoint.IsAConstantExpr().IsNil() || ptrtoint.Opcode() != llvm.PtrToInt {
								panic("expected const ptrtoint")
							}
							use.ReplaceAllUsesWith(llvm.ConstInt(c.uintptrType, uint64(fn.id), false))
						}
					}
				}
			}
		}

		for _, getFuncPtrCall := range getUses(info.sig) {
			if getFuncPtrCall.IsACallInst().IsNil() {
				continue
			}
			if getFuncPtrCall.CalledValue().Name() != "runtime.getFuncPtr" {
				panic("expected all call uses to be runtime.getFuncPtr")
			}
			funcID := getFuncPtrCall.Operand(1)
			switch len(functions) {
			case 0:
				// There are no functions used in a func value that implement
				// this signature. The only possible value is a nil value.
				for _, inttoptr := range getUses(getFuncPtrCall) {
					if inttoptr.IsAIntToPtrInst().IsNil() {
						panic("expected inttoptr")
					}
					nilptr := llvm.ConstPointerNull(inttoptr.Type())
					inttoptr.ReplaceAllUsesWith(nilptr)
					inttoptr.EraseFromParentAsInstruction()
				}
				getFuncPtrCall.EraseFromParentAsInstruction()
			case 1:
				// There is exactly one function with this signature that is
				// used in a func value. The func value itself can be either nil
				// or this one function.
				c.builder.SetInsertPointBefore(getFuncPtrCall)
				zero := llvm.ConstInt(c.uintptrType, 0, false)
				isnil := c.builder.CreateICmp(llvm.IntEQ, funcID, zero, "")
				funcPtrNil := llvm.ConstPointerNull(functions[0].funcPtr.Type())
				funcPtr := c.builder.CreateSelect(isnil, funcPtrNil, functions[0].funcPtr, "")
				for _, inttoptr := range getUses(getFuncPtrCall) {
					if inttoptr.IsAIntToPtrInst().IsNil() {
						panic("expected inttoptr")
					}
					inttoptr.ReplaceAllUsesWith(funcPtr)
					inttoptr.EraseFromParentAsInstruction()
				}
				getFuncPtrCall.EraseFromParentAsInstruction()
			default:
				// There are multiple functions used in a func value that
				// implement this signature.
				// What we'll do is transform the following:
				//     rawPtr := runtime.getFuncPtr(func.ptr)
				//     if rawPtr == nil {
				//         runtime.nilPanic()
				//     }
				//     result := rawPtr(...args, func.context)
				// into this:
				//     if false {
				//         runtime.nilPanic()
				//     }
				//     var result // Phi
				//     switch fn.id {
				//     case 0:
				//         runtime.nilPanic()
				//     case 1:
				//         result = call first implementation...
				//     case 2:
				//         result = call second implementation...
				//     default:
				//         unreachable
				//     }

				// Remove some casts, checks, and the old call which we're going
				// to replace.
				for _, callIntPtr := range getUses(getFuncPtrCall) {
					if !callIntPtr.IsACallInst().IsNil() && callIntPtr.CalledValue().Name() == "runtime.makeGoroutine" {
						for _, inttoptr := range getUses(callIntPtr) {
							if inttoptr.IsAIntToPtrInst().IsNil() {
								panic("expected a inttoptr")
							}
							for _, use := range getUses(inttoptr) {
								c.addFuncLoweringSwitch(funcID, use, c.emitStartGoroutine, functions)
								use.EraseFromParentAsInstruction()
							}
							inttoptr.EraseFromParentAsInstruction()
						}
						callIntPtr.EraseFromParentAsInstruction()
						continue
					}
					if callIntPtr.IsAIntToPtrInst().IsNil() {
						panic("expected inttoptr")
					}
					for _, ptrUse := range getUses(callIntPtr) {
						if !ptrUse.IsABitCastInst().IsNil() {
							for _, bitcastUse := range getUses(ptrUse) {
								if bitcastUse.IsACallInst().IsNil() || bitcastUse.CalledValue().IsAFunction().IsNil() {
									panic("expected a call instruction")
								}
								switch bitcastUse.CalledValue().Name() {
								case "runtime.isnil":
									bitcastUse.ReplaceAllUsesWith(llvm.ConstInt(c.ctx.Int1Type(), 0, false))
									bitcastUse.EraseFromParentAsInstruction()
								default:
									panic("expected a call to runtime.isnil")
								}
							}
						} else if !ptrUse.IsACallInst().IsNil() && ptrUse.CalledValue() == callIntPtr {
							c.addFuncLoweringSwitch(funcID, ptrUse, func(funcPtr llvm.Value, params []llvm.Value) llvm.Value {
								return c.builder.CreateCall(funcPtr, params, "")
							}, functions)
						} else {
							panic("unexpected getFuncPtrCall")
						}
						ptrUse.EraseFromParentAsInstruction()
					}
					callIntPtr.EraseFromParentAsInstruction()
				}
				getFuncPtrCall.EraseFromParentAsInstruction()
			}
		}
	}
}

// addFuncLoweringSwitch creates a new switch on a function ID and inserts calls
// to the newly created direct calls. The funcID is the number to switch on,
// call is the call instruction to replace, and createCall is the callback that
// actually creates the new call. By changing createCall to something other than
// c.builder.CreateCall, instead of calling a function it can start a new
// goroutine for example.
func (c *Compiler) addFuncLoweringSwitch(funcID, call llvm.Value, createCall func(funcPtr llvm.Value, params []llvm.Value) llvm.Value, functions funcWithUsesList) {
	// The block that cannot be reached with correct funcValues (to help the
	// optimizer).
	c.builder.SetInsertPointBefore(call)
	defaultBlock := c.ctx.AddBasicBlock(call.InstructionParent().Parent(), "func.default")
	c.builder.SetInsertPointAtEnd(defaultBlock)
	c.builder.CreateUnreachable()

	// Create the switch.
	c.builder.SetInsertPointBefore(call)
	sw := c.builder.CreateSwitch(funcID, defaultBlock, len(functions)+1)

	// Split right after the switch. We will need to insert a few basic blocks
	// in this gap.
	nextBlock := c.splitBasicBlock(sw, llvm.NextBasicBlock(sw.InstructionParent()), "func.next")

	// The 0 case, which is actually a nil check.
	nilBlock := c.ctx.InsertBasicBlock(nextBlock, "func.nil")
	c.builder.SetInsertPointAtEnd(nilBlock)
	c.createRuntimeCall("nilPanic", nil, "")
	c.builder.CreateUnreachable()
	sw.AddCase(llvm.ConstInt(c.uintptrType, 0, false), nilBlock)

	// Gather the list of parameters for every call we're going to make.
	callParams := make([]llvm.Value, call.OperandsCount()-1)
	for i := range callParams {
		callParams[i] = call.Operand(i)
	}

	// If the call produces a value, we need to get it using a PHI
	// node.
	phiBlocks := make([]llvm.BasicBlock, len(functions))
	phiValues := make([]llvm.Value, len(functions))
	for i, fn := range functions {
		// Insert a switch case.
		bb := c.ctx.InsertBasicBlock(nextBlock, "func.call"+strconv.Itoa(fn.id))
		c.builder.SetInsertPointAtEnd(bb)
		result := createCall(fn.funcPtr, callParams)
		c.builder.CreateBr(nextBlock)
		sw.AddCase(llvm.ConstInt(c.uintptrType, uint64(fn.id), false), bb)
		phiBlocks[i] = bb
		phiValues[i] = result
	}
	// Create the PHI node so that the call result flows into the
	// next block (after the split). This is only necessary when the
	// call produced a value.
	if call.Type().TypeKind() != llvm.VoidTypeKind {
		c.builder.SetInsertPointBefore(nextBlock.FirstInstruction())
		phi := c.builder.CreatePHI(call.Type(), "")
		phi.AddIncoming(phiValues, phiBlocks)
		call.ReplaceAllUsesWith(phi)
	}
}
