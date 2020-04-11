.section ".text.boot","",@progbits

/*
 * Interrupt vectors that are jumped to by the hardware when we get
 * an exception. There are 4 states and 4 types of exceptions so
 * we have 0x10 entries, and each entry cannot exceed 0x80
 */

#define S_FRAME_SIZE			256 		// size of all saved registers

#define SYNC_EL1t		0
#define IRQ_EL1t		1
#define FIQ_EL1t		2
#define ERROR_EL1t		3

#define SYNC_EL1h		4
#define IRQ_EL1h		5
#define FIQ_EL1h		6
#define ERROR_EL1h		7

#define SYNC_EL0_64	    	8
#define IRQ_EL0_64	    	9
#define FIQ_EL0_64		10
#define ERROR_EL0_64		11

#define SYNC_EL0_32		12
#define IRQ_EL0_32		13
#define FIQ_EL0_32		14
#define ERROR_EL0_32		15

	.macro interrupt_entry type
	kernel_entry
	mov	x0, #\type
	mrs	x1, esr_el1
	mrs	x2, elr_el1

    mov x3,#\type
    lsl x3,x3,#3 //multiply by 8
    ldr x4,excptrs
    add x4,x4,x3
    ldr x5,[x4]

    br x5
	b	err_hang
	.endm

	.macro	ventry	label
	.align	7
	b	\label
	.endm

	.macro	kernel_entry
	sub	sp, sp, #S_FRAME_SIZE
	stp	x0, x1, [sp, #16 * 0]
	stp	x2, x3, [sp, #16 * 1]
	stp	x4, x5, [sp, #16 * 2]
	stp	x6, x7, [sp, #16 * 3]
	stp	x8, x9, [sp, #16 * 4]
	stp	x10, x11, [sp, #16 * 5]
	stp	x12, x13, [sp, #16 * 6]
	stp	x14, x15, [sp, #16 * 7]
	stp	x16, x17, [sp, #16 * 8]
	stp	x18, x19, [sp, #16 * 9]
	stp	x20, x21, [sp, #16 * 10]
	stp	x22, x23, [sp, #16 * 11]
	stp	x24, x25, [sp, #16 * 12]
	stp	x26, x27, [sp, #16 * 13]
	stp	x28, x29, [sp, #16 * 14]
	str	x30, [sp, #16 * 15]
	.endm

	.macro	kernel_exit
	ldp	x0, x1, [sp, #16 * 0]
	ldp	x2, x3, [sp, #16 * 1]
	ldp	x4, x5, [sp, #16 * 2]
	ldp	x6, x7, [sp, #16 * 3]
	ldp	x8, x9, [sp, #16 * 4]
	ldp	x10, x11, [sp, #16 * 5]
	ldp	x12, x13, [sp, #16 * 6]
	ldp	x14, x15, [sp, #16 * 7]
	ldp	x16, x17, [sp, #16 * 8]
	ldp	x18, x19, [sp, #16 * 9]
	ldp	x20, x21, [sp, #16 * 10]
	ldp	x22, x23, [sp, #16 * 11]
	ldp	x24, x25, [sp, #16 * 12]
	ldp	x26, x27, [sp, #16 * 13]
	ldp	x28, x29, [sp, #16 * 14]
	ldr	x30, [sp, #16 * 15]
	add	sp, sp, #S_FRAME_SIZE
	eret
	.endm

.align 7
//vectors:
	ventry	sync_el1t			// Synchronous EL1t
	ventry	irq_el1t			// IRQ EL1t
	ventry	fiq_el1t			// FIQ EL1t
	ventry	error_el1t			// Error EL1t

	ventry	sync_el1h			// Synchronous EL1h
	ventry	irq_el1h		    // IRQ EL1h -- big dog
	ventry	fiq_el1h			// FIQ EL1h
	ventry	error_el1h			// Error EL1h

	ventry	sync_el0_64			// Synchronous 64-bit EL0
	ventry	irq_el0_64			// IRQ 64-bit EL0
	ventry	fiq_el0_64			// FIQ 64-bit EL0
	ventry	error_el0_64		// Error 64-bit EL0

	ventry	sync_el0_32			// Synchronous 32-bit EL0
	ventry	irq_el0_32			// IRQ 32-bit EL0
	ventry	fiq_el0_32			// FIQ 32-bit EL0
	ventry	error_el0_32		// Error 32-bit EL0


// these functions need to be broken out so each macro instation can jump to a different point.
.globl sync_el1t
sync_el1t:
	interrupt_entry  SYNC_EL1t

.globl irq_el1t
irq_el1t:
	interrupt_entry  IRQ_EL1t

.globl fiq_el1t
fiq_el1t:
	interrupt_entry  FIQ_EL1t

.globl error_el1t
error_el1t:
	interrupt_entry  ERROR_EL1t

.globl sync_el1h
sync_el1h:
	interrupt_entry  SYNC_EL1h

.globl irq_el1h
irq_el1h:
    interrupt_entry IRQ_EL1h

.globl fiq_el1h
fiq_el1h:
	interrupt_entry  FIQ_EL1h

.globl error_el1h
error_el1h:
	interrupt_entry  ERROR_EL1h

.globl sync_el0_64
sync_el0_64:
	interrupt_entry  SYNC_EL0_64

.globl irq_el0_64
irq_el0_64:
	interrupt_entry  IRQ_EL0_64

.globl fiq_el0_64
fiq_el0_64:
	interrupt_entry  FIQ_EL0_64

.globl error_el0_64
error_el0_64:
	interrupt_entry  ERROR_EL0_64

.globl sync_el0_32
sync_el0_32:
	interrupt_entry  SYNC_EL0_32

.globl irq_el0_32
irq_el0_32:
	interrupt_entry  IRQ_EL0_32

.globl fiq_el0_32
fiq_el0_32:
	interrupt_entry  FIQ_EL0_32

.globl error_el0_32
error_el0_32:
	interrupt_entry  ERROR_EL0_32

.globl err_hang
err_hang: b err_hang

.globl excptrs
excptrs:
.space 0x8*16



.section ".text"
.globl _start
_start:
    mov x0, #0x760
    lsl x0,x0,#8
    mov sp,x0 //safety if we fail during boot
	mrs	x0, mpidr_el1		
	and	x0, x0,#0xFF		// Check processor id
	cbz	x0, master		// Hang for all non-primary CPU
	b	proc_hang

.align 3
proc_hang: 
	b 	proc_hang

master:
    bl get_el
    sub x4,x0,#2 //store 0 if EL2 and 1 (assumed) if EL3 (hardware?)
	//cant do this
	//orr x0, xzr, #0x30D00800 // SystemControlRegisterValueMMUDisabled
	//so forced to do this
    orr x0,xzr, #3
    lsl x0,x0,#28
    orr x1,xzr, #3
    lsl x1,x1,22
    orr x2,xzr,#1
    lsl x2,x2,20
    orr x2,x2,#0x800
    orr x1,x1,x2
    orr x0,x0,x1
    msr	sctlr_el1, x0

	orr	x0,xzr, #0x80000000 // HypervisorConfigurationRegisterValue
	msr	hcr_el2, x0

    //do we need to do the drop from el2 because we are in qemu?
    cbz x4,el2_drop_to_el1

    //we are in el3
	orr	x0,xzr, #0x400 // SecureConfigurationRegisterValue
	orr x0,x0, #0x30
	orr x0,x0,#1
	msr	scr_el3, x0

    //xxx seems like this should be C3 not C0? to go to EL1
	and x0,x0, #0x1C0 // SavedProgramStatusRegisterValue, we use a const for el3
	msr	spsr_el3, x0

	adr	x0, el1_entry		
	msr	elr_el3, x0
	eret

el2_drop_to_el1:
	mrs x0, spsr_el2 // SavedProgramStatusRegisterValue but for EL2
	orr x0,x0,#0x4 //bit 2, bit 0  (bit 2 => level 1 is desired EL, bit0 means use own stack for EL1), so EL1h
	orr x0,x0,#0x1 //second step because #0x5 isn't a legal const
	msr	spsr_el2, x0

	adr	x0, el1_entry
	msr	elr_el2, x0
    eret

el1_entry:
    // retset stack before our code (we are now in el1)
    mov     x1, #0x760
    lsl     x1,x1,#0x8
    mov     sp, x1
    mov     x30,xzr

    //can't branch and link here because the stack and link are in a bad state
	b main
	b err_hang		// nuke the site from orbit...

.globl get_el
get_el:
	mrs x0, CurrentEL
	lsr x0, x0, #2
	ret

