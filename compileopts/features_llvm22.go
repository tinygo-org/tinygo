//go:build llvm22

package compileopts

import "strings"

// patchFeatures applies LLVM-version-specific feature name mappings.
// LLVM 22 renamed several Xtensa target features.
func patchFeatures(features string) string {
	// Xtensa feature renames in LLVM 22:
	//   atomctl → (removed, no direct replacement)
	//   memctl  → (removed, no direct replacement)
	//   esp32s3 → esp32s3ops
	//   timerint → timers3 (for esp32/esp32s3) or timers1 (for esp8266)
	// Since we can't distinguish which timer variant at this level,
	// just remove the obsolete features. The CPU definition already
	// implies the correct features in LLVM 22.
	replacer := strings.NewReplacer(
		"+atomctl,", "",
		"+memctl,", "",
		"+esp32s3,", "+esp32s3ops,",
		"+timerint,", "",
		",+timerint", "",
	)
	return replacer.Replace(features)
}
