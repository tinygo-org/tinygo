#define REGSIZE 4
#define SREG sw
#define LREG lw
#define SAVE_REGS 32
#define CONTEXT_SIZE (SAVE_REGS * 4)

#define CPU_INT_THRESH (0x600c2000 + 0x194 )

	.equ panic_from_exception, handleException
	.equ panic_from_isr, handleException
	.equ _global_interrupt_handler, handleInterrupt

.macro save_regs
	addi sp, sp, -CONTEXT_SIZE
	sw   ra,  0*REGSIZE(sp)
	sw   tp,  1*REGSIZE(sp)
	sw   t0,  2*REGSIZE(sp)
	sw   t1,  3*REGSIZE(sp)
	sw   t2,  4*REGSIZE(sp)
	sw   s0,  5*REGSIZE(sp)
	sw   s1,  6*REGSIZE(sp)
	sw   a0,  7*REGSIZE(sp)
	sw   a1,  8*REGSIZE(sp)
	sw   a2,  9*REGSIZE(sp)
	sw   a3,  10*REGSIZE(sp)
	sw   a4,  11*REGSIZE(sp)
	sw   a5,  12*REGSIZE(sp)
	sw   a6,  13*REGSIZE(sp)
	sw   a7,  14*REGSIZE(sp)
	sw   s2,  15*REGSIZE(sp)
	sw   s3,  16*REGSIZE(sp)
	sw   s4,  17*REGSIZE(sp)
	sw   s5,  18*REGSIZE(sp)
	sw   s6,  19*REGSIZE(sp)
	sw   s7,  20*REGSIZE(sp)
	sw   s8,  21*REGSIZE(sp)
	sw   s9,  22*REGSIZE(sp)
	sw   s10, 23*REGSIZE(sp)
	sw   s11, 24*REGSIZE(sp)
	sw   t3,  25*REGSIZE(sp)
	sw   t4,  26*REGSIZE(sp)
	sw   t5,  27*REGSIZE(sp)
	sw   t6,  28*REGSIZE(sp)
.endm

.macro save_mepc
	csrr t0, mepc
	sw   t0, 29*REGSIZE(sp)
.endm

.macro restore_regs
	lw   ra,  0*REGSIZE(sp)
	lw   tp,  1*REGSIZE(sp)
	lw   t0,  2*REGSIZE(sp)
	lw   t1,  3*REGSIZE(sp)
	lw   t2,  4*REGSIZE(sp)
	lw   s0,  5*REGSIZE(sp)
	lw   s1,  6*REGSIZE(sp)
	lw   a0,  7*REGSIZE(sp)
	lw   a1,  8*REGSIZE(sp)
	lw   a2,  9*REGSIZE(sp)
	lw   a3,  10*REGSIZE(sp)
	lw   a4,  11*REGSIZE(sp)
	lw   a5,  12*REGSIZE(sp)
	lw   a6,  13*REGSIZE(sp)
	lw   a7,  14*REGSIZE(sp)
	lw   s2,  15*REGSIZE(sp)
	lw   s3,  16*REGSIZE(sp)
	lw   s4,  17*REGSIZE(sp)
	lw   s5,  18*REGSIZE(sp)
	lw   s6,  19*REGSIZE(sp)
	lw   s7,  20*REGSIZE(sp)
	lw   s8,  21*REGSIZE(sp)
	lw   s9,  22*REGSIZE(sp)
	lw   s10, 23*REGSIZE(sp)
	lw   s11, 24*REGSIZE(sp)
	lw   t3,  25*REGSIZE(sp)
	lw   t4,  26*REGSIZE(sp)
	lw   t5,  27*REGSIZE(sp)
	lw   t6,  28*REGSIZE(sp)
	addi sp, sp, CONTEXT_SIZE
.endm

.macro restore_mepc
	lw	t0, 29*REGSIZE(sp)
	csrw	mepc, t0
