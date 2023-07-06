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
}

func CallersFrames(callers []uintptr) *Frames {
	return nil
}

func (ci *Frames) Next() (frame Frame, more bool) {
	return Frame{}, false
}
