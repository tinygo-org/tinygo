.section ".text.boot_bl"

//
// The bootloader changes execution to user code in EL2.
//

// startup parameters are:
// x0 - start address (also in PC)
// x1 - stack ptr
// x2 - heap ptr
// x3 - unix time
.global _start

	msr     sp_el1, x1
	// enable CNTP for EL1
	mrs     x29, cnthctl_el2
	orr     x29, x29, #3
	msr     cnthctl_el2, x29
	msr     cntvoff_el2, xzr
	// enable AArch64 in EL1
	mov     x29, #(1 << 31)      // AArch64
	orr     x29, x29, #(1 << 1)   // SWIO hardwired on Pi3
	msr     hcr_el2, x29
	mrs     x29, hcr_el2

	//insert default vectors for core 0
	ldr x29, =vectors
	msr vbar_el1, x29

	// change execution level to EL1
	mov     x29, #0x3c4
	msr     spsr_el2, x29
	adr     x29, 1f
	msr     elr_el2, x29
	eret

1:  mov     sp, x1

	//tell hardware that EL0 has own stack
	msr SPSel, #1

	// jump to go code
	bl      mainFromBootloader
	// for failsafe, halt this core too
	wfe
