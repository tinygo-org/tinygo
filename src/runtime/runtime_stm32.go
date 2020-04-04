// +build stm32

package runtime

type timeUnit int64

func postinit() {}

//export Reset_Handler
func main() {
	preinit()
	run()
	abort()
}
