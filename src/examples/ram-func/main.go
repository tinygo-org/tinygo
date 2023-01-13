package main

// This example demonstrates how to use go:section to place code into RAM for
// execution.  The code is present in flash in the `.data` region and copied
// into the correct place in RAM early in startup sequence (at the same time
// as non-zero global variables are initialized).
//
// This example should work on any ARM Cortex MCU.
//
// For Go code use the pragma "//go:section", for cgo use the "section" and
// "noinline" attributes.  The `.ramfuncs` section is explicitly placed into
// the `.data` region by the linker script.
//
// Running the example should print out the program counter from the functions
// below.  The program counters should be in different memory regions.
//
// On RP2040, for example, the output is something like this:
//
//  Go in RAM:    0x20000DB4
//  Go in flash:  0x10007610
//  cgo in RAM:   0x20000DB8
//  cgo in flash: 0x10002C26
//
// This can be confirmed using `objdump -t xxx.elf | grep main | sort`:
//
//  00000000 l    df *ABS*  00000000 main
//  1000760d l     F .text  00000004 main.in_flash
//  10007611 l     F .text  0000000c __Thumbv6MABSLongThunk_main.in_ram
//  1000761d l     F .text  0000000c __Thumbv6MABSLongThunk__Cgo_static_eea7585d7291176ad3bb_main_c_in_ram
//  1000bdb5 l     O .text  00000013 main$string
//  1000bdc8 l     O .text  00000013 main$string.1
//  1000bddb l     O .text  00000013 main$string.2
//  1000bdee l     O .text  00000013 main$string.3
//  20000db1 l     F .data  00000004 main.in_ram
//  20000db5 l     F .data  00000004 _Cgo_static_eea7585d7291176ad3bb_main_c_in_ram
//

import (
	"device"
	"fmt"
	"time"
	_ "unsafe" // unsafe is required for "//go:section"
)

/*
	#define ram_func __attribute__((section(".ramfuncs"),noinline))

	static ram_func void* main_c_in_ram() {
		void* p = 0;

		asm(
			"MOV %0, PC"
			: "=r"(p)
		);

		return p;
	}

	static void* main_c_in_flash() {
		void* p = 0;

		asm(
			"MOV %0, PC"
			: "=r"(p)
		);

		return p;
	}
*/
import "C"

func main() {
	time.Sleep(2 * time.Second)

	fmt.Printf("Go in RAM:    0x%X\n", in_ram())
	fmt.Printf("Go in flash:  0x%X\n", in_flash())
	fmt.Printf("cgo in RAM:   0x%X\n", C.main_c_in_ram())
	fmt.Printf("cgo in flash: 0x%X\n", C.main_c_in_flash())
}

//go:section .ramfuncs
func in_ram() uintptr {
	return device.AsmFull("MOV {}, PC", nil)
}

// go:noinline used here to prevent function being 'inlined' into main()
// so it appears in objdump output.  In normal use, go:inline is not
// required for functions running from flash (flash is the default).
//
//go:noinline
func in_flash() uintptr {
	return device.AsmFull("MOV {}, PC", nil)
}
