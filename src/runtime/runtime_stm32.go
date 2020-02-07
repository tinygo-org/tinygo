// +build stm32

package runtime

type timeUnit int64

func postinit() {}

//go:export Reset_Handler
func main() {
	preinit()
	run()
	abort()
}
