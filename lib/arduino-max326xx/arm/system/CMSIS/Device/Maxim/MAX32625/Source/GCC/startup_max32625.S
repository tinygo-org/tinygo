/*******************************************************************************
 * Copyright (C) 2016 Maxim Integrated Products, Inc., All Rights Reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
 * OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL MAXIM INTEGRATED BE LIABLE FOR ANY CLAIM, DAMAGES
 * OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 *
 * Except as contained in this notice, the name of Maxim Integrated
 * Products, Inc. shall not be used except as stated in the Maxim Integrated
 * Products, Inc. Branding Policy.
 *
 * The mere transfer of this software does not imply any licenses
 * of trade secrets, proprietary technology, copyrights, patents,
 * trademarks, maskwork rights, or any other form of intellectual
 * property whatsoever. Maxim Integrated Products, Inc. retains all
 * ownership rights.
 *******************************************************************************
 */

    .syntax unified
    .arch armv7-m

    .section .stack
    .align 3
#ifdef __STACK_SIZE
    .equ    Stack_Size, __STACK_SIZE
#else
    .equ    Stack_Size, 0x00005000
#endif
    .globl    __StackTop
    .globl    __StackLimit
__StackLimit:
    .space    Stack_Size
    .size __StackLimit, . - __StackLimit
__StackTop:
    .size __StackTop, . - __StackTop

    .section .heap
    .align 3
#ifdef __HEAP_SIZE
    .equ    Heap_Size, __HEAP_SIZE
#else
    .equ    Heap_Size, 0x0000A000
#endif
    .globl    __HeapBase
    .globl    __HeapLimit
__HeapBase:
    .if    Heap_Size
    .space    Heap_Size
    .endif
    .size __HeapBase, . - __HeapBase
__HeapLimit:
    .size __HeapLimit, . - __HeapLimit


    .section .isr_vector
    .align 2
    .globl __isr_vector
__isr_vector:
    .long    __StackTop            /* Top of Stack */
    .long    Reset_Handler         /* Reset Handler */
    .long    NMI_Handler           /* NMI Handler */
    .long    HardFault_Handler     /* Hard Fault Handler */
    .long    MemManage_Handler     /* MPU Fault Handler */
    .long    BusFault_Handler      /* Bus Fault Handler */
    .long    UsageFault_Handler    /* Usage Fault Handler */
    .long    0                     /* Reserved */
    .long    0                     /* Reserved */
    .long    0                     /* Reserved */
    .long    0                     /* Reserved */
    .long    SVC_Handler           /* SVCall Handler */
    .long    DebugMon_Handler      /* Debug Monitor Handler */
    .long    0                     /* Reserved */
    .long    PendSV_Handler        /* PendSV Handler */
    .long    SysTick_Handler       /* SysTick Handler */

    /* MAX32625 Interrupts */
    .long    CLKMAN_IRQHandler     /* 16:01 CLKMAN */
    .long    PWRMAN_IRQHandler     /* 17:02 PWRMAN */
    .long    FLC_IRQHandler        /* 18:03 Flash Controller */
    .long    RTC0_IRQHandler       /* 19:04 RTC INT0 */
    .long    RTC1_IRQHandler       /* 20:05 RTC INT1 */
    .long    RTC2_IRQHandler       /* 21:06 RTC INT2 */
    .long    RTC3_IRQHandler       /* 22:07 RTC INT3 */
    .long    PMU_IRQHandler        /* 23:08 PMU */
    .long    USB_IRQHandler        /* 24:09 USB */
    .long    AES_IRQHandler        /* 25:10 AES */
    .long    MAA_IRQHandler        /* 26:11 MAA */
    .long    WDT0_IRQHandler       /* 27:12 WATCHDOG0 */
    .long    WDT0_P_IRQHandler     /* 28:13 WATCHDOG0 PRE-WINDOW */
    .long    WDT1_IRQHandler       /* 29:14 WATCHDOG1 */
    .long    WDT1_P_IRQHandler     /* 30:15 WATCHDOG1 PRE-WINDOW */
    .long    GPIO_P0_IRQHandler    /* 31:16 GPIO Port 0  */
    .long    GPIO_P1_IRQHandler    /* 32:17 GPIO Port 1  */
    .long    GPIO_P2_IRQHandler    /* 33:18 GPIO Port 2  */
    .long    GPIO_P3_IRQHandler    /* 34:19 GPIO Port 3  */
    .long    GPIO_P4_IRQHandler    /* 35:20 GPIO Port 4  */
    .long    GPIO_P5_IRQHandler    /* 36:21 GPIO Port 5  */
    .long    GPIO_P6_IRQHandler    /* 37:22 GPIO Port 6  */
    .long    TMR0_IRQHandler       /* 38:23 Timer32-0 */
    .long    TMR16_0_IRQHandler    /* 39:24 Timer16-s0 */
    .long    TMR1_IRQHandler       /* 40:25 Timer32-1 */
    .long    TMR16_1_IRQHandler    /* 41:26 Timer16-s1 */
    .long    TMR2_IRQHandler       /* 42:27 Timer32-2 */
    .long    TMR16_2_IRQHandler    /* 43:28 Timer16-s2 */
    .long    TMR3_IRQHandler       /* 44:29 Timer32-3 */
    .long    TMR16_3_IRQHandler    /* 45:30 Timer16-s3 */
    .long    TMR4_IRQHandler       /* 46:31 Timer32-4 */
    .long    TMR16_4_IRQHandler    /* 47:32 Timer16-s4 */
    .long    TMR5_IRQHandler       /* 48:33 Timer32-5 */
    .long    TMR16_5_IRQHandler    /* 49:34 Timer16-s5 */
    .long    UART0_IRQHandler      /* 50:35 UART0 */
    .long    UART1_IRQHandler      /* 51:36 UART1 */
    .long    UART2_IRQHandler      /* 52:37 UART2 */
    .long    UART3_IRQHandler      /* 53:38 UART3 */
    .long    PT_IRQHandler         /* 54:39 PT */
    .long    I2CM0_IRQHandler      /* 55:40 I2C Master 0 */
    .long    I2CM1_IRQHandler      /* 56:41 I2C Master 1 */
    .long    I2CM2_IRQHandler      /* 57:42 I2C Master 2 */
    .long    I2CS_IRQHandler       /* 58:43 I2C Slave */
    .long    SPIM0_IRQHandler      /* 59:44 SPIM0 */
    .long    SPIM1_IRQHandler      /* 60:45 SPIM1 */
    .long    SPIM2_IRQHandler      /* 61:46 SPIM2 */
    .long    SPIB_IRQHandler       /* 62:47 SPI Bridge */
    .long    OWM_IRQHandler        /* 63:48 One-wire Master */
    .long    AFE_IRQHandler        /* 64:49 AFE */
    .long    SPIS_IRQHandler       /* 65:50 SPI Slave  */
    .long    GPIO_P7_IRQHandler    /* 66:51 GPIO Port 7 */
    .long    GPIO_P8_IRQHandler    /* 67:52 GPIO Port 8 */


    .text 	
    .thumb
    .thumb_func
    .align 2
    .globl   Reset_Handler
    .type    Reset_Handler, %function
