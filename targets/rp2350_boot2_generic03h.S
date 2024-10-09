// ----------------------------------------------------------------------------
// Second stage boot code
// Copyright (c) 2019-2021 Raspberry Pi (Trading) Ltd.
// SPDX-License-Identifier: BSD-3-Clause
//
// Device:      Anything which responds to 03h serial read command
//
// Details:     * Configure SSI to translate each APB read into a 03h command
//              * 8 command clocks, 24 address clocks and 32 data clocks
//              * This enables you to boot from almost anything: you can pretty
//                much solder a potato to your PCB, or a piece of cheese
//              * The tradeoff is performance around 3x worse than QSPI XIP
//
// Building:    * This code must be position-independent, and use stack only
//              * The code will be padded to a size of 256 bytes, including a
//                4-byte checksum. Therefore code size cannot exceed 252 bytes.
// ----------------------------------------------------------------------------

// #include "pico.h" // https://github.com/raspberrypi/pico-sdk/blob/master/src/boards/include/boards/pico.h
// Figure out what boot uses...

// #include "pico/asm_helper.S" // https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2350/pico_platform/include/pico/asm_helper.S
#if !PICO_ASSEMBLER_IS_CLANG
#define apsr_nzcv r15
#endif
# note we don't do this by default in this file for backwards comaptibility with user code
# that may include this file, but not use unified syntax. Note that this macro does equivalent
# setup to the pico_default_asm macro for inline assembly in C code.
.macro pico_default_asm_setup
#ifndef __riscv
.syntax unified         // Selects Unified Assembly Language (UAL) syntax for assembler. Unifies ARM and thumb into single syntax.
.cpu cortex-m33         // Specify target CPU architecture.
.fpu fpv5-sp-d16        // Specify the type of FPU available on CPU. Supports SP (single precision) operations with 16 double precision registers.
.thumb                  // Assemble subsequent code with Thumb instruction set, the compact 16-bit encoding of the most frequently used 32-bit ARM instructions.
.section .boot2, "ax"   // Defines a new section named .boot2 with attributes 'a': Allocatable-This section occupies space in memory image, and 'x': This section contains executable code.
// Code and data assembled after this point will be placed in .boot2 section. The Linker script can specify where this section should be located in memory, such as at a specific address.
// We need to ensure this code is placed at the very start of flash. See rp2350.ld, .boot2 is placed at 0x10000000 and has max length 256.
// The compiled code cannot exceed 256 bytes!
// One can check contents by compiling and using objdump to visualize the code:
//  tinygo build -target=rp2350 -o=rp2350.elf -serial=none examples/blinky1
//  objdump -s -j .boot2 rp2350.elf
#endif
.endm

.macro regular_func fnname
.global \fnname             // Declares fnname as a global symbol.
.type \fnname,%function     // Specifies fnname is a type of function. Helps tools understand fnname is executable code, not data.
#ifndef __riscv
.thumb_func                 // Indicate fnname is a thumb function. 
#endif
\fnname:                    // Mark point of entry for function.
.endm


// #include "hardware/platform_defs.h" // https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2350/hardware_regs/include/hardware/platform_defs.h
#ifndef _u
#ifdef __ASSEMBLER__
#define _u(x) x
#else
#define _u(x) x ## u
#endif
#endif

// #include "hardware/regs/addressmap.h" https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2350/hardware_regs/include/hardware/regs/addressmap.h
// Register address offsets for atomic RMW aliases
#define REG_ALIAS_RW_BITS  (_u(0x0) << _u(12))
#define REG_ALIAS_XOR_BITS (_u(0x1) << _u(12))
#define REG_ALIAS_SET_BITS (_u(0x2) << _u(12))
#define REG_ALIAS_CLR_BITS (_u(0x3) << _u(12))

