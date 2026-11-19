//go:build socinit

package runtime

func hwInit0() {
	socInit0()
	boardInit0()
}

func hwInit1() {
	socInit1()
	boardInit1()
}

//go:weak
func socInit0() {}

//go:weak
func socInit1() {}

//go:weak
func boardInit0() {}

//go:weak
func boardInit1() {}