Reset_Handler:
    ldr r0, =__StackTop
    mov sp, r0

    ldr     r0, =PreInit
    blx     r0
    cbnz    r0, .SKIPRAMINIT

/*     Loop to copy data from read only memory to RAM. The ranges
 *      of copy from/to are specified by following symbols evaluated in
 *      linker script.
 *      _etext: End of code section, i.e., begin of data sections to copy from.
 *      _data /_edata: RAM address range that data should be
 *      copied to. Both must be aligned to 4 bytes boundary.  */

    ldr    r1, =__etext
    ldr    r2, =__data_start__
    ldr    r3, =__data_end__

#if 0
/* Here are two copies of loop implemenations. First one favors code size
 * and the second one favors performance. Default uses the first one.
 * Change to "#if 0" to use the second one */
.LC0:
    cmp     r2, r3
    ittt    lt
    ldrlt   r0, [r1], #4
    strlt   r0, [r2], #4
    blt    .LC0
#else
    subs    r3, r2
    ble    .LC1
.LC0:
    subs    r3, #4
    ldr    r0, [r1, r3]
    str    r0, [r2, r3]
    bgt    .LC0
.LC1:
#endif

/*
 *     Loop to zero out BSS section, which uses following symbols
 *     in linker script:
 *      _bss  : start of BSS section. Must align to 4
 *      _ebss : end of BSS section. Must align to 4
 */
    ldr r1, =__bss_start__
    ldr r2, =__bss_end__ 

    movs    r0, 0
.LC2:
    cmp     r1, r2
    itt    lt
    strlt   r0, [r1], #4
    blt    .LC2

.SKIPRAMINIT:

    ldr     r0, =SystemInit
    blx     r0

    ldr     r0, =_start
    blx     r0

.SPIN:
    /* Enter LP2 if main() ever returns. */
    wfi
    bl .SPIN

/*    Macro to define default handlers. Default handler
 *    will be weak symbol and just dead loops. They can be
 *    overwritten by other handlers */
    .macro    def_irq_handler    handler_name
    .align 1
    .thumb_func
    .weak    \handler_name
    .type    \handler_name, %function