#define ROM_BASE _u(0x00000000)
#define XIP_BASE _u(0x10000000)
#define XIP_SRAM_BASE _u(0x13ffc000)
#define XIP_END _u(0x14000000)
#define XIP_NOCACHE_NOALLOC_BASE _u(0x14000000)
#define XIP_SRAM_END _u(0x14000000)
#define XIP_NOCACHE_NOALLOC_END _u(0x18000000)
#define XIP_MAINTENANCE_BASE _u(0x18000000)
#define XIP_NOCACHE_NOALLOC_NOTRANSLATE_BASE _u(0x1c000000)
#define SRAM0_BASE _u(0x20000000)
#define XIP_NOCACHE_NOALLOC_NOTRANSLATE_END _u(0x20000000)
#define SRAM_BASE _u(0x20000000)
#define SRAM_STRIPED_BASE _u(0x20000000)
#define SRAM4_BASE _u(0x20040000)
#define SRAM8_BASE _u(0x20080000)
#define SRAM_STRIPED_END _u(0x20080000)
#define SRAM_SCRATCH_X_BASE _u(0x20080000)
#define SRAM9_BASE _u(0x20081000)
#define SRAM_SCRATCH_Y_BASE _u(0x20081000)
#define SRAM_END _u(0x20082000)
#define SYSINFO_BASE _u(0x40000000)
#define SYSCFG_BASE _u(0x40008000)
#define CLOCKS_BASE _u(0x40010000)
#define PSM_BASE _u(0x40018000)
#define RESETS_BASE _u(0x40020000)
#define IO_BANK0_BASE _u(0x40028000)
#define IO_QSPI_BASE _u(0x40030000)
#define PADS_BANK0_BASE _u(0x40038000)
#define PADS_QSPI_BASE _u(0x40040000)
#define XOSC_BASE _u(0x40048000)
#define PLL_SYS_BASE _u(0x40050000)
#define PLL_USB_BASE _u(0x40058000)
#define ACCESSCTRL_BASE _u(0x40060000)
#define BUSCTRL_BASE _u(0x40068000)
#define UART0_BASE _u(0x40070000)
#define UART1_BASE _u(0x40078000)
#define SPI0_BASE _u(0x40080000)
#define SPI1_BASE _u(0x40088000)
#define I2C0_BASE _u(0x40090000)
#define I2C1_BASE _u(0x40098000)
#define ADC_BASE _u(0x400a0000)
#define PWM_BASE _u(0x400a8000)
#define TIMER0_BASE _u(0x400b0000)
#define TIMER1_BASE _u(0x400b8000)
#define HSTX_CTRL_BASE _u(0x400c0000)
#define XIP_CTRL_BASE _u(0x400c8000)
#define XIP_QMI_BASE _u(0x400d0000)
#define WATCHDOG_BASE _u(0x400d8000)
#define BOOTRAM_BASE _u(0x400e0000)
#define BOOTRAM_END _u(0x400e0400)
#define ROSC_BASE _u(0x400e8000)
#define TRNG_BASE _u(0x400f0000)
#define SHA256_BASE _u(0x400f8000)
#define POWMAN_BASE _u(0x40100000)
#define TICKS_BASE _u(0x40108000)
#define OTP_BASE _u(0x40120000)
#define OTP_DATA_BASE _u(0x40130000)
#define OTP_DATA_RAW_BASE _u(0x40134000)
#define OTP_DATA_GUARDED_BASE _u(0x40138000)
#define OTP_DATA_RAW_GUARDED_BASE _u(0x4013c000)
#define CORESIGHT_PERIPH_BASE _u(0x40140000)
#define CORESIGHT_ROMTABLE_BASE _u(0x40140000)
#define CORESIGHT_AHB_AP_CORE0_BASE _u(0x40142000)
#define CORESIGHT_AHB_AP_CORE1_BASE _u(0x40144000)
#define CORESIGHT_TIMESTAMP_GEN_BASE _u(0x40146000)
#define CORESIGHT_ATB_FUNNEL_BASE _u(0x40147000)
#define CORESIGHT_TPIU_BASE _u(0x40148000)
#define CORESIGHT_CTI_BASE _u(0x40149000)
#define CORESIGHT_APB_AP_RISCV_BASE _u(0x4014a000)
#define DFT_BASE _u(0x40150000)
#define GLITCH_DETECTOR_BASE _u(0x40158000)
#define TBMAN_BASE _u(0x40160000)
#define DMA_BASE _u(0x50000000)
#define USBCTRL_BASE _u(0x50100000)
#define USBCTRL_DPRAM_BASE _u(0x50100000)
#define USBCTRL_REGS_BASE _u(0x50110000)
#define PIO0_BASE _u(0x50200000)
#define PIO1_BASE _u(0x50300000)
#define PIO2_BASE _u(0x50400000)
#define XIP_AUX_BASE _u(0x50500000)
#define HSTX_FIFO_BASE _u(0x50600000)
#define CORESIGHT_TRACE_BASE _u(0x50700000)
#define SIO_BASE _u(0xd0000000)
#define SIO_NONSEC_BASE _u(0xd0020000)
#define PPB_BASE _u(0xe0000000)
#define PPB_NONSEC_BASE _u(0xe0020000)
#define EPPB_BASE _u(0xe0080000)

// #include "hardware/regs/qmi.h" // https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2350/hardware_regs/include/hardware/regs/qmi.h
// Thats a lot of content...
#define QMI_M0_TIMING_CLKDIV_LSB    _u(0)
#define QMI_M0_TIMING_CLKDIV_BITS   _u(0x000000ff)
#define QMI_M0_TIMING_RXDELAY_LSB    _u(8)
#define QMI_M0_TIMING_RXDELAY_BITS   _u(0x00000700)
#define QMI_M0_TIMING_COOLDOWN_RESET  _u(0x1)
#define QMI_M0_TIMING_COOLDOWN_BITS   _u(0xc0000000)
#define QMI_M0_TIMING_COOLDOWN_MSB    _u(31)
#define QMI_M0_TIMING_COOLDOWN_LSB    _u(30)
#define QMI_M0_TIMING_COOLDOWN_ACCESS "RW"
#define QMI_M0_RCMD_PREFIX_RESET  _u(0x03)
#define QMI_M0_RCMD_PREFIX_BITS   _u(0x000000ff)
#define QMI_M0_RCMD_PREFIX_MSB    _u(7)
#define QMI_M0_RCMD_PREFIX_LSB    _u(0)
#define QMI_M0_RCMD_PREFIX_ACCESS "RW"
#define QMI_M0_RFMT_PREFIX_WIDTH_RESET  _u(0x0)
#define QMI_M0_RFMT_PREFIX_WIDTH_BITS   _u(0x00000003)
#define QMI_M0_RFMT_PREFIX_WIDTH_MSB    _u(1)
#define QMI_M0_RFMT_PREFIX_WIDTH_LSB    _u(0)
#define QMI_M0_RFMT_PREFIX_WIDTH_ACCESS "RW"
#define QMI_M0_RFMT_PREFIX_WIDTH_VALUE_S _u(0x0)
#define QMI_M0_RFMT_PREFIX_WIDTH_VALUE_D _u(0x1)
#define QMI_M0_RFMT_PREFIX_WIDTH_VALUE_Q _u(0x2)

#define QMI_M0_RFMT_ADDR_WIDTH_RESET  _u(0x0)
#define QMI_M0_RFMT_ADDR_WIDTH_BITS   _u(0x0000000c)
#define QMI_M0_RFMT_ADDR_WIDTH_MSB    _u(3)
#define QMI_M0_RFMT_ADDR_WIDTH_LSB    _u(2)
#define QMI_M0_RFMT_ADDR_WIDTH_ACCESS "RW"
#define QMI_M0_RFMT_ADDR_WIDTH_VALUE_S _u(0x0)
#define QMI_M0_RFMT_ADDR_WIDTH_VALUE_D _u(0x1)
#define QMI_M0_RFMT_ADDR_WIDTH_VALUE_Q _u(0x2)

#define QMI_M0_RFMT_SUFFIX_WIDTH_RESET  _u(0x0)
#define QMI_M0_RFMT_SUFFIX_WIDTH_BITS   _u(0x00000030)
#define QMI_M0_RFMT_SUFFIX_WIDTH_MSB    _u(5)
#define QMI_M0_RFMT_SUFFIX_WIDTH_LSB    _u(4)
#define QMI_M0_RFMT_SUFFIX_WIDTH_ACCESS "RW"
#define QMI_M0_RFMT_SUFFIX_WIDTH_VALUE_S _u(0x0)
#define QMI_M0_RFMT_SUFFIX_WIDTH_VALUE_D _u(0x1)
#define QMI_M0_RFMT_SUFFIX_WIDTH_VALUE_Q _u(0x2)

#define QMI_M0_RFMT_DUMMY_WIDTH_RESET  _u(0x0)
#define QMI_M0_RFMT_DUMMY_WIDTH_BITS   _u(0x000000c0)
#define QMI_M0_RFMT_DUMMY_WIDTH_MSB    _u(7)
#define QMI_M0_RFMT_DUMMY_WIDTH_LSB    _u(6)
#define QMI_M0_RFMT_DUMMY_WIDTH_ACCESS "RW"
#define QMI_M0_RFMT_DUMMY_WIDTH_VALUE_S _u(0x0)
#define QMI_M0_RFMT_DUMMY_WIDTH_VALUE_D _u(0x1)
#define QMI_M0_RFMT_DUMMY_WIDTH_VALUE_Q _u(0x2)

#define QMI_M0_RFMT_DATA_WIDTH_RESET  _u(0x0)
#define QMI_M0_RFMT_DATA_WIDTH_BITS   _u(0x00000300)
#define QMI_M0_RFMT_DATA_WIDTH_MSB    _u(9)
#define QMI_M0_RFMT_DATA_WIDTH_LSB    _u(8)
#define QMI_M0_RFMT_DATA_WIDTH_ACCESS "RW"
#define QMI_M0_RFMT_DATA_WIDTH_VALUE_S _u(0x0)
#define QMI_M0_RFMT_DATA_WIDTH_VALUE_D _u(0x1)
#define QMI_M0_RFMT_DATA_WIDTH_VALUE_Q _u(0x2)

#define QMI_M0_RFMT_PREFIX_LEN_RESET  _u(0x1)
#define QMI_M0_RFMT_PREFIX_LEN_BITS   _u(0x00001000)
#define QMI_M0_RFMT_PREFIX_LEN_MSB    _u(12)
#define QMI_M0_RFMT_PREFIX_LEN_LSB    _u(12)
#define QMI_M0_RFMT_PREFIX_LEN_ACCESS "RW"
#define QMI_M0_RFMT_PREFIX_LEN_VALUE_NONE _u(0x0)
#define QMI_M0_RFMT_PREFIX_LEN_VALUE_8 _u(0x1)

#define QMI_M0_RCMD_OFFSET _u(0x00000014)
#define QMI_M0_RCMD_BITS   _u(0x0000ffff)
#define QMI_M0_RCMD_RESET  _u(0x0000a003)

#define QMI_M0_TIMING_OFFSET _u(0x0000000c)
#define QMI_M0_TIMING_BITS   _u(0xf3fff7ff)
#define QMI_M0_TIMING_RESET  _u(0x40000004)

#define QMI_M0_RFMT_OFFSET _u(0x00000010)
#define QMI_M0_RFMT_BITS   _u(0x1007d3ff)
#define QMI_M0_RFMT_RESET  _u(0x00001000)

// ----------------------------------------------------------------------------
// Config section
// ----------------------------------------------------------------------------
// It should be possible to support most flash devices by modifying this section

