package interp

// This file provides useful types for errors encountered during IR evaluation.

import (
	"github.com/aykevl/go-llvm"
)

type Unsupported struct {
	Inst llvm.Value
}

func (e Unsupported) Error() string {
	// TODO: how to return the actual instruction string?
	// It looks like LLVM provides no function for that...
	return "interp: unsupported instruction"
}
