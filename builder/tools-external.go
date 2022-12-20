//go:build !byollvm

package builder

import "errors"

const hasBuiltinTools = false

// RunTool runs the given tool (such as clang).
//
// This version doesn't actually run the tool: TinyGo has not been compiled by
// statically linking to LLVM.
func RunTool(tool string, args ...string) error {
	return errors.New("cannot run tool: " + tool)
}
