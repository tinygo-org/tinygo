// +build cortexm

package runtime

import (
	"device/arm"
	"unsafe"
)

//go:extern _sbss
var _sbss unsafe.Pointer

//go:extern _ebss
var _ebss unsafe.Pointer

//go:extern _sdata
var _sdata unsafe.Pointer

//go:extern _sidata
var _sidata unsafe.Pointer

//go:extern _edata
var _edata unsafe.Pointer

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint32)(dst) = *(*uint32)(src)
		dst = unsafe.Pointer(uintptr(dst) + 4)
		src = unsafe.Pointer(uintptr(src) + 4)
	}
}

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see scheduler_cortexm.S that relies on the
// exact layout of this struct.
type calleeSavedRegs struct {
	r4  uintptr
	r5  uintptr
	r6  uintptr
	r7  uintptr
	r8  uintptr
	r9  uintptr
	r10 uintptr
	r11 uintptr
}

// prepareStartTask stores fn and args in some callee-saved registers that can
// then be used by the startTask function (implemented in assembly) to set up
// the initial stack pointer and initial argument with the pointer to the object
// with the goroutine start arguments.
func (r *calleeSavedRegs) prepareStartTask(fn, args uintptr) {
	r.r4 = fn
	r.r5 = args
}

func abort() {
	// disable all interrupts
	arm.DisableInterrupts()

	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

// The stack layout at the moment an interrupt occurs.
// Registers can be accessed if the stack pointer is cast to a pointer to this
// struct.
type interruptStack struct {
	R0  uintptr
	R1  uintptr
	R2  uintptr
	R3  uintptr
	R12 uintptr
	LR  uintptr
	PC  uintptr
	PSR uintptr
}

// This function is called at HardFault.
// Before this function is called, the stack pointer is reset to the initial
// stack pointer (loaded from addres 0x0) and the previous stack pointer is
// passed as an argument to this function. This allows for easy inspection of
// the stack the moment a HardFault occurs, but it means that the stack will be
// corrupted by this function and thus this handler must not attempt to recover.
//
// For details, see:
// https://community.arm.com/developer/ip-products/system/f/embedded-forum/3257/debugging-a-cortex-m0-hard-fault
// https://blog.feabhas.com/2013/02/developing-a-generic-hard-fault-handler-for-arm-cortex-m3cortex-m4/
//go:export handleFault
func handleFault(code int, sp *interruptStack) {
	// Only print the PC if it points into memory.
	// It may not point into memory during a stack overflow, so check that first
	// before accessing the stack.
	var pc uintptr
	if uintptr(unsafe.Pointer(&sp.PC)) >= 0x20000000 {
		pc = sp.PC
	}
	if faultHandler != nil {
		faultHandler(FaultInfo{
			PC:   sp.PC,
			SP:   uintptr(unsafe.Pointer(sp)),
			Code: code,
		})
	}
	print("fatal error: ")
	if uintptr(unsafe.Pointer(sp)) < 0x20000000 {
		print("stack overflow")
	} else {
		// TODO: try to find the cause of the fault. Especially on Cortex-M3 and
		// higher it is possible to find more detailed information in special
		// status registers.
		print("fault with exception=", code)
	}
	print(" sp=", sp)
	if pc != 0 {
		print(" pc=", pc)
	}
	println()
	abort()
}

// Implement memset for LLVM and compiler-rt.
//go:export memset
func libc_memset(ptr unsafe.Pointer, c byte, size uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = c
	}
}

// Implement memmove for LLVM and compiler-rt.
//go:export memmove
func libc_memmove(dst, src unsafe.Pointer, size uintptr) {
	memmove(dst, src, size)
}

// Implement memcpy for LLVM and compiler-rt.
//go:export memcpy
func libc_memcpy(dst, src unsafe.Pointer, size uintptr) {
	memcpy(dst, src, size)
}
