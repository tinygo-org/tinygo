package interp

// This file provides useful types for errors encountered during IR evaluation.

import (
	"errors"
	"go/scanner"
	"go/token"
	"path/filepath"

	"tinygo.org/x/go-llvm"
)

// errUnreachable is returned when an unreachable instruction is executed. This
// error should not be visible outside of the interp package.
var errUnreachable = &Error{Err: errors.New("interp: unreachable executed")}

// unsupportedInstructionError returns a new "unsupported instruction" error for
// the given instruction. It includes source location information, when
// available.
func (e *evalPackage) unsupportedInstructionError(inst llvm.Value) *Error {
	return e.errorAt(inst, errors.New("interp: unsupported instruction"))
}

// ErrorLine is one line in a traceback. The position may be missing.
type ErrorLine struct {
	Pos  token.Position
	Inst llvm.Value
}

// Error encapsulates compile-time interpretation errors with an associated
// import path. The errors may not have a precise location attached.
type Error struct {
	ImportPath string
	Inst       llvm.Value
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
func (e *evalPackage) errorAt(inst llvm.Value, err error) *Error {
	return &Error{
		ImportPath: e.packagePath,
		Pos:        getPosition(inst),
		Err:        err,
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
