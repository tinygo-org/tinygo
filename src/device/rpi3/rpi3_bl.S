.section ".text.boot_bl"

// startup protocol is:
// x0 - start address (also in PC)
// x1 - stack ptr
// x2 - heap ptr
// x3 - unix time
.global _start
	mov sp, x1
	// jump to go code
	bl      mainFromBootloader
	// for failsafe, halt this core too
	wfe
