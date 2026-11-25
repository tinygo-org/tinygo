package transform

import (
	"go/scanner"
	"go/token"
	"path/filepath"

	"tinygo.org/x/go-llvm"
)

// errorAt returns an error value at the location of the value.
// The location information may not be complete as it depends on debug
// information in the IR.
func errorAt(val llvm.Value, msg string) scanner.Error {
	return scanner.Error{
		Pos: getPosition(val),
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
