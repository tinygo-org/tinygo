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
	funcValueWithSignaturePtr := llvm.PointerType(c.mod.GetTypeByName("runtime.funcValueWithSignature"), 0)
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
				//     rawPtr := runtime.getFuncPtr(fn)
				//     if func.rawPtr == nil {
				//         runtime.nilpanic()
				//     }
				//     result := func.rawPtr(...args, func.context)
				// into this:
				//     if false {
				//         runtime.nilpanic()
				//     }
				//     var result // Phi
				//     switch fn.id {
				//     case 0:
				//         runtime.nilpanic()
				//     case 1:
				//         result = call first implementation...
				//     case 2:
				//         result = call second implementation...
				//     default:
				//         unreachable
				//     }

				// Remove some casts, checks, and the old call which we're going
				// to replace.
				var funcCall llvm.Value
				for _, inttoptr := range getUses(getFuncPtrCall) {
					if inttoptr.IsAIntToPtrInst().IsNil() {
						panic("expected inttoptr")
					}
					for _, ptrUse := range getUses(inttoptr) {
						if !ptrUse.IsABitCastInst().IsNil() {
							for _, bitcastUse := range getUses(ptrUse) {
								if bitcastUse.IsACallInst().IsNil() || bitcastUse.CalledValue().Name() != "runtime.isnil" {
									panic("expected a call to runtime.isnil")
								}
								bitcastUse.ReplaceAllUsesWith(llvm.ConstInt(c.ctx.Int1Type(), 0, false))
								bitcastUse.EraseFromParentAsInstruction()
							}
							ptrUse.EraseFromParentAsInstruction()
						} else if !ptrUse.IsACallInst().IsNil() && ptrUse.CalledValue() == inttoptr {
							if !funcCall.IsNil() {
								panic("multiple calls on a single runtime.getFuncPtr")
							}
							funcCall = ptrUse
						} else {
							panic("unexpected getFuncPtrCall")
						}
					}
				}
				if funcCall.IsNil() {
					panic("expected exactly one call use of a runtime.getFuncPtr")
				}

				// The block that cannot be reached with correct funcValues (to
				// help the optimizer).
				c.builder.SetInsertPointBefore(funcCall)
				defaultBlock := llvm.AddBasicBlock(funcCall.InstructionParent().Parent(), "func.default")
				c.builder.SetInsertPointAtEnd(defaultBlock)
				c.builder.CreateUnreachable()

				// Create the switch.
				c.builder.SetInsertPointBefore(funcCall)
				sw := c.builder.CreateSwitch(funcID, defaultBlock, len(functions)+1)

				// Split right after the switch. We will need to insert a few
				// basic blocks in this gap.
				nextBlock := c.splitBasicBlock(sw, llvm.NextBasicBlock(sw.InstructionParent()), "func.next")

				// The 0 case, which is actually a nil check.
				nilBlock := llvm.InsertBasicBlock(nextBlock, "func.nil")
				c.builder.SetInsertPointAtEnd(nilBlock)
				c.createRuntimeCall("nilpanic", nil, "")
				c.builder.CreateUnreachable()
				sw.AddCase(llvm.ConstInt(c.uintptrType, 0, false), nilBlock)

				// Gather the list of parameters for every call we're going to
				// make.
				callParams := make([]llvm.Value, funcCall.OperandsCount()-1)
				for i := range callParams {
					callParams[i] = funcCall.Operand(i)
				}

				// If the call produces a value, we need to get it using a PHI
				// node.
				phiBlocks := make([]llvm.BasicBlock, len(functions))
				phiValues := make([]llvm.Value, len(functions))
				for i, fn := range functions {
					// Insert a switch case.
					bb := llvm.InsertBasicBlock(nextBlock, "func.call"+strconv.Itoa(fn.id))
					c.builder.SetInsertPointAtEnd(bb)
					result := c.builder.CreateCall(fn.funcPtr, callParams, "")
					c.builder.CreateBr(nextBlock)
					sw.AddCase(llvm.ConstInt(c.uintptrType, uint64(fn.id), false), bb)
					phiBlocks[i] = bb
					phiValues[i] = result
				}
				// Create the PHI node so that the call result flows into the
				// next block (after the split). This is only necessary when the
				// call produced a value.
				if funcCall.Type().TypeKind() != llvm.VoidTypeKind {
					c.builder.SetInsertPointBefore(nextBlock.FirstInstruction())
					phi := c.builder.CreatePHI(funcCall.Type(), "")
					phi.AddIncoming(phiValues, phiBlocks)
					funcCall.ReplaceAllUsesWith(phi)
				}

				// Finally, remove the old instructions.
				funcCall.EraseFromParentAsInstruction()
				for _, inttoptr := range getUses(getFuncPtrCall) {
					inttoptr.EraseFromParentAsInstruction()
				}
				getFuncPtrCall.EraseFromParentAsInstruction()
			}
		}
	}
}
