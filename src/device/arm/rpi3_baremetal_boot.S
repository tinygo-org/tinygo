//"     Modified by iansmith May 2020                                       "
//
//"+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
//"                                                                         "
//"      Filename: smartstart64.s                                           "
//"      Copyright(c): Leon de Boer(LdB) 2017                               "
//"      Version: 2.03                                                      "
//"                                                                         "
//"***************[ THIS CODE IS FREEWARE UNDER CC Attribution]*************"
//"                                                                         "
//"    This sourcecode is released for the purpose to promote programming   "
//"  on the Raspberry Pi. You may redistribute it and/or modify with the    "
//"  following disclaimer and condition.                                    "
//"                                                                         "
//"      The SOURCE CODE is distributed "AS IS" WITHOUT WARRANTIES AS TO    "
//"   PERFORMANCE OF MERCHANTABILITY WHETHER EXPRESSED OR IMPLIED.          "
//"   Redistributions of source code must retain the copyright notices to   "
//"   maintain the author credit (attribution) .                            "
//"                                                                         "
//"*************************************************************************"
//"                                                                         "
//"      This code expands on my earlier SmartStart bootstub assembler for  "
//" the Pi3. It directly supports multicore operation in C/C++. To do that  "
//" it provides stack space to each core and provides modified bootloader   "
//" spinlock that protects against registers X0-X7 trashed. As any C/C++    "
//" 64 bit compiler will trash those registers, to use C/C++ in multicore   "
//" programming this protection must be done.                               "
//"       This is a matching paired AARCH64 stub for the 64bit linker file  "
//" and carries the samme arrangement as AARCH32 to ease porting.           "
//"                                                                         "
//"+++++++++++++++++++++++[ REVISIONS ]+++++++++++++++++++++++++++++++++++++"
//"  1.01 Initial release .. Pi autodetection main aim                      "
//"  1.02 Many functions moved out C to aide 32/64 bit compatability        "
//"  2.01 Futher reductions to bare minmum assembeler code                  "
//"  2.02 Multicore functionality added                                     "
//"  2.03 Timer Irq support added                                           "
//"  2.04 David Welch routines GET32/PUT32 compability added                "
//"  2.05 Irq stub made weak linked so it can be easily overridden          "
//"+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"

.section ".init", "ax", %progbits
.balign	4
.globl _start

//"================================================================"
//  This simply hold startup data for SmartStart its trivial
//"================================================================"
_start:
	adr x12, _start                         // Hold boot address in high register R12
	ldr x1, =RPi_BootAddr                   // Address of RPi_BootAddr
	str w12, [x1]                           // Store the boot address

	ldr x0, =0x3F000000                     // No need to detect address in 64bit mode it has to be 0x3F000000 on Pi3
	ldr x1, =RPi_IO_Base_Addr               // Address of RPi_IO_Base_Addr
	str w0, [x1]                            // Store the IO base address

	ldr x0, =0xC0000000                     // No need to detect address in 64bit mode it has to be 0xC0000000 on Pi3
	ldr x1, =RPi_ARM_TO_GPU_Alias           // Address of RPi_IO_Base_Addr
	str w0, [x1]                            // Store the IO base address

	mov x0, #1                              // Multicore support starts as 1 core
	ldr x1, =RPi_CoresReady                 // Address of RPi_CoresReady
	str w0, [x1]                            // Store the CoresReady count as 1

	mov x0, #0x98                           // Compiled for ARM8 CPU in AARCH64 and supports 4 cores
	ldr x1, =RPi_CompileMode                // Address of RPi_CompileMode
	str w0, [x1]                            // Store the compiler mode

	mrs x0, midr_el1                        // Read CPU Id register
	ldr x1, =RPi_CpuId                      // Address of RPi_CpuId
	str w0, [x1]                            // Store the CPU id


//"================================================================"
//  Initialize Generic Timers for Core0
//"================================================================"
	mrs	x0, cnthctl_el2
	orr	x0, x0, #0x3 						/* Enable EL1 access to timers */
	msr	cnthctl_el2, x0
	msr	cntvoff_el2, xzr

	/* Not sure if I want this yet */
	mrs	x0, cntkctl_el1
	orr	x0, x0, #0x3						/* Enable EL0 access to timers */
	msr	cntkctl_el1, x0

