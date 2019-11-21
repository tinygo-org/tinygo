package compiler

// This file contains some utility functions related to error handling.

import (
	"go/scanner"
	"go/token"
	"go/types"
	"path/filepath"

	"tinygo.org/x/go-llvm"
)

// makeError makes it easy to create an error from a token.Pos with a message.
func (c *compilerContext) makeError(pos token.Pos, msg string) types.Error {
	return types.Error{
		Fset: c.ir.Program.Fset,
		Pos:  pos,
		Msg:  msg,
	}
}

func (c *Compiler) addError(pos token.Pos, msg string) {
	c.diagnostics = append(c.diagnostics, c.makeError(pos, msg))
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

// getPosition returns the position information for the given value, as far as
// it is available.
func getPosition(val llvm.Value) token.Position {
	if !val.IsAInstruction().IsNil() {
		loc := val.InstructionDebugLoc()
		if loc.IsNil() {
			return token.Position{}
		}
		file := loc.LocationScope().ScopeFile()
		return token.Position{
			Filename: filepath.Join(file.FileDirectory(), file.FileFilename()),
			Line:     int(loc.LocationLine()),
			Column:   int(loc.LocationColumn()),
		}
	} else if !val.IsAFunction().IsNil() {
		loc := val.Subprogram()
		if loc.IsNil() {
			return token.Position{}
		}
		file := loc.ScopeFile()
		return token.Position{
			Filename: filepath.Join(file.FileDirectory(), file.FileFilename()),
			Line:     int(loc.SubprogramLine()),
		}
	} else {
		return token.Position{}
	}
}
