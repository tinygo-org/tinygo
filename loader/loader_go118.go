//go:build go1.18
// +build go1.18

package loader

// Workaround for Go 1.17 support. Should be removed once we drop Go 1.17
// support.

import (
	"go/ast"
	"go/types"
)

func init() {
	addInstances = func(info *types.Info) {
		info.Instances = make(map[*ast.Ident]types.Instance)
	}
}
