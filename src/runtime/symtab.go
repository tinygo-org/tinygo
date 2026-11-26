package runtime

type Frames struct {
	//
}

type Frame struct {
	PC uintptr

	Func *Func

	Function string

	File string
	Line int

	Entry uintptr
}

func CallersFrames(callers []uintptr) *Frames {
	return nil
}

func (ci *Frames) Next() (frame Frame, more bool) {
	return Frame{}, false
}

func Callers(skip int, pc []uintptr) int {
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
