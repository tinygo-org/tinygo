package runtime

func Callers(skip int, pc []uintptr) int {
	if len(pc) > 0 {
		// The testing package expects at least one caller in all cases.
		pc[0] = 0
		return 1
	}
	return 0
}

// buildVersion is the Tinygo tree's version string at build time.
//
// This is set by the linker.
var buildVersion string

// Version returns the Tinygo tree's version string.
// It is the same as goenv.Version().
func Version() string {
	return buildVersion
}
