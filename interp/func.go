package interp

import (
	"fmt"
	"strings"

	"tinygo.org/x/go-llvm"
)

func (c *constParser) parseCallSignature(call llvm.Value) (signature, error) {
	// Get basic function type info.
	ty := call.CalledValue().Type()
	i, err := c.fnTyInfo(ty)
	if err != nil {
		return signature{}, err
	}

	// Check the calling convention.
	conv := call.InstructionCallConv()

	// TODO: check for an exception-handling personality (not available in bindings right now).

	// TODO: many attributes can significantly impact how a function behaves (e.g. returns_twice).
	// It would be nice if we could loop through the attributes rather than checking for specific ones.
	// That way, we could revert if we located something that we cannot understand.

	// TODO: the bindings only currently allow adding (but not checking) call site attributes.
	// This is potentially an issue if we do not know whether a call could be returns_twice.

	// Parse arguments.
	args := make([]sigArg, len(i.args))
	for i, t := range i.args {
		args[i] = sigArg{t: t}
	}

	return signature{
		ty:        ty,
		args:      args,
		ret:       i.ret,
		conv:      conv,
		addrSpace: i.addrSpace,
	}, nil
}

func (c *constParser) parseFuncSignature(fn llvm.Value) (signature, error) {
	// Get basic function type info.
	ty := fn.Type()
	i, err := c.fnTyInfo(ty)
	if err != nil {
		return signature{}, err
	}

	// Check the calling convention.
	conv := fn.FunctionCallConv()

	// Check for "GC".
	if fn.GC() != "" {
		return signature{}, todo("function with GC")
	}

	// TODO: check for an exception-handling personality (not available in bindings right now).

	// TODO: many attributes can significantly impact how a function behaves (e.g. returns_twice).
	// It would be nice if we could loop through the attributes rather than checking for specific ones.
	// That way, we could revert if we located something that we cannot understand.

	// Parse arguments.
	args := make([]sigArg, len(i.args))
	for i, t := range i.args {
		noCapture := !fn.GetEnumAttributeAtIndex(i, attrNoCapture).IsNil()
		readNone := !fn.GetEnumAttributeAtIndex(i, attrReadNone).IsNil()
		readOnly := !fn.GetEnumAttributeAtIndex(i, attrReadOnly).IsNil()
		writeOnly := !fn.GetEnumAttributeAtIndex(i, attrWriteOnly).IsNil()
		if readOnly && writeOnly {
			readNone = true
			readOnly = false
			writeOnly = false
		}
		args[i] = sigArg{
			t:         t,
			noCapture: noCapture,
			readNone:  readNone,
			readOnly:  readOnly,
			writeOnly: writeOnly,
		}
	}

	// Parse function attributes.
	noSync := !fn.GetEnumFunctionAttribute(attrNoSync).IsNil()
	argMemOnly := !fn.GetEnumFunctionAttribute(attrArgMemOnly).IsNil()
	readNone := !fn.GetEnumFunctionAttribute(attrReadNone).IsNil()
	readOnly := !fn.GetEnumFunctionAttribute(attrReadOnly).IsNil()
	writeOnly := !fn.GetEnumFunctionAttribute(attrWriteOnly).IsNil()
	if readOnly && writeOnly {
		readNone = true
		readOnly = false
		writeOnly = false
	}
	noInline := !fn.GetEnumFunctionAttribute(attrNoInline).IsNil()

	return signature{
		ty:         ty,
		args:       args,
		ret:        i.ret,
		conv:       conv,
		addrSpace:  i.addrSpace,
		noSync:     noSync,
		argMemOnly: argMemOnly,
		readNone:   readNone,
		readOnly:   readOnly,
		writeOnly:  writeOnly,
		noInline:   noInline,
	}, nil
}

func (c *constParser) fnTyInfo(t llvm.Type) (fnTyInfo, error) {
	if i, ok := c.fCache[t]; ok {
		return i, nil
	}

	i, err := c.parseFnTyInfo(t)
	if err != nil {
		return fnTyInfo{}, err
	}

	c.fCache[t] = i
	return i, nil
}

func (c *constParser) parseFnTyInfo(t llvm.Type) (fnTyInfo, error) {
	spc := t.PointerAddressSpace()
	t = t.ElementType()
	ret, err := c.typ(t.ReturnType())
	if err != nil {
		return fnTyInfo{}, err
	}
	rawArgTypes := t.ParamTypes()
	args := make([]typ, len(rawArgTypes))
	for i, t := range rawArgTypes {
		args[i], err = c.typ(t)
		if err != nil {
			return fnTyInfo{}, err
		}
	}
	return fnTyInfo{args, ret, addrSpace(spc)}, nil
}

type fnTyInfo struct {
	args      []typ
	ret       typ
	addrSpace addrSpace
}

type signature struct {
	// ty is the raw LLVM type associated with the signature.
	ty llvm.Type

	// args are the arguments to the function (before decomposition).
	args []sigArg

	// ret is the type of the function return.
	ret typ

	// conv is the calling convention of the function.
	conv llvm.CallConv

	// addrSpace is the address space containing the function.
	addrSpace addrSpace

	// noSync indicates that the function does not synchronize.
	// This can constrain side-effects when combined with some other attributes.
	noSync bool

	// argMemOnly indicates that the function only accesses argument memory.
	argMemOnly bool

	// readNone indicates that this function does not access any memory.
	// This name is somewhat misleading, as it also implies that no writes are done.
	readNone bool

	// readOnly indicates that this function does not modify any memory.
	readOnly bool

	// writeOnly indicates that this function does not read any memory.
	writeOnly bool

	// noInline indicates that this function must not be inlined.
	// This blocks interp from processing it.
	noInline bool
}

