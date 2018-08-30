package runtime

type Func struct {
}

func FuncForPC(pc uintptr) *Func {
	return nil
}

func (f *Func) Name() string {
	return ""
}

func Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	return 0, "", 0, false
}
