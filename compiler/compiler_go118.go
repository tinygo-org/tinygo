//go:build go1.18
// +build go1.18

package compiler

// Workaround for Go 1.17 support. Should be removed once we drop Go 1.17
// support.

import "go/types"

func init() {
	typeParamUnderlyingType = func(t types.Type) types.Type {
		if t, ok := t.(*types.TypeParam); ok {
			return t.Underlying()
		}
		return t
	}
}
