package runtime

func Callers(skip int, pc []uintptr) int {
	return 0
}

// buildVersion is the Tinygo tree's version string at build time.
//
// This is set by the linker.
var buildVersion string

// Version returns the Tinygo tree's version string.
// It is the same as goenv.Version, or in case of a development build,
// it will be the concatenation of goenv.Version and the git commit hash.
func Version() string {
	return buildVersion
}
