package main

type wrapperValue struct{ n int }

func (v wrapperValue) get() int {
	return v.n
}

func boundWrapper(v wrapperValue) func() int {
	return v.get
}

func thunkWrapper() func(wrapperValue) int {
	return wrapperValue.get
}

func pointerWrapper(v *wrapperValue) int {
	var i interface{ get() int } = v
	return i.get()
}
