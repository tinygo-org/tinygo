//go:build atsamd21 || nrf51

package runtime

// This function is called at HardFault.
//
// For details, see:
// https://community.arm.com/developer/ip-products/system/f/embedded-forum/3257/debugging-a-cortex-m0-hard-fault
// https://blog.feabhas.com/2013/02/developing-a-generic-hard-fault-handler-for-arm-cortex-m3cortex-m4/
//
//export HardFault_Handler
func HardFault_Handler() {
	// Obtain the stack pointer as it was on entry to the HardFault. It contains
	// the registers that were pushed by the NVIC and that we can now read back
	// to print the PC value at the time of the hard fault, for example.
	sp := (*interruptStack)(llvm_sponentry())

	// Note: by reusing the string "panic: runtime error at " we save a little
	// bit in terms of code size as the string can be deduplicated.
	print("panic: runtime error at ", sp.PC, ": HardFault with sp=", sp)
	// TODO: try to find the cause of the hard fault. Especially on Cortex-M3
	// and higher it is possible to find more detailed information in special
	// status registers.
	println()
	abort()
}
