.section ".text.boot_bl"

//
// The bootloader changes execution state for bootloaded code to EL2. This
// drops to EL1 once the stack is setup ok.
//

// startup parameters are:
// x0 - start address (also in PC)
// x1 - stack ptr
// x2 - heap ptr
// x3 - unix time
.global _start
	//set the sp for el1, but we have to have the fp space so subtract 16 (two 64bit ptr)
	sub x1,x1,#16
	msr     sp_el1, x1

	//https://static.docs.arm.com/100878/0100/fundamentals_of_armv8_a_100878_0100_en.pdf
	//msr     SPSel, #1


	// enable CNTP for EL1
	//mrs     x27, cnthctl_el2
	//orr     x27, x27, #3
	//msr     cnthctl_el2, x27
	//msr     cntvoff_el2, xzr
	// enable AArch64 in EL1
	mov     x27, #(1 << 31)      // AArch64
	orr     x27, x27, #(1 << 1)   // SWIO hardwired on Pi3
	msr     hcr_el2, x27
	mrs     x27, hcr_el2

	mov     x4, #0x0800
	movk    x4, #0x30d0, lsl #16
	msr     sctlr_el1, x4

	mov x4,#0x02
	bl debugint

	//insert default vectors for core 0
//	ldr x27, =vectors
//	msr vbar_el1, x27

	mov x4,#0x04
	bl debugint

	// change execution level to EL1
	mov     x27, #0x3c5  //5! not 4! we have our own stack
	msr     spsr_el2, x27
	adr     x27, 1f
	msr     elr_el2, x27
	mov x4,x27
	bl debugint
	eret

## getting the stack and fp in order, we subtracted 16 from x1 above
## note that we store 0 for the link and ret values because we are at the
## bottom of the stack and want stack to correctly show in GDB and similar
1:
	mov x4,sp
	bl debugint

	# fp to bottom of frame
	add x27, x1, #16
	mov x29, x27

	mov x4,x29
	bl debugint

	#shove values in based on top of stack
	str xzr, [sp, #8]
	str xzr, [sp, #16]

	mov x4,#6
	bl debugint

	//mov sp, x1
	//sp points to top of stack (for BL code 0x30000 - 0x18)
	//fp points to bottom (x29) (for BL code 0x30000 - 0x8)

	//tell hardware that EL0 has own stack
	msr SPSel, #1

	mov x4,#0x07
	bl debugint


	// jump to go code
	bl      mainFromBootloader
	// for failsafe, halt this core too
	wfe
