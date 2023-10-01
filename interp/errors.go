package interp

// This file provides useful types for errors encountered during IR evaluation.

import (
	"errors"
	"go/scanner"
	"go/token"
	"path/filepath"

	"tinygo.org/x/go-llvm"
)

// These errors are expected during normal execution and can be recovered from
// by running the affected function at runtime instead of compile time.
var (
	errIntegerAsPointer       = errors.New("interp: trying to use an integer as a pointer (memory-mapped I/O?)")
	errUnsupportedInst        = errors.New("interp: unsupported instruction")
	errUnsupportedRuntimeInst = errors.New("interp: unsupported instruction (to be emitted at runtime)")
	errMapAlreadyCreated      = errors.New("interp: map already created")
	errLoopUnrolled           = errors.New("interp: loop unrolled")
	errDepthExceeded          = errors.New("interp: depth exceeded")
)

// This is one of the errors that can be returned from toLLVMValue when the
// passed type does not fit the data to serialize. It is recoverable by
// serializing without a type (using rawValue.rawLLVMValue).
var errInvalidPtrToIntSize = errors.New("interp: ptrtoint integer size does not equal pointer size")

func isRecoverableError(err error) bool {
	return err == errIntegerAsPointer || err == errUnsupportedInst ||
		err == errUnsupportedRuntimeInst || err == errMapAlreadyCreated ||
		err == errLoopUnrolled || err == errDepthExceeded
}

// ErrorLine is one line in a traceback. The position may be missing.
type ErrorLine struct {
	Pos  token.Position
	Inst string
}

// Error encapsulates compile-time interpretation errors with an associated
// import path. The errors may not have a precise location attached.
type Error struct {
	ImportPath string
	Inst       string
	Pos        token.Position
	Err        error
	Traceback  []ErrorLine
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
		Inst:       inst.llvmInst.String(),
		Pos:        pos,
		Err:        err,
		Traceback:  []ErrorLine{{pos, inst.llvmInst.String()}},
	}
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
