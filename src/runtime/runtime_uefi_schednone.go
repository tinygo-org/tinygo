//go:build uefi && scheduler.none

package runtime

//go:noinline
func runProgram() {
	runProgramInitRand()
	runProgramInitHeap()
	runProgramInitAll()
	runProgramCallMain()
	mainExited = true
}

//go:noinline
func runProgramInitRand() {
	initRand()
}

//go:noinline
func runProgramInitHeap() {
	initHeap()
}

//go:noinline
func runProgramInitAll() {
	initAll()
}

//go:noinline
func runProgramCallMain() {
	callMain()
}