\handler_name :
    b    .
    .size    \handler_name, . - \handler_name
    .endm

    def_irq_handler    NMI_Handler
    def_irq_handler    HardFault_Handler
    def_irq_handler    MemManage_Handler
    def_irq_handler    BusFault_Handler
    def_irq_handler    UsageFault_Handler
    def_irq_handler    SVC_Handler
    def_irq_handler    DebugMon_Handler
    def_irq_handler    PendSV_Handler
    def_irq_handler    SysTick_Handler
    def_irq_handler    Default_Handler

    /* MAX32625 Interrupts */
    def_irq_handler CLKMAN_IRQHandler         /* 16:01 CLKMAN */
    def_irq_handler PWRMAN_IRQHandler         /* 17:02 PWRMAN */
    def_irq_handler FLC_IRQHandler            /* 18:03 Flash Controller */
    def_irq_handler RTC0_IRQHandler           /* 19:04 RTC INT0 */
    def_irq_handler RTC1_IRQHandler           /* 20:05 RTC INT1 */
    def_irq_handler RTC2_IRQHandler           /* 21:06 RTC INT2 */
    def_irq_handler RTC3_IRQHandler           /* 22:07 RTC INT3 */
    def_irq_handler PMU_IRQHandler            /* 23:08 PMU */
    def_irq_handler USB_IRQHandler            /* 24:09 USB */
    def_irq_handler AES_IRQHandler            /* 25:10 AES */
    def_irq_handler MAA_IRQHandler            /* 26:11 MAA */
    def_irq_handler WDT0_IRQHandler           /* 27:12 WATCHDOG0 */
    def_irq_handler WDT0_P_IRQHandler         /* 28:13 WATCHDOG0 PRE-WINDOW */
    def_irq_handler WDT1_IRQHandler           /* 29:14 WATCHDOG1 */
    def_irq_handler WDT1_P_IRQHandler         /* 30:15 WATCHDOG1 PRE-WINDOW */
    def_irq_handler GPIO_P0_IRQHandler        /* 31:16 GPIO Port 0  */
    def_irq_handler GPIO_P1_IRQHandler        /* 32:17 GPIO Port 1  */
    def_irq_handler GPIO_P2_IRQHandler        /* 33:18 GPIO Port 2  */
    def_irq_handler GPIO_P3_IRQHandler        /* 34:19 GPIO Port 3  */
    def_irq_handler GPIO_P4_IRQHandler        /* 35:20 GPIO Port 4  */
    def_irq_handler GPIO_P5_IRQHandler        /* 36:21 GPIO Port 5  */
    def_irq_handler GPIO_P6_IRQHandler        /* 37:22 GPIO Port 6  */
    def_irq_handler TMR0_IRQHandler           /* 38:23 Timer32-0 */
    def_irq_handler TMR16_0_IRQHandler        /* 39:24 Timer16-s0 */
    def_irq_handler TMR1_IRQHandler           /* 40:25 Timer32-1 */
    def_irq_handler TMR16_1_IRQHandler        /* 41:26 Timer16-s1 */
    def_irq_handler TMR2_IRQHandler           /* 42:27 Timer32-2 */
    def_irq_handler TMR16_2_IRQHandler        /* 43:28 Timer16-s2 */
    def_irq_handler TMR3_IRQHandler           /* 44:29 Timer32-3 */
    def_irq_handler TMR16_3_IRQHandler        /* 45:30 Timer16-s3 */
    def_irq_handler TMR4_IRQHandler           /* 46:31 Timer32-4 */
    def_irq_handler TMR16_4_IRQHandler        /* 47:32 Timer16-s4 */
    def_irq_handler TMR5_IRQHandler           /* 48:33 Timer32-5 */
    def_irq_handler TMR16_5_IRQHandler        /* 49:34 Timer16-s5 */
    def_irq_handler PT_IRQHandler             /* 50:35 PT */
    def_irq_handler UART0_IRQHandler          /* 51:36 UART0 */
    def_irq_handler UART1_IRQHandler          /* 52:37 UART1 */
    def_irq_handler UART2_IRQHandler          /* 53:38 UART0 */
    def_irq_handler UART3_IRQHandler          /* 54:39 UART1 */
    def_irq_handler I2CM0_IRQHandler          /* 55:40 I2C Master 0 */
    def_irq_handler I2CM1_IRQHandler          /* 56:41 I2C Master 1 */
    def_irq_handler I2CM2_IRQHandler          /* 57:42 I2C Master 2 */
    def_irq_handler I2CS_IRQHandler           /* 58:43 I2C Slave */
    def_irq_handler SPIM0_IRQHandler          /* 59:44 SPIM0 */
    def_irq_handler SPIM1_IRQHandler          /* 60:45 SPIM1 */
    def_irq_handler SPIM2_IRQHandler          /* 61:46 SPIM2 */
    def_irq_handler SPIB_IRQHandler           /* 62:47 SPI Bridge */
    def_irq_handler OWM_IRQHandler            /* 63:48 SPI Bridge */
    def_irq_handler AFE_IRQHandler            /* 64:49 AFE */
    def_irq_handler SPIS_IRQHandler           /* 65:50 SPI Slave  */
    def_irq_handler GPIO_P7_IRQHandler        /* 66:51 GPIO Port 7 */
    def_irq_handler GPIO_P8_IRQHandler        /* 67:52 GPIO Port 8 */
    .end