// The serial flash interface will run at clk_sys/PICO_FLASH_SPI_CLKDIV.
// This must be a positive integer.
// The bootrom is very conservative with SPI frequency, but here we should be
// as aggressive as possible.

#ifndef PICO_FLASH_SPI_CLKDIV
#define PICO_FLASH_SPI_CLKDIV 4
#endif
#if (PICO_FLASH_SPI_CLKDIV << QMI_M0_TIMING_CLKDIV_LSB) & ~QMI_M0_TIMING_CLKDIV_BITS
#error "CLKDIV greater than maximum"
#endif

// RX sampling delay is measured in units of one half clock cycle.

#ifndef PICO_FLASH_SPI_RXDELAY
#define PICO_FLASH_SPI_RXDELAY 1
#endif
#if (PICO_FLASH_SPI_RXDELAY << QMI_M0_TIMING_RXDELAY_LSB) & ~QMI_M0_TIMING_RXDELAY_BITS
#error "RX delay greater than maximum"
#endif

#define CMD_READ 0x03

// ----------------------------------------------------------------------------
// Register initialisation values -- same in Arm/RISC-V code.
// ----------------------------------------------------------------------------

// The QMI is automatically configured for 03h XIP straight out of reset,
// but this code can't assume it's still in that state. Set up memory
// window 0 for 03h serial reads.

// Setup timing parameters: short sequential-access cooldown, configured
// CLKDIV and RXDELAY, and no constraints on CS max assertion, CS min
// deassertion, or page boundary burst breaks.


#define INIT_M0_TIMING ((1  << QMI_M0_TIMING_COOLDOWN_LSB) |(PICO_FLASH_SPI_RXDELAY << QMI_M0_TIMING_RXDELAY_LSB) |(PICO_FLASH_SPI_CLKDIV << QMI_M0_TIMING_CLKDIV_LSB) |0)

// Set command constants
#define INIT_M0_RCMD (CMD_READ << QMI_M0_RCMD_PREFIX_LSB | 0)

// Set read format to all-serial with a command prefix
#define INIT_M0_RFMT ((QMI_M0_RFMT_PREFIX_WIDTH_VALUE_S << QMI_M0_RFMT_PREFIX_WIDTH_LSB) | (QMI_M0_RFMT_ADDR_WIDTH_VALUE_S << QMI_M0_RFMT_ADDR_WIDTH_LSB) | (QMI_M0_RFMT_SUFFIX_WIDTH_VALUE_S << QMI_M0_RFMT_SUFFIX_WIDTH_LSB) | (QMI_M0_RFMT_DUMMY_WIDTH_VALUE_S  << QMI_M0_RFMT_DUMMY_WIDTH_LSB) | (QMI_M0_RFMT_DATA_WIDTH_VALUE_S << QMI_M0_RFMT_DATA_WIDTH_LSB) | (QMI_M0_RFMT_PREFIX_LEN_VALUE_8 << QMI_M0_RFMT_PREFIX_LEN_LSB) | 0)


// load_r0 loads the register's value at address regaddr into R0 by stepping on r1 with the regaddr.
.macro load_r0 regaddr
    ldr   r1, =\regaddr // Load immediate value of regaddr with = pseudoinstr. May translate into several instructions depending on size of integer.
    ldr   r0, [r1]      // Load memory at addr described by a register.
.endm


// fload_r0 loads the register's value at address regaddr into R0 and leaves rest of registers in same state.
.macro fload_r0 regaddr
    push {r1}  // Pushes contents of r1 onto stack. Stack grows downward in Cortex arch.
    load_r0 regaddr
    pop {r1}   // Pops stack onto register r1. Braces can contain a range of addresses.
.endm

