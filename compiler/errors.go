package compiler

import (
	"go/token"
	"go/types"
)

func (c *Compiler) makeError(pos token.Pos, msg string) types.Error {
	return types.Error{
		Fset: c.ir.Program.Fset,
		Pos:  pos,
		Msg:  msg,
	}
}

func (c *Compiler) addError(pos token.Pos, msg string) {
	c.diagnostics = append(c.diagnostics, c.makeError(pos, msg))
}