.endm

	/* This is the vector table. MTVEC points here.
	 *
	 * Use 4-byte intructions here. 1 instruction = 1 entry of the table.
	 * The CPU jumps to MTVEC (i.e. the first entry) in case of an exception,
	 * and (MTVEC & 0xfffffffc) + (mcause & 0x7fffffff) * 4, in case of an interrupt.
	 *
	 * Note: for our CPU, we need to place this on a 256-byte boundary, as CPU
	 * only uses the 24 MSBs of the MTVEC, i.e. (MTVEC & 0xffffff00).
	 */

.section ._vector_table.text

	.global _vector_table
	.type _vector_table,@function
_vector_table:
	.option push
	.option norvc

	j _panic_handler			/* exception handler, entry 0 */
	.rept 24
	j _interrupt_handler		/* 24 identical entries, all pointing to the interrupt handler */
	.endr
	j _panic_handler			/* Call panic handler for ETS_T1_WDT_INUM interrupt (soc-level panic)*/
	j _panic_handler			/* Call panic handler for ETS_CACHEERR_INUM interrupt (soc-level panic)*/
	.rept 5
	j _interrupt_handler		/* 6 identical entries, all pointing to the interrupt handler */
	.endr

	.option pop
	.size _vector_table, .-_vector_table


	/* Exception handler.*/
	.type _panic_handler, @function
_panic_handler:
	addi sp, sp, -CONTEXT_SIZE /* allocate space on stack to store necessary registers */
	/* save general registers */
	sw   ra,  0*REGSIZE(sp)
	sw   tp,  1*REGSIZE(sp)
	sw   t0,  2*REGSIZE(sp)
	sw   t1,  3*REGSIZE(sp)
	sw   t2,  4*REGSIZE(sp)
	sw   s0,  5*REGSIZE(sp)
	sw   s1,  6*REGSIZE(sp)
	sw   a0,  7*REGSIZE(sp)
	sw   a1,  8*REGSIZE(sp)
	sw   a2,  9*REGSIZE(sp)
	sw   a3,  10*REGSIZE(sp)
	sw   a4,  11*REGSIZE(sp)
	sw   a5,  12*REGSIZE(sp)
	sw   a6,  13*REGSIZE(sp)
	sw   a7,  14*REGSIZE(sp)
	sw   s2,  15*REGSIZE(sp)
	sw   s3,  16*REGSIZE(sp)
	sw   s4,  17*REGSIZE(sp)
	sw   s5,  18*REGSIZE(sp)
	sw   s6,  19*REGSIZE(sp)
	sw   s7,  20*REGSIZE(sp)
	sw   s8,  21*REGSIZE(sp)
	sw   s9,  22*REGSIZE(sp)
	sw   s10, 23*REGSIZE(sp)
	sw   s11, 24*REGSIZE(sp)
	sw   t3,  25*REGSIZE(sp)
	sw   t4,  26*REGSIZE(sp)
	sw   t5,  27*REGSIZE(sp)
	sw   t6,  28*REGSIZE(sp)
	addi t0, sp, CONTEXT_SIZE /* restore sp with the value when trap happened */
	sw t0,  29*REGSIZE(sp)
	csrr t0, mepc
	sw t0,  30*REGSIZE(sp)
	csrr t0, mstatus
	sw t0,  31*REGSIZE(sp)
	csrr t0, mtvec
	sw t0,  32*REGSIZE(sp)
	csrr t0, mtval
	sw t0,  33*REGSIZE(sp)
	csrr t0, mhartid
	sw t0,  34*REGSIZE(sp)

	/* Call panic_from_exception(sp) or panic_from_isr(sp)
	 * depending on whether we have a pseudo excause or not.
	 * If mcause's highest bit is 1, then an interrupt called this routine,
	 * so we have a pseudo excause. Else, it is due to a exception, we don't
	 * have an pseudo excause */
	mv a0, sp
	csrr a1, mcause
	/* Branches instructions don't accept immediates values, so use t1 to
	 * store our comparator */
	li t0, 0x80000000
	bgeu a1, t0, _call_panic_handler
	sw a1,  35*REGSIZE(sp)
	/* exception_from_panic never returns */
	j panic_from_exception

