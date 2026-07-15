//go:build !llvm22 && !llvm14 && !llvm15 && !llvm16 && !llvm17 && !llvm18 && !llvm19

package compileopts

// patchFeatures applies LLVM-version-specific feature name mappings.
// For LLVM 20/21, features in the target JSON files are already correct.
func patchFeatures(features string) string {
	return features
}
