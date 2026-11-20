package transform

import (
	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// funcValueTargets finds the functions that an indirect call can end up in.
// Calling a func value or a bound method is an indirect call, and the
// function that is called is not known. It can only be a function that has its
// address taken and that has the type of the call, so those are the candidates.
type funcValueTargets struct {
	mod    llvm.Module
	byType map[llvm.Type][]llvm.Value
}

// contextCaptured returns whether the last argument of the indirect call, the
// context of a func value, can be kept by the called function, or by the
// caller of the call, after the call returns.
//
// A closure body only reads its context, but that is not enough to assume the
// context does not escape. A closure that captures a single variable by
// reference gets the address of that variable as its context, so a body that
// returns or stores &x does hand out its context. The called function can be
// any one of the candidates, so all of them must be known not to capture it.
func (t *funcValueTargets) contextCaptured(call llvm.Value) bool {
	if t.byType == nil {
		t.byType = make(map[llvm.Type][]llvm.Value)
		for fn := t.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
			if addressTaken(fn) {
				typ := fn.GlobalValueType()
				t.byType[typ] = append(t.byType[typ], fn)
			}
		}
	}
	candidates := t.byType[call.CalledFunctionType()]
	if len(candidates) == 0 {
		// The target is not in this module.
		return true
	}
	kindNoCapture := llvm.AttributeKindID(llvmutil.NoCaptureAttrName())
	kindReturned := llvm.AttributeKindID("returned")
	for _, fn := range candidates {
		n := fn.ParamsCount()
		if n == 0 || fn.IsDeclaration() {
			return true
		}
		index := n // attribute indexes of parameters start at 1
		// A function that only reads the address of its context, such as a
		// closure whose context is an integer packed into the pointer, cannot
		// keep the object (captures(address), LLVM 21 and newer).
		if !llvmutil.CapturesNoProvenance(fn.GetEnumAttributeAtIndex(index, kindNoCapture)) ||
			!fn.GetEnumAttributeAtIndex(index, kindReturned).IsNil() {
			return true
		}
	}
	return false
}

// addressTaken returns whether the function is used for anything but being
// the called function of a call.
func addressTaken(fn llvm.Value) bool {
	for _, use := range getUses(fn) {
		if use.IsACallInst().IsNil() || use.CalledValue() != fn {
			return true
		}
		// A call that also passes the function as an argument.
		for i := 0; i < use.OperandsCount()-1; i++ {
			if use.Operand(i) == fn {
				return true
			}
		}
	}
	return false
}
