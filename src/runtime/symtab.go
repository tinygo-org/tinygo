package runtime

type Frames struct {
	//
}

type Frame struct {
	Function string

	File string
	Line int
	PC   uintptr
}

func CallersFrames(callers []uintptr) *Frames {
	return nil
}

func (ci *Frames) Next() (frame Frame, more bool) {
	return Frame{}, false
}
