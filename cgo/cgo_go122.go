//go:build go1.22

package cgo

// Code specifically for Go 1.22.

import (
	"go/ast"
	"go/token"
)

func init() {
	setASTFileFields = func(f *ast.File, start, end token.Pos) {
		f.FileStart = start
		f.FileEnd = end
	}
}
