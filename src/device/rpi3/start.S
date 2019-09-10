/*
 * This is partially derived from excellent RPI3 tutorial
 * https://github.com/bztsrc/raspi3-tutorial/tree/master/05_uart0
 * and partially from the super os examples of
 * https://github.com/s-matyukevich/raspberry-pi-os
 */

 .section ".text.boot"

 .global _start

 _start:
     // read cpu id, stop slave cores
     mrs     x1, mpidr_el1
     and     x1, x1, #3
     cbz     x1, 2f
     // cpu id > 0, stop
 1:  wfe
     b       1b
 2:  // cpu id == 0

     // set stack before our code
		 bl setupStack

     // set up EL1
     mrs     x0, CurrentEL
     and     x0, x0, #12 // clear reserved bits

     // running at EL3?
     cmp     x0, #12
     bne     5f
     // should never be executed, just for completeness
     mov     x2, #0x5b1
     msr     scr_el3, x2
     mov     x2, #0x3c9
     msr     spsr_el3, x2
     adr     x2, 5f
     msr     elr_el3, x2
     eret

     // running at EL2?
 5:  cmp     x0, #4
     beq     6f

 6:
     // clear bss
     ldr     x1, =__bss_start
     ldr     w2, =__bss_size
 3:  cbz     w2, 4f
     str     xzr, [x1], #8
     sub     w2, w2, #1
     cbnz    w2, 3b

		 // jump to GO code, should not return
 4:
 		mov sp, x1
 		bl      main
		// for failsafe, halt this core too
  	wfe

.weak setupStack
.set setupStack, stackNoBL
stackNoBL:
	mov     x1, #0x90000
	ret
