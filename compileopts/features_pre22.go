//go:build !llvm22

package compileopts

// patchFeatures applies LLVM-version-specific feature name mappings.
// For pre-LLVM 22, no patching is needed.
func patchFeatures(features string) string {
	return features
}