//"================================================================"
//  Setup stack pointers for each core and each CPU operation mode
//"================================================================"
multicore_start:
    ldr x2, = __EL2_stack_core0             // Address of EL2_stack_core0 stack pointer value
    ldr x3, = __EL1_stack_core0             // Address of EL1_stack_core0 stack pointer value
    ldr x4, = __EL0_stack_core0             // Address of EL0_stack_core0 stack pointer value
    mrs x6, mpidr_el1                       // Read core id on ARM8
    ands x6, x6, #0x3                       // Make cpu id bitmask
    beq set_stacks                          // Ready to set core 0 stacks
    ldr x2, = __EL2_stack_core1             // Address of EL2_stack_core1 stack pointer value
    ldr x3, = __EL1_stack_core1             // Address of EL1_stack_core1 stack pointer value
    ldr x4, = __EL0_stack_core1             // Address of EL0_stack_core1 stack pointer value
    cmp x6, #1                              // Check cpu id for core 1
    beq set_stacks                          // Ready to set core 1 stacks
    ldr x2, = __EL2_stack_core2             // Address of EL2_stack_core2 stack pointer value
    ldr x3, = __EL1_stack_core2             // Address of EL1_stack_core2 stack pointer value
    ldr x4, = __EL0_stack_core2             // Address of EL0_stack_core2 stack pointer value
    cmp x6, #2                              // Check cpu id for core 2
    beq set_stacks                          // Ready to set core 1 stacks
    ldr x2, = __EL2_stack_core3             // Address of EL2_stack_core3 stack pointer value
    ldr x3, = __EL1_stack_core3             // Address of EL1_stack_core3 stack pointer value
    ldr x4, = __EL0_stack_core3             // Address of EL0_stack_core3 stack pointer value
set_stacks:
	mov	sp, x2                              /* EL2 stack set */
	msr	sp_el1, x3                          /* EL1 stack set */
	msr	sp_el0, x4                          /* EL0 stack set */

//"================================================================"
//  Initilize MPID/MPIDR registers for all Cores
//"================================================================"
	mrs	x0, midr_el1
	mrs	x1, mpidr_el1
	msr	vpidr_el2, x0
	msr	vmpidr_el2, x1

//"================================================================"
//  Disable coprocessor traps for all Cores
//"================================================================"
	mov	x0, #0x33ff
	msr	cptr_el2, x0						// Disable coprocessor traps to EL2
	msr	hstr_el2, xzr						// Disable coprocessor traps to EL2
	mov	x0, #3 << 20
	msr	cpacr_el1, x0						// Enable FP/SIMD at EL1

//"================================================================"
//  Initialize HCR_EL2 so EL1 is 64 bits for all Cores
//"================================================================"
	mov	x0, #(1 << 31)						// 64bit EL1
	msr	hcr_el2, x0

//"================================================================"
//  Initialize SCTLR_EL1 for all Cores
//"================================================================"
    /*  RES1 bits (29,28,23,22,20,11) to 1
	 *  RES0 bits (31,30,27,21,17,13,10,6) +
	 *  UCI,EE,EOE,WXN,nTWE,nTWI,UCT,DZE,I,UMA,SED,ITD,
	 *  CP15BEN,SA0,SA,C,A,M to 0 */
	mov	x0, #0x0800
	movk	x0, #0x30d0, lsl #16
	orr    x0, x0, #(0x1 << 2)            // The C bit on (data cache).
	orr    x0, x0, #(0x1 << 12)           // The I bit on (instruction cache)
	msr	sctlr_el1, x0
//"================================================================"
//  Setup separate vector table for EL2 on all cores
//"================================================================"
    ldr x0, =VectorTableEL2
    msr vbar_el2,x0
//"================================================================"
//  Return to the EL1_SP1 mode from EL2 for all Cores
//"================================================================"
    mov x0, #0x3c5                          // EL1_SP1 | D | A | I | F
    msr spsr_el2, x0                        // Set spsr_el2 with settings
    adr x0, exit_el1                        // Address to exit EL2
    msr elr_el2, x0                         // Set elevated return register
    eret                                    // Call elevated return
