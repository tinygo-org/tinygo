package interp

// This file provides useful types for errors encountered during IR evaluation.

import (
	"errors"
	"go/scanner"
	"go/token"
	"path/filepath"

	"tinygo.org/x/go-llvm"
)

type recoverableError struct {
	msg string
	op  string // instruction involved, if any
	fn  string // function being interpreted, if known
}

func (e *recoverableError) Error() string {
	s := e.msg
	if e.op != "" {
		s += ": " + e.op
	}
	if e.fn != "" {
		s += " (in " + e.fn + ")"
	}
	return s
}

// These errors are expected during normal execution and can be recovered from
// by running the affected function at runtime instead of compile time.
var (
	errIntegerAsPointer = &recoverableError{msg: "interp: trying to use an integer as a pointer (memory-mapped I/O?)"}
	errUnsupportedInst  = &recoverableError{msg: "interp: unsupported instruction"}

	// This is one of the errors that can be returned from toLLVMValue when the
	// passed type does not fit the data to serialize. It is recoverable by
	// serializing without a type (using rawValue.rawLLVMValue).
	errInvalidPtrToIntSize = &recoverableError{msg: "interp: ptrtoint integer size does not equal pointer size"}
)

func isRecoverableError(err error) bool {
	var target *recoverableError
	return errors.As(err, &target)
}

// ErrorLine is one line in a traceback. The position may be missing.
type ErrorLine struct {
	Pos  token.Position
	Inst string

	// inst is printed into Inst by Error.materialize.
	inst llvm.Value
}

// Error encapsulates compile-time interpretation errors with an associated
// import path. The errors may not have a precise location attached.
type Error struct {
	ImportPath string
	Inst       string
	Pos        token.Position
	Err        error
	Traceback  []ErrorLine

	// inst is printed into Inst by materialize. Printing walks the whole
	// module, so errors that are recovered from are never printed.
	inst llvm.Value
}

// Error returns the string of the first error in the list of errors.
func (e *Error) Error() string {
	return e.Pos.String() + ": " + e.Err.Error()
}

// errorAt returns an error value for the currently interpreted package at the
// location of the instruction. The location information may not be complete as
// it depends on debug information in the IR.
func (r *runner) errorAt(inst instruction, err error) *Error {
	pos := getPosition(inst.llvmInst)
	return &Error{
		ImportPath: r.pkgName,
		Pos:        pos,
		Err:        err,
		Traceback:  []ErrorLine{{Pos: pos, inst: inst.llvmInst}},
		inst:       inst.llvmInst,
	}
}

// materialize fills in the printed instructions of e and its traceback. It
// must be called before the error is returned out of the interp package.
func (e *Error) materialize() *Error {
	if e.Inst == "" && !e.inst.IsNil() {
		e.Inst = e.inst.String()
	}
	for i := range e.Traceback {
		line := &e.Traceback[i]
		if line.Inst == "" && !line.inst.IsNil() {
			line.Inst = line.inst.String()
		}
	}
	return e
}

// errorAt returns an error value at the location of the instruction.
// The location information may not be complete as it depends on debug
// information in the IR.
func errorAt(inst llvm.Value, msg string) scanner.Error {
	return scanner.Error{
		Pos: getPosition(inst),
		Msg: msg,
	}
}

// getPosition returns the position information for the given instruction, as
// far as it is available.
func getPosition(inst llvm.Value) token.Position {
	if inst.IsAInstruction().IsNil() {
		return token.Position{}
	}
	loc := inst.InstructionDebugLoc()
	if loc.IsNil() {
		return token.Position{}
	}
	file := loc.LocationScope().ScopeFile()
	return token.Position{
		Filename: filepath.Join(file.FileDirectory(), file.FileFilename()),
		Line:     int(loc.LocationLine()),
		Column:   int(loc.LocationColumn()),
	}
}
