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
.syntax unified
.cpu cortex-m33
.fpu fpv5-sp-d16
.thumb
#endif
.endm

// do not put align in here as it is used mid function sometimes
.macro regular_func x
.global \x
.type \x,%function
#ifndef __riscv
.thumb_func
#endif
\x:
.endm

.macro weak_func x
.weak \x
.type \x,%function
#ifndef __riscv
.thumb_func
#endif
\x:
.endm

.macro regular_func_with_section x
.section .text.\x
regular_func \x
.endm

// do not put align in here as it is used mid function sometimes
.macro wrapper_func x
regular_func WRAPPER_FUNC_NAME(\x)
.endm

.macro weak_wrapper_func x
weak_func WRAPPER_FUNC_NAME(\x)
.endm

.macro __pre_init_with_offset func, offset, priority_string1
.section .preinit_array.\priority_string1
.p2align 2
.word \func + \offset
.endm

# backwards compatibility
.macro __pre_init func, priority_string1
__pre_init_with_offset func, 0, priority_string1
.endm

#ifdef __riscv
// rd = (rs1 >> rs2[4:0]) & ~(-1 << nbits)
.macro h3.bextm rd rs1 rs2 nbits
.if (\nbits < 1) || (\nbits > 8)
.err
.endif
    .insn r 0x0b, 0x4, (((\nbits - 1) & 0x7 ) << 1), \rd, \rs1, \rs2
.endm

// rd = (rs1 >> shamt) & ~(-1 << nbits)
.macro h3.bextmi rd rs1 shamt nbits
.if (\nbits < 1) || (\nbits > 8)
.err
.endif
.if (\shamt < 0) || (\shamt > 31)
.err
.endif
    .insn i 0x0b, 0x4, \rd, \rs1, (\shamt & 0x1f) | (((\nbits - 1) & 0x7 ) << 6)
.endm
#endif


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


#define INIT_M0_TIMING (1  << QMI_M0_TIMING_COOLDOWN_LSB |PICO_FLASH_SPI_RXDELAY << QMI_M0_TIMING_RXDELAY_LSB |PICO_FLASH_SPI_CLKDIV  << QMI_M0_TIMING_CLKDIV_LSB |0)

// Set command constants
#define INIT_M0_RCMD (CMD_READ << QMI_M0_RCMD_PREFIX_LSB | 0)

// Set read format to all-serial with a command prefix
#define INIT_M0_RFMT ((QMI_M0_RFMT_PREFIX_WIDTH_VALUE_S << QMI_M0_RFMT_PREFIX_WIDTH_LSB) | (QMI_M0_RFMT_ADDR_WIDTH_VALUE_S   << QMI_M0_RFMT_ADDR_WIDTH_LSB) | (QMI_M0_RFMT_SUFFIX_WIDTH_VALUE_S << QMI_M0_RFMT_SUFFIX_WIDTH_LSB) | (QMI_M0_RFMT_DUMMY_WIDTH_VALUE_S  << QMI_M0_RFMT_DUMMY_WIDTH_LSB) | (QMI_M0_RFMT_DATA_WIDTH_VALUE_S   << QMI_M0_RFMT_DATA_WIDTH_LSB) | (QMI_M0_RFMT_PREFIX_LEN_VALUE_8   << QMI_M0_RFMT_PREFIX_LEN_LSB) | 0)

// ----------------------------------------------------------------------------
// Start of 2nd Stage Boot Code
// ----------------------------------------------------------------------------

pico_default_asm_setup

.section .text

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
.else
    push {lr}
    ldr r3, =XIP_QMI_BASE
    ldr r0, =INIT_M0_TIMING
    str r0, [r3, #QMI_M0_TIMING_OFFSET]
    ldr r0, =INIT_M0_RCMD
    str r0, [r3, #QMI_M0_RCMD_OFFSET]
    ldr r0, =INIT_M0_RFMT
    str r0, [r3, #QMI_M0_RFMT_OFFSET]
.endif

// Pull in standard exit routine
// #include "boot2_helpers/exit_from_boot2.S" // https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2350/boot_stage2/asminclude/boot2_helpers/exit_from_boot2.S
.ifdef __riscv
    jr t0
.else
    pop {pc}
.endif

.ifndef __riscv
.global literals
literals:
.ltorg
.endif