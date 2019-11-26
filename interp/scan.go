package interp

import (
	"tinygo.org/x/go-llvm"
)

type sideEffectSeverity int

func (severity sideEffectSeverity) String() string {
	switch severity {
	case sideEffectInProgress:
		return "in progress"
	case sideEffectNone:
		return "none"
	case sideEffectLimited:
		return "limited"
	case sideEffectAll:
		return "all"
	default:
		return "unknown"
	}
}

const (
	sideEffectInProgress sideEffectSeverity = iota // computing side effects is in progress (for recursive functions)
	sideEffectNone                                 // no side effects at all (pure)
	sideEffectLimited                              // has side effects, but the effects are known
	sideEffectAll                                  // has unknown side effects
)

// sideEffectResult contains the scan results after scanning a function for side
// effects (recursively).
type sideEffectResult struct {
	severity        sideEffectSeverity
	mentionsGlobals map[llvm.Value]struct{}
}

// hasSideEffects scans this function and all descendants, recursively. It
// returns whether this function has side effects and if it does, which globals
// it mentions anywhere in this function or any called functions.
func (e *evalPackage) hasSideEffects(fn llvm.Value) (*sideEffectResult, error) {
	switch fn.Name() {
	case "runtime.alloc":
		// Cannot be scanned but can be interpreted.
		return &sideEffectResult{severity: sideEffectNone}, nil
	case "runtime.nanotime":
		// Fixed value at compile time.
		return &sideEffectResult{severity: sideEffectNone}, nil
	case "runtime._panic":
		return &sideEffectResult{severity: sideEffectLimited}, nil
	case "runtime.interfaceImplements":
		return &sideEffectResult{severity: sideEffectNone}, nil
	case "runtime.sliceCopy":
		return &sideEffectResult{severity: sideEffectNone}, nil
	case "runtime.trackPointer":
		return &sideEffectResult{severity: sideEffectNone}, nil
	case "llvm.dbg.value":
		return &sideEffectResult{severity: sideEffectNone}, nil
	}
	if fn.IsDeclaration() {
		return &sideEffectResult{severity: sideEffectLimited}, nil
	}
	if e.sideEffectFuncs == nil {
		e.sideEffectFuncs = make(map[llvm.Value]*sideEffectResult)
	}
	if se, ok := e.sideEffectFuncs[fn]; ok {
		return se, nil
	}
	result := &sideEffectResult{
		severity:        sideEffectInProgress,
		mentionsGlobals: map[llvm.Value]struct{}{},
	}
	e.sideEffectFuncs[fn] = result
	dirtyLocals := map[llvm.Value]struct{}{}
	for bb := fn.EntryBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			if inst.IsAInstruction().IsNil() {
				// Should not happen in valid IR.
				panic("not an instruction")
			}

			// Check for any globals mentioned anywhere in the function. Assume
			// any mentioned globals may be read from or written to when
			// executed, thus must be marked dirty with a call.
			for i := 0; i < inst.OperandsCount(); i++ {
				operand := inst.Operand(i)
				if !operand.IsAGlobalVariable().IsNil() {
					result.mentionsGlobals[operand] = struct{}{}
				}
			}

			switch inst.InstructionOpcode() {
			case llvm.IndirectBr, llvm.Invoke:
				// Not emitted by the compiler.
				return nil, e.errorAt(inst, "unknown instructions")
			case llvm.Call:
				child := inst.CalledValue()
				if !child.IsAInlineAsm().IsNil() {
					// Inline assembly. This most likely has side effects.
					// Assume they're only limited side effects, similar to
					// external function calls.
					result.updateSeverity(sideEffectLimited)
					continue
				}
				if child.IsAFunction().IsNil() {
					// Indirect call?
					// In any case, we can't know anything here about what it
					// affects exactly so mark this function as invoking all
					// possible side effects.
					result.updateSeverity(sideEffectAll)
					continue
				}
				if child.IsDeclaration() {
					// External function call. Assume only limited side effects
					// (no affected globals, etc.).
					if e.hasLocalSideEffects(dirtyLocals, inst) {
						result.updateSeverity(sideEffectLimited)
					}
					continue
				}
				childSideEffects, err := e.hasSideEffects(child)
				if err != nil {
					return nil, err
				}
				switch childSideEffects.severity {
				case sideEffectInProgress, sideEffectNone:
					// no side effects or recursive function - continue scanning
				case sideEffectLimited:
					// The return value may be problematic.
					if e.hasLocalSideEffects(dirtyLocals, inst) {
						result.updateSeverity(sideEffectLimited)
					}
				case sideEffectAll:
					result.updateSeverity(sideEffectAll)
				default:
					panic("unreachable")
				}
			case llvm.Load:
				if inst.IsVolatile() {
					result.updateSeverity(sideEffectLimited)
				}
				if _, ok := e.dirtyGlobals[inst.Operand(0)]; ok {
					if e.hasLocalSideEffects(dirtyLocals, inst) {
						result.updateSeverity(sideEffectLimited)
					}
				}
			case llvm.Store:
				if inst.IsVolatile() {
					result.updateSeverity(sideEffectLimited)
				}
			case llvm.IntToPtr:
				// Pointer casts are not yet supported.
				result.updateSeverity(sideEffectLimited)
			default:
				// Ignore most instructions.
				// Check this list for completeness:
				// https://godoc.org/github.com/llvm-mirror/llvm/bindings/go/llvm#Opcode
			}
		}
	}

	if result.severity == sideEffectInProgress {
		// No side effect was reported for this function.
		result.severity = sideEffectNone
	}
	return result, nil
}