exit_el1:
    mrs x6, mpidr_el1                       // Read core id on ARM8
    and x6, x6, #0x3                        // Make cpu id bitmask
    cbz x6, cpu0_exit_multicore_park        // Core0 continues on

//"================================================================"
//      Now park Core 1,2,3 into secondary spinloop on BCM2837
//"================================================================"
    ldr x1, =RPi_CoresReady                 // Address of CoreReady count
    ldr w0, [x1]                            // Load current core count
    add w0, w0, #1                          // Add one as core about to go ready
    str w0, [x1]                            // Update CoreReady count
    b StartSecondarySpin                    // Jump to setup secondary spin
cpu0_exit_multicore_park:

//"================================================================"
//  Set vector table for EL1 for Core0 (All others cores parked)
//"================================================================"
    ldr x0, =VectorTable
    msr vbar_el1,x0

//"================================================================"
//  About to go to into go kernel clear BSS (Core0 only)
//"================================================================"
	ldr x3, =_ebss //iansmith modification
	ldr x0, =_sbss //iansmith modification
	cmp	x0, x3
	bcs	.bss_cleared
.bss_zero_loop:
	str	wzr, [x0], 4
	cmp	x3, x0
	bhi	.bss_zero_loop
.bss_cleared:

//"================================================================"
//			Core0 will bring Core 1,2,3 to secondary spin
//"================================================================"
.equ spin_cpu1, 0xe0
    mov x1, #spin_cpu1                      // Spin core1 jump address
    ldr x2, =multicore_start                // Function we are going to call
    str x2, [x1]                            // Store the function address to core1
    sev                                     // Wake core1 up
    ldr x3, =RPi_CoresReady                 // Set CoresReady count address
.WaitCore1ACK:
    ldr w1, [x3]                            // Read CoresReady count
    cmp w1, #2                              // Wait for setting of second core ready
    bne .WaitCore1ACK                       // Core1 not ready so read again
.equ spin_cpu2, 0xe8
    mov x1, #spin_cpu2                      // Spin core2 jump address
    ldr x2, =multicore_start                // Function we are going to call
    str x2, [x1]                            // Store the function address to core2
    sev                                     // Wake core2 up
    ldr x3, =RPi_CoresReady                 // Set CoresReady count address
.WaitCore2ACK:
    ldr w1, [x3]                            // Read CoresReady count
    cmp w1, #3                              // Wait for setting of third core ready
    bne .WaitCore2ACK                       // Core2 not ready so read again
.equ spin_cpu3, 0xf0
    mov x1, #spin_cpu3                      // Spin core3 jump address
    ldr x2, =multicore_start                // Function we are going to call
    str x2, [x1]                            // Store the function address to core3
    sev                                     // Wake core3 up
    ldr x3, =RPi_CoresReady                 // Set CoresReady count address
.WaitCore3ACK:
    ldr w1, [x3]                            // Read CoresReady count
    cmp w1, #4                              // Wait for setting of third core ready
    bne .WaitCore3ACK                       // Core3 not ready so read again

//"================================================================"
// Finally that all done Core0 jumps to the go compiler entry point
//"================================================================"
    b main                                  // Jump out to go kernel

/*================================================================"
      Just safety in case Go code ever returns back
"================================================================*/
hang:
    b hang

/*++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++}
{    Modified bootloader Spin loop but tolerant on registers X0-X7 for C    }
{++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/
.equ spin_cpu0, 0xd8
.balign	4
StartSecondarySpin:
    mrs x6, MPIDR_EL1                       // Fetch core Id
    and x6, x6, #0x3                        // Create 2 bit mask of core Id
    mov x5, #spin_cpu0                      // Load address of spins
    mov x1, #0                              // zero x1 register
    str x1, [x5, x6, lsl #3]                // Make sure caller address is zeroed
secondary_spin:
    wfe                                     // This is all I really want: processor to sleep
    ldr x4, [x5, x6, lsl #3]                // Fetch address that has been set
    cbz x4, secondary_spin                  // If zero spin
    mov x0, #0
    str x0, [x5, x6, lsl #3]                // Zero caller address
    mov x1, #0
    mov x2, #0
    mov x3, #0
    blr x4                                  // Call the function set
    b StartSecondarySpin                    // Loop back to spinlock

.balign 4
.ltorg                                      // Tell assembler ltorg data for code above can go here

/*++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++}
{			GLOBAL INTERRUPT CONTROL PROVIDE BY RPi-SmartStart API          }
{++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

/* "PROVIDE C FUNCTION: void EnableInterrupts (void);" */
.section .text.EnableInterrupts, "ax", %progbits
.balign 4
.globl EnableInterrupts
.type EnableInterrupts, %function
EnableInterrupts:
	msr daifclr,#3
	ret
