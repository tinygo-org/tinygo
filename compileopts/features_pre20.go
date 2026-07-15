//go:build llvm14 || llvm15 || llvm16 || llvm17 || llvm18 || llvm19

package compileopts

import "strings"

// patchFeatures applies LLVM-version-specific feature name mappings.
// LLVM 19 and earlier do not have +bulk-memory-opt or
// +call-indirect-overlong for WebAssembly (added in LLVM 20).
func patchFeatures(features string) string {
	features = strings.ReplaceAll(features, ",+bulk-memory-opt", "")
	features = strings.ReplaceAll(features, "+bulk-memory-opt,", "")
	features = strings.ReplaceAll(features, ",+call-indirect-overlong", "")
	features = strings.ReplaceAll(features, "+call-indirect-overlong,", "")
	return features
}
