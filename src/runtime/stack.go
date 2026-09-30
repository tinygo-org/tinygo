package runtime

type Func struct {
}

func FuncForPC(pc uintptr) *Func {
	return nil
}

func (f *Func) Name() string {
	return ""
}

func (f *Func) FileLine(pc uintptr) (file string, line int) {
	return "", 0
}

// Entry returns the entry address of the function. It is stubbed like the
// rest of runtime.Func, so that callers such as testify/mock compile.
func (f *Func) Entry() uintptr {
	return 0
}

func Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	return 0, "", 0, false
}

func Stack(buf []byte, all bool) int {
	return 0
}