func (s signature) merge(with signature) (signature, error) {
	if s.conv != with.conv {
		return signature{}, fmt.Errorf("calling convention mismatch: expected %s but got %s", callConvStr(with.conv), callConvStr(s.conv))
	}
	if s.ret != with.ret {
		return signature{}, fmt.Errorf("return type mismatch: %w", typeError{with.ret, s.ret})
	}
	if len(s.args) != len(with.args) {
		return signature{}, fmt.Errorf("function with %d args is not compatible with a signature with %d args", len(s.args), len(with.args))
	}
	args := make([]sigArg, len(s.args))
	for i, a := range s.args {
		arg, err := a.merge(with.args[i])
		if err != nil {
			return signature{}, fmt.Errorf("argument %d incompatible: %w", i, err)
		}
		args[i] = arg
	}
	if s.addrSpace != with.addrSpace {
		return signature{}, fmt.Errorf("address space mismatch: expected %s but got %s", with.addrSpace.String(), s.addrSpace.String())
	}
	noSync := s.noSync || with.noSync
	argMemOnly := s.argMemOnly || with.argMemOnly
	readNone := s.readNone || with.readNone
	readOnly := s.readOnly || with.readOnly
	writeOnly := s.writeOnly || with.writeOnly
	if readOnly && writeOnly {
		readNone = true
		readOnly = false
		writeOnly = false
	}
	noInline := s.noInline || with.noInline
	return signature{
		ty:         s.ty,
		args:       args,
		ret:        s.ret,
		conv:       s.conv,
		addrSpace:  s.addrSpace,
		noSync:     noSync,
		argMemOnly: argMemOnly,
		readNone:   readNone,
		readOnly:   readOnly,
		writeOnly:  writeOnly,
		noInline:   noInline,
	}, nil
}

func (s signature) String() string {
	args := make([]string, len(s.args))
	var inHeight uint
	for i, a := range s.args {
		var v value
		v, inHeight = createDecomposedIndices(inHeight, a.t)
		args[i] = v.String()
	}
	attrs := ""
	if s.noSync {
		attrs += " nosync"
	}
	if s.argMemOnly {
		attrs += " argmemonly"
	}
	if s.readNone {
		attrs += " readnone"
	}
	if s.readOnly {
		attrs += " readonly"
	}
	if s.writeOnly {
		attrs += " writeonly"
	}
	var ret string
	if s.ret != nil {
		r, _ := createDecomposedIndices(0, s.ret)
		ret = r.String()
	}
	return "fn " + callConvStr(s.conv) + attrs + " (" + strings.Join(args, ", ") + ")->(" + ret + ") in " + s.addrSpace.String() + attrs
}

func callConvStr(cc llvm.CallConv) string {
	switch cc {
	case llvm.CCallConv:
		return "ccc"
	case llvm.FastCallConv:
		return "fastcc"
	case llvm.ColdCallConv:
		return "coldcc"
	default:
		return "unknowncc"
	}
}

type sigArg struct {
	// t is the type of the argument.
	t typ

	// noCapture indicates that the argument is not escaped by the function.
	noCapture bool

	// readNone indicates that this function does not access any memory from this argument.
	// This name is somewhat misleading, as it also implies that no writes are done.
	readNone bool

	// readOnly indicates that this function does not modify any memory from this argument.
	readOnly bool

	// writeOnly indicates that this function does not read any memory from this argument.
	writeOnly bool
}

func (a sigArg) merge(with sigArg) (sigArg, error) {
	if a.t != with.t {
		return sigArg{}, typeError{with.t, a.t}
	}
	noCapture := a.noCapture || with.noCapture
	readOnly := a.readOnly || with.readOnly
	writeOnly := a.writeOnly || with.writeOnly
	readNone := a.readNone || with.readNone
	if readOnly && writeOnly {
		readNone = true
		readOnly = false
		writeOnly = false
	}
	return sigArg{
		t:         a.t,
		noCapture: noCapture,
		readNone:  readNone,
		readOnly:  readOnly,
		writeOnly: writeOnly,
	}, nil
}

func (a sigArg) String() string {
	str := a.t.String()
	if a.noCapture {
		str += " nocapture"
	}
	if a.readNone {
		str += " readnone"
	}
	if a.readOnly {
		str += " readonly"
	}
	if a.writeOnly {
		str += " writeonly"
	}
	return str
}

var (
	attrNoCapture  = llvm.AttributeKindID("nocapture")
	attrReadNone   = llvm.AttributeKindID("readnone")
	attrReadOnly   = llvm.AttributeKindID("readonly")
	attrWriteOnly  = llvm.AttributeKindID("writeonly")
	attrNoSync     = llvm.AttributeKindID("nosync")
	attrArgMemOnly = llvm.AttributeKindID("argmemonly")
	attrNoInline   = llvm.AttributeKindID("noinline")
)