_call_panic_handler:
	/* Remove highest bit from mcause (a1) register and save it in the
	 * structure */
	not t0, t0
	and a1, a1, t0
	sw a1, 35*REGSIZE(sp)
	/* exception_from_isr never returns */
	j panic_from_isr
	.size  _panic_handler, .-_panic_handler

	/* This is the interrupt handler.
	 * It saves the registers on the stack,
	 * prepares for interrupt nesting,
	 * re-enables the interrupts,
	 * then jumps to the C dispatcher in interrupt.c.
	 */
	.global _interrupt_handler
	.type _interrupt_handler, @function
_interrupt_handler:
	/* entry */
	save_regs
	save_mepc

	/* Before doing anythig preserve the stack pointer */
	/* It will be saved in current TCB, if needed */
	/*
	mv a0, sp
	call rtos_int_enter
	*/

	/* Before dispatch c handler, restore interrupt to enable nested intr */
	csrr s1, mcause
	csrr s2, mstatus

	/* Save the interrupt threshold level */
	li t0, CPU_INT_THRESH
	lw s3, 0(t0)

	/* Increase interrupt threshold level */
	li t2, 0x7fffffff
	and t1, s1, t2		/* t1 = mcause & mask */
	slli t1, t1, 2 		/* t1 = mcause * 4 */
	li t2, 0x600c2000   /* INTERRUPT_CORE0_CPU_INT_PRI_0_REG */
	add t1, t2, t1		/* t1 = INTC_INT_PRIO_REG + 4 * mcause */
	lw t2, 0(t1)		/* t2 = INTC_INT_PRIO_REG[mcause] */
	addi t2, t2, 1		/* t2 = t2 +1 */
	sw t2, 0(t0)		/* CPU_INT_THRESH = t2 */
	fence

	li t0, 0x8
	csrrs t0, mstatus, t0

	#ifdef CONFIG_PM_TRACE
	li	  a0, 0	   /* = ESP_PM_TRACE_IDLE */
	#if SOC_CPU_CORES_NUM == 1
	li	  a1, 0	   /* No need to check core ID on single core hardware */
	#else
	csrr	a1, mhartid
	#endif
	la	  t0, esp_pm_trace_exit
	jalr	t0		  /* absolute jump, avoid the 1 MiB range constraint */
	#endif

	#ifdef CONFIG_PM_ENABLE
	la	  t0, esp_pm_impl_isr_hook
	jalr	t0		  /* absolute jump, avoid the 1 MiB range constraint */
	#endif

	/* call the C dispatcher */
	mv	  a0, sp		/* argument 1, stack pointer */
	csrr	a1, mcause	/* argument 2, interrupt number */
	/* mask off the interrupt flag of mcause */
	li		t0, 0x7fffffff
	and	 a1, a1, t0
	jal	 _global_interrupt_handler

	/* After dispatch c handler, disable interrupt to make freertos make context switch */

	li t0, 0x8
	csrrc t0, mstatus, t0

	/* restore the interrupt threshold level */
	li t0, CPU_INT_THRESH
	sw s3, 0(t0)
	fence

	/* Yield to the next task is needed: */
	/* 
	mv a0, sp
	call rtos_int_exit
	*/

	/* The next (or current) stack pointer is returned in a0 */
	mv sp, a0

	/* restore the rest of the registers */
	csrw mcause, s1
	csrw mstatus, s2
	restore_mepc
	restore_regs

	/* exit, this will also re-enable the interrupts */
	mret
	.size  _interrupt_handler, .-_interrupt_handler
