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
var errUnreachable = errors.New("interp: unreachable executed")

// Unsupported is the specific error that is returned when an unsupported
// instruction is hit while trying to interpret all initializers.
type Unsupported struct {
	ImportPath string
	Inst       llvm.Value
	Pos        token.Position
}

func (e Unsupported) Error() string {
	// TODO: how to return the actual instruction string?
	// It looks like LLVM provides no function for that...
	return scanner.Error{
		Pos: e.Pos,
		Msg: "interp: unsupported instruction",
	}.Error()
}

// unsupportedInstructionError returns a new "unsupported instruction" error for
// the given instruction. It includes source location information, when
// available.
func (e *evalPackage) unsupportedInstructionError(inst llvm.Value) *Unsupported {
	return &Unsupported{
		ImportPath: e.packagePath,
		Inst:       inst,
		Pos:        getPosition(inst),
	}
}

// Error encapsulates compile-time interpretation errors with an associated
// import path. The errors may not have a precise location attached.
type Error struct {
	ImportPath string
	Errs       []scanner.Error
}

// Error returns the string of the first error in the list of errors.
func (e Error) Error() string {
	return e.Errs[0].Error()
}

// errorAt returns an error value for the currently interpreted package at the
// location of the instruction. The location information may not be complete as
// it depends on debug information in the IR.
func (e *evalPackage) errorAt(inst llvm.Value, msg string) Error {
	return Error{
		ImportPath: e.packagePath,
		Errs:       []scanner.Error{errorAt(inst, msg)},
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