.balign 4
.ltorg                                      // Tell assembler ltorg data for this code can go here
.size	EnableInterrupts, .-EnableInterrupts

/* "PROVIDE C FUNCTION: void DisableInterrupts (void);" */
.section .text.DisableInterrupts, "ax", %progbits
.balign 4
.globl DisableInterrupts
.type DisableInterrupts, %function
DisableInterrupts:
	msr daifset,#3
	ret
.balign	4
.ltorg                                    // Tell assembler ltorg data for this code can go here
.size	DisableInterrupts, .-DisableInterrupts

/*++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++}
{	   	RPi-SmartStart API TO SET CORE EXECUTE ROUTINE AT ADDRESS 		    }
{++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

/* "PROVIDE C FUNCTION: bool CoreExecute (uint8_t coreNum, void (*func) (void));" */
.section .text.CoreExecute, "ax", %progbits
.balign 4
.globl CoreExecute
.type CoreExecute, %function
CoreExecute:
    ands x0, x0, #255
    beq CoreExecuteFail
    ldr x3, =RPi_CoresReady
    ldr w2, [x3]                            // Fetch cores ready count
    cmp w0, w2
    bcs CoreExecuteFail
    mov x6, #0
    mov w6, w0
    mov x5, #spin_cpu0                      // Load address of spins
    str x1, [x5, x6, lsl #3]                // Save caller address
    dsb sy
    sev
    mov x0, #1
    ret
CoreExecuteFail:
    mov x0, #0
    ret
.balign 4
.ltorg                                     // Tell assembler ltorg data for this code can go here
.size CoreExecute, .-CoreExecute

/*++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++}
{	  RPi-SmartStart Compatability for David Welch CALLS he always uses     }
{++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

/* "PROVIDE C FUNCTION: void PUT32 (uint32_t addr, uint32_t value);" */
.section .text.PUT32, "ax", %progbits
.balign 4
.globl PUT32
.type PUT32, %function
PUT32:
  str w1,[x0]
  ret
.balign 4
.ltorg                                      // Tell assembler ltorg data for this code can go here
.size PUT32, .-PUT32

 /* "PROVIDE C FUNCTION: uint32_t GET32 (uint32_t addr);" */
 .section .text.GET32, "ax", %progbits
.balign	4
.globl GET32
.type GET32, %function
GET32:
    ldr w0,[x0]
    ret
.balign 4
.ltorg                                     // Tell assembler ltorg data for this code can go here
.size GET32, .-GET32

/*++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++}
{           VC4 ADDRESS HELPER ROUTINES PROVIDE BY RPi-SmartStart API       }
{++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

/* "PROVIDE C FUNCTION: uint32_t ARMaddrToGPUaddr (void* ARMaddress);" */
.section .text.ARMaddrToGPUaddr, "ax", %progbits
.balign	4
.globl ARMaddrToGPUaddr;
.type ARMaddrToGPUaddr, %function
//"================================================================"
// ARMaddrToGPUaddr -- AARCH64 Pi3 code
// C Function: uint32_t ARMaddrToGPUaddr (void* ARMaddress);
// Entry: x0 will have ARMAddress value
//"================================================================"
ARMaddrToGPUaddr:
    mov x1, #0xC0000000                     // ARM to VC conversion value
    orr x0, x0, x1                          // Create bus address
    ret                                     // Return
.balign 4
.ltorg                                      // Tell assembler ltorg data for this code can go here
.size ARMaddrToGPUaddr, .-ARMaddrToGPUaddr


/* "PROVIDE C FUNCTION: uint32_t GPUaddrToARMaddr (uint32_t BUSaddress);" */
.section .text.GPUaddrToARMaddr, "ax", %progbits
.balign	4
.globl GPUaddrToARMaddr;
.type GPUaddrToARMaddr, %function
//"================================================================"
// GPUaddrToARMaddr -- AARCH64 Pi3 code
// C Function: uint32_t GPUaddrToARMaddr (uint32_t BUSaddress);
// Entry: x0 will have GPUAddress value
//"================================================================"
GPUaddrToARMaddr:
    mov x1, #0xC0000000                     // ARM to VC conversion value
    bic x0, x0, x1                          // Create arm address
    ret                                     // Return
.balign 4
.ltorg                                      // Tell assembler ltorg data for this code can go here
.size GPUaddrToARMaddr, .-GPUaddrToARMaddr

.globl semihosting_call                     // support for ARM v2 semihosting (qemu)
semihosting_call:
    hlt 0xf000
    ret

/*++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++}
{				IRQ HELPER ROUTINES PROVIDE BY RPi-SmartStart API		    }
{++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

/* "PROVIDE C FUNCTION: uintptr_t setTimerIrqAddress (void (*ARMaddress)(void));" */
.section .text.setTimerIrqAddress, "ax", %progbits
.balign 4
.globl setTimerIrqAddress;
.type setTimerIrqAddress, %function
setTimerIrqAddress:
    msr daifset,#2                          // Disable irq interrupts as we are clearly changing call
    ldr x1, =RPi_TimerIrqAddr               // Load address of function to call on interrupt
    ldr x2, [x1]                            // Load current irq call address
    str x0, [x1]                            // Store the new function pointer address we were given
    mov x0, x2                              // return the old call function
    ret                                     // Return
.balign 4
.ltorg                                      // Tell assembler ltorg data for this code can go here
.size setTimerIrqAddress, .-setTimerIrqAddress



/* Re-entrant interrupt handler stub */
/* http://infocenter.arm.com/help/index.jsp?topic=/com.arm.doc.den0024a/ch10s05.html */
irq_handler_stub:
    stp	x29, x30, [sp, #-16]!
    stp	x27, x28, [sp, #-16]!
    stp	x25, x26, [sp, #-16]!
    stp	x23, x24, [sp, #-16]!
    stp	x21, x22, [sp, #-16]!
    stp	x19, x20, [sp, #-16]!
    stp	x17, x18, [sp, #-16]!
    stp	x15, x16, [sp, #-16]!
    stp	x13, x14, [sp, #-16]!
    stp	x11, x12, [sp, #-16]!
    stp	x9, x10, [sp, #-16]!
    stp	x7, x8, [sp, #-16]!
    stp	x5, x6, [sp, #-16]!
    stp	x3, x4, [sp, #-16]!
    stp	x1, x2, [sp, #-16]!
    str	x0, [sp, #-16]!

    ldr x0, =RPi_TimerIrqAddr           // Address to TimerIrqAddr
    ldr x0, [x0]                        // Load TimerIrqAddr value
    cbz x0, no_irqset                   // If zero no irq set
    blr x0                              // Call Irqhandler that has been set
no_irqset:
    ldr	x0, [sp], #16
    ldp	x1, x2, [sp], #16
    ldp	x3, x4, [sp], #16
    ldp	x5, x6, [sp], #16
    ldp	x7, x8, [sp], #16
    ldp	x9, x10, [sp], #16
    ldp	x11, x12, [sp], #16
    ldp	x13, x14, [sp], #16
    ldp	x15, x16, [sp], #16
    ldp	x17, x18, [sp], #16
    ldp	x19, x20, [sp], #16
    ldp	x21, x22, [sp], #16
    ldp	x23, x24, [sp], #16
    ldp	x25, x26, [sp], #16
    ldp	x27, x28, [sp], #16
    ldp	x29, x30, [sp], #16
    eret

/* macro to align handlers every 0x80 bytes */
.macro vector handler type
.balign 0x80
mov x0,\type
b \handler
.endm

.balign 0x800
.globl VectorTable
VectorTable:
    // from current EL with sp_el0
    vector ex_el1 #0       // Synchronous
    vector ex_el1 #1       // IRQ
    vector ex_el1 #2       // FIQ
    vector ex_el1 #3       // SError

    // from current EL with sp_elx, x != 0
    vector  ex_el1 #4      // Synchronous
    vector  ex_el1 #5      // IRQ
    vector  ex_el1 #6      // FIQ
    vector  ex_el1 #7      // SError

    // from lower EL, target EL minus 1 is AArch64
    vector  ex_el1 #8       // Synchronous
    vector  ex_el1 #9       // IRQ
    vector  ex_el1 #10      // FIQ
    vector  ex_el1 #11      // SError

    // from lower EL, target EL minus 1 is AArch32
    vector  ex_el1 #12      // Synchronous
    vector  ex_el1 #13      // IRQ
    vector  ex_el1 #14      // FIQ
    vector  ex_el1 #15      // SError

.balign 0x800
.globl VectorTableEL2
VectorTableEL2:
    // from current EL with sp_el0
    vector ex_el1 #0       // Synchronous
    vector ex_el1 #1       // IRQ
    vector ex_el1 #2       // FIQ
    vector ex_el1 #3       // SError

    // from current EL with sp_elx, x != 0
    vector  ex_el1 #4      // Synchronous
    vector  ex_el1 #5      // IRQ
    vector  ex_el1 #6      // FIQ
    vector  ex_el1 #7      // SError

    // from lower EL, target EL minus 1 is AArch64
    vector  ex_el1 #8       // Synchronous
    vector  ex_el1 #9       // IRQ
    vector  ex_el1 #10      // FIQ
    vector  ex_el1 #11      // SError

    // from lower EL, target EL minus 1 is AArch32
    vector  ex_el1 #12      // Synchronous
    vector  ex_el1 #13      // IRQ
    vector  ex_el1 #14      // FIQ
    vector  ex_el1 #15      // SError

/****************************************************************
           DATA FOR SMARTSTART64  EXPOSED TO INTERFACE
****************************************************************/
.section ".data.smartstart64", "aw"
.balign 4

.globl RPi_IO_Base_Addr;                // Make sure Pi_IO_Base_Addr label is global
RPi_IO_Base_Addr : .4byte 0;            // Peripheral Base addr is 4 byte variable in 64bit mode

.globl RPi_ARM_TO_GPU_Alias;            // Make sure RPi_ARM_TO_GPU_Alias label is global
RPi_ARM_TO_GPU_Alias : .4byte 0;        // ARM to GPU alias is 4 byte variable in 32bit mode

.globl RPi_BootAddr;                    // Make sure RPi_BootAddr label is global
RPi_BootAddr : .4byte 0;                // CPU boot address is 4 byte variable in 64bit mode

.globl RPi_CoresReady;                  // Make sure RPi_CoresReady label is global
RPi_CoresReady : .4byte 0;              // CPU cores ready for use is 4 byte variable in 32bit mode

.globl RPi_CPUBootMode;                 // Make sure RPi_CPUBootMode label is global
RPi_CPUBootMode : .4byte 0;             // CPU Boot Mode is 4 byte variable in 64bit mode

.globl RPi_CpuId;                       // Make sure RPi_CpuId label is global
RPi_CpuId : .4byte 0;                   // CPU Id is 4 byte variable in 64bit mode

.globl RPi_CompileMode;                 // Make sure RPi_CompileMode label is global
RPi_CompileMode : .4byte 0;             // Compile mode is 4 byte variable in 64bit mode

.globl RPi_SmartStartVer;               // Make sure RPi_SmartStartVer label is global
RPi_SmartStartVer : .4byte 0x00020007;  // SmartStart version is 4 byte variable in 32bit mode

/****************************************************************
           DATA FOR SMARTSTART64  NOT EXPOSED TO INTERFACE
****************************************************************/
.balign 8
RPi_TimerIrqAddr : .8byte 0;           // Timer Irq Address

.section .text.golang, "ax", %progbits
.balign 8
.extern raw_exception_handler

//.globl ex_el1
ex_el1:
    stp	x29, x30, [sp, #-16]!
    stp	x27, x28, [sp, #-16]!
    stp	x25, x26, [sp, #-16]!
    stp	x23, x24, [sp, #-16]!
    stp	x21, x22, [sp, #-16]!
    stp	x19, x20, [sp, #-16]!
    stp	x17, x18, [sp, #-16]!
    stp	x15, x16, [sp, #-16]!
    stp	x13, x14, [sp, #-16]!
    stp	x11, x12, [sp, #-16]!
    stp	x9, x10, [sp, #-16]!
    stp	x7, x8, [sp, #-16]!
    stp	x5, x6, [sp, #-16]!
    stp	x3, x4, [sp, #-16]!
    stp	x1, x2, [sp, #-16]!
    str	x0, [sp, #-16]!

    ldr x5, =raw_exception_handler      // Address to raw exception handler
    mrs x1, esr_el1                     // get syndrome register
    mrs x2, elr_el1                     // get link register
    mrs x3, CurrentEL
    lsr x3,x3,#2
    mrs x4, mpidr_el1                   // Fetch core Id
    and x4, x4, #0x3                    // Create 2 bit mask of core Id
    blr x5                              // Call raw exception handler

    ldr	x0, [sp], #16
    ldp	x1, x2, [sp], #16
    ldp	x3, x4, [sp], #16
    ldp	x5, x6, [sp], #16
    ldp	x7, x8, [sp], #16
    ldp	x9, x10, [sp], #16
    ldp	x11, x12, [sp], #16
    ldp	x13, x14, [sp], #16
    ldp	x15, x16, [sp], #16
    ldp	x17, x18, [sp], #16
    ldp	x19, x20, [sp], #16
    ldp	x21, x22, [sp], #16
    ldp	x23, x24, [sp], #16
    ldp	x25, x26, [sp], #16
    ldp	x27, x28, [sp], #16
    ldp	x29, x30, [sp], #16
    eret


/* Re-entrant interrupt handler stub */
/* http://infocenter.arm.com/help/index.jsp?topic=/com.arm.doc.den0024a/ch10s05.html */
ex_el2:
    stp	x29, x30, [sp, #-16]!
    stp	x27, x28, [sp, #-16]!
    stp	x25, x26, [sp, #-16]!
    stp	x23, x24, [sp, #-16]!
    stp	x21, x22, [sp, #-16]!
    stp	x19, x20, [sp, #-16]!
    stp	x17, x18, [sp, #-16]!
    stp	x15, x16, [sp, #-16]!
    stp	x13, x14, [sp, #-16]!
    stp	x11, x12, [sp, #-16]!
    stp	x9, x10, [sp, #-16]!
    stp	x7, x8, [sp, #-16]!
    stp	x5, x6, [sp, #-16]!
    stp	x3, x4, [sp, #-16]!
    stp	x1, x2, [sp, #-16]!
    str	x0, [sp, #-16]!

    ldr x5, =raw_exception_handler      // Address to exception handler
    ldr x6, [x5]
    mrs x1, esr_el1                     // get syndrome register
    mrs x2, elr_el1                     // get link register
    mrs x3, CurrentEL
    lsr x3, x3, #2
    mrs x4, mpidr_el1                   // Fetch core Id
    and x4, x4, #0x3                    // Create 2 bit mask of core Id

    blr x6                              // Call exception handler

    ldr	x0, [sp], #16
    ldp	x1, x2, [sp], #16
    ldp	x3, x4, [sp], #16
    ldp	x5, x6, [sp], #16
    ldp	x7, x8, [sp], #16
    ldp	x9, x10, [sp], #16
    ldp	x11, x12, [sp], #16
    ldp	x13, x14, [sp], #16
    ldp	x15, x16, [sp], #16
    ldp	x17, x18, [sp], #16
    ldp	x19, x20, [sp], #16
    ldp	x21, x22, [sp], #16
    ldp	x23, x24, [sp], #16
    ldp	x25, x26, [sp], #16
    ldp	x27, x28, [sp], #16
    ldp	x29, x30, [sp], #16
    eret