// hasLocalSideEffects checks whether the given instruction flows into a branch
// or return instruction, in which case the whole function must be marked as
// having side effects and be called at runtime.
func (e *Eval) hasLocalSideEffects(dirtyLocals map[llvm.Value]struct{}, inst llvm.Value) bool {
	if _, ok := dirtyLocals[inst]; ok {
		// It is already known that this local is dirty.
		return true
	}

	for use := inst.FirstUse(); !use.IsNil(); use = use.NextUse() {
		user := use.User()
		if user.IsAInstruction().IsNil() {
			// Should not happen in valid IR.
			panic("user not an instruction")
		}
		switch user.InstructionOpcode() {
		case llvm.Br, llvm.Switch:
			// A branch on a dirty value makes this function dirty: it cannot be
			// interpreted at compile time so has to be run at runtime. It is
			// marked as having side effects for this reason.
			return true
		case llvm.Ret:
			// This function returns a dirty value so it is itself marked as
			// dirty to make sure it is called at runtime.
			return true
		case llvm.Store:
			ptr := user.Operand(1)
			if !ptr.IsAGlobalVariable().IsNil() {
				// Store to a global variable.
				// Already handled in (*Eval).hasSideEffects.
				continue
			}
			// This store might affect all kinds of values. While it is
			// certainly possible to traverse through all of them, the easiest
			// option right now is to just assume the worst and say that this
			// function has side effects.
			// TODO: traverse through all stores and mark all relevant allocas /
			// globals dirty.
			return true
		default:
			// All instructions that take 0 or more operands (1 or more if it
			// was a use) and produce a result.
			// For a list:
			// https://godoc.org/github.com/llvm-mirror/llvm/bindings/go/llvm#Opcode
			dirtyLocals[user] = struct{}{}
			if e.hasLocalSideEffects(dirtyLocals, user) {
				return true
			}
		}
	}

	// No side effects found.
	return false
}

// updateSeverity sets r.severity to the max of r.severity and severity,
// conservatively assuming the worst severity.
func (r *sideEffectResult) updateSeverity(severity sideEffectSeverity) {
	if severity > r.severity {
		r.severity = severity
	}
}

// updateSeverity updates the severity with the severity of the child severity,
// like in a function call. This means it also copies the mentioned globals.
func (r *sideEffectResult) update(child *sideEffectResult) {
	r.updateSeverity(child.severity)
	for global := range child.mentionsGlobals {
		r.mentionsGlobals[global] = struct{}{}
	}
}