// store_off_r3 is a helper that stores value into a register at address given by r3+off. Uses r0.
.macro store_off_r3 value, off
    ldr r0, =\value       // Store literal value we want to store at address r3+off into r0.
    str r0, [r3, #\off]   // Store value r0 into register at r3+off.
.endm

// wait_bitclr checks value at address specified by register rx against bitmask until the AND between them yields 0. Steps on r0.
.macro wait_bitclr rx, bitmask
1:
    tst     \rx, #\bitmask  // Perform bitwise AND between register and operand (register or immediate value) updating condition flags. Z flag set if AND yields zero.
    bne     1b             // Conditionally branch if Z flag set. BNE: Branch if Not Equal -> Will branch if Z set, so if r0&bitmask == 0.
.endm

#define USESIO SIO_BASE
#define SIO_GPIOOUTSET_OFFSET _u(0x18)
#define SIO_GPIOOUTCLR_OFFSET _u(0x20)
#define pinfnSIO _u(1<<5)
#define pinfnUART _u(1<<2)
.macro pin_init pin, pinfn
    #define pinmask (1<<\pin)
    #define SIO_GPIOOECLR_OFFSET _u(0x40)
    #define SIO_GPIOOESET_OFFSET _u(0x38)
    #define pdtctrl_replace_bits _u(0x40|)
    #define PADSBNK_IOPIN_OFFSET _u(4+4*\pin)
    #define IOCTLREG _u(IO_BANK0_BASE+8*\pin+4)
    #define padinputenablemsk _u(0x40)
    #define padoutuptdisablemsk _u(0x80)
    // First load in padbank io reg and replace bits in it.
    ldr r3, =PADS_BANK0_BASE
    ldr r0, [r3, PADSBNK_IOPIN_OFFSET]
    bic r0, r0, (padinputenablemsk|padoutuptdisablemsk) // r0 &^= mask
    orr r0, r0, padinputenablemsk
    str r0, [r3, PADSBNK_IOPIN_OFFSET]
    // Set the pin function in IO_BANK0.
    ldr r3, =IOCTLREG
    ldr r0, =\pinfn
    str r0, [r3]
    // Clear pinstate. 
    ldr r3, =USESIO
    ldr r0, =pinmask
    str r0, [r3, SIO_GPIOOECLR_OFFSET]  // GPIO Output enable 
    str r0, [r3, SIO_GPIOOUTCLR_OFFSET] // Set pin to low.
    .if \pinfn == pinfnSIO
    str r0, [r3, SIO_GPIOOESET_OFFSET] // Assume pin is output: enable output.
    .endif
.endm

.macro pinout_init pin
    pin_init \pin, pinfnSIO
.endm

.macro pin_set pin, value
    #define pinmsk _u(1<<\pin)
    ldr r1, =pinmsk
    ldr r0, =USESIO
    .if \value == 0
        str r1, [r0, SIO_GPIOOUTCLR_OFFSET]
    .else
        str r1, [r0, SIO_GPIOOUTSET_OFFSET]
    .endif
.endm

.macro uart_init txpin
    #define RESET_UART0 _u(0x4000000)
    ldr r3, =RESETS_BASE
    ldr r0, [r3]                   // Load value @r3 into r0.
    orr r0, r0, #RESET_UART0        // Set reset bit for UART0.
    str r0, [r3]                   // Storing the reset bit resets UART.
    bic r0, r0, #RESET_UART0        // Unset the reset bit
    str r0, [r3]                   // Clear the UART0 reset bit.
    wait_bitclr r3, RESET_UART0    // Wait until peripheral is fully reset.
    
    // baud supported between ~200..6452000 
    #define baud 115200
    #define div (8*125000000/baud)
    #define ibrd _u(div>>7)
    #define fbrd _u(((div&0x7f)+1)/2)
    #define UART_IBRD_OFFSET _u(0x24)
    #define UART_FBRD_OFFSET _u(0x28)
    #define UART_LCRH_OFFSET _u(0x2C)
    ldr r3, =UART0_BASE
    // Start setting baud.
    store_off_r3 ibrd, UART_IBRD_OFFSET
    store_off_r3 fbrd, UART_FBRD_OFFSET
    ldr r0, [r3, UART_LCRH_OFFSET] // Needs dummy write with contents in LCR.
    str r0, [r3, UART_LCRH_OFFSET] // write back what we read.
    #define uartenable (0x1)
    #define uarttxenable (0x100)
    #define UARTSETTINGS _u(uartenable|uarttxenable)
    #define UART_CR_OFFSET _u(0x30)
    ldr r0, [r3, UART_CR_OFFSET]
    ldr r1, =UARTSETTINGS 
    orr r0, r0, r1         // Cannot use UARTSETTINGS as literal here, exceeds size of 255.
    str r0, [r3, UART_CR_OFFSET]
    // Configure pin 0 as UART (is Tx pin).
    pin_init \txpin, pinfnUART
.endm

.macro uart_write value
    #define UART_FR_OFFSET _u(0x18)
    ldr r0, =\value // Mark as immediate value, needs to be constant at compile time.
    ldr r1, =UART0_BASE
    str r0, [r1, UART_FR_OFFSET] // Write value in r0 to tx fifo.
.endm

// ----------------------------------------------------------------------------
// Start of 2nd Stage Boot Code
// ----------------------------------------------------------------------------

pico_default_asm_setup

// On RP2350 boot stage2 is always called as a regular function, and should return normally
regular_func _stage2_boot
.ifdef __riscv
    mv t0, ra
    li a3, XIP_QMI_BASE
    li a0, INIT_M0_TIMING
    sw a0, QMI_M0_TIMING_OFFSET(a3)
    li a0, INIT_M0_RCMD
    sw a0, QMI_M0_RCMD_OFFSET(a3)
    li a0, INIT_M0_RFMT
    sw a0, QMI_M0_RFMT_OFFSET(a3)
    // Exit routine.
    jr t0
.else
    // Save link register for exit. lr hold the return address when a function call is made (bl or blx).
    // We save it because we entered this bootload from the boot ROM, thus lr will contain the flash address 
    // immediately after this second-stage bootloader code. If bootloader is called by user lr will contain return address
    // to resume execution after the bootloader ends.
    // The bootloader uses r0..r1 as scratch and r3 holds the current register-of-interest base address. No other GPRs are used.
    push {lr}
    #define LED _u(25)
    #define txpin _u(0)
    // pinout_init LED
    // pin_set LED, 1
    ldr r3, =XIP_QMI_BASE
    // Note: STORE_OFF_R3 steps on r0.
    // Here we configure/initialize QMI.
    store_off_r3 INIT_M0_TIMING, QMI_M0_TIMING_OFFSET
    store_off_r3 INIT_M0_RCMD, QMI_M0_RCMD_OFFSET
    store_off_r3 INIT_M0_RFMT, QMI_M0_RFMT_OFFSET
    // We pop the link value we saved in push {lr} above and set the program counter to it. This effectively returns from the function
    // and returns to the calling function, this is because loading into pc causes the execution to jump to that address.
    pop {pc}
    // Declares the local symbol literals as a global symbol, making it accessible to the linker and other object files.
    // ltorg instructs assembler to emit the literal pool at this point in the code.
    // Places all pending literals (constants that cannot be encoded directly into instructions) at this point.    //
    // This ensures any ldr pseudo-instructions that reference literals have the corresponding data placed here.
    // For example, when using the Thumb1 instruction set (16bit encoding) the ldr instruction offset range is limited to 0 to 1020 bytes (increments of 4 bytes),
    // this is because it encodes offset as an 8 bit integer. If using Thumb2 (32bit) offset has range 0..4095, with 1 byte increment.
    //  
    // We need to specify where we place the literals since we also need to have tight control of our bootloader size, limited to 256 bytes including 4 byte checksum.
    // At the time of writing this bootloader literals section is empty, all instructions are fully encoded above.
    .global literals
    literals:
    .ltorg
.endif

