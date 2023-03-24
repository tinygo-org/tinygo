//
// Implementation of RP2040 stage 2 boot loader.  This code is derived from the
// Winbond W25Q080 implementation (as found in the Pico) in the official Pico SDK.
//
// This implementation has been made 'stand-alone' by including necessary code /
// symbols from the included files in the reference implementation directly into
// the source.  It has also been modified to include the conditional logic from
// the CircuitPython implementation that supports additional flash chips.  The
// CircuitPython source is here:
//   https://github.com/adafruit/circuitpython/blob/main/ports/raspberrypi/stage2.c.jinja
//
// This file cannot be assembled directly, instead assemble the board-specific file
// (such as pico-boot-stage2.S) which defines the parameters specific to the flash
// chip included on that board.
//
// Care has been taken to preserve ordering and it has been verified the generated
// binary is byte-for-byte identical to the reference code binary when assembled for
// the Pico.
//
// Note: the stage 2 boot loader must be 256 bytes in length and have a checksum
// present.  In TinyGo, the linker script is responsible for allocating 256 bytes
// for the .boot2 section and the build logic patches the checksum into the
// binary after linking, controlled by the '<target>.json' flag 'rp2040-boot-patch'.
//
// The stage 2 bootstrap section can be inspected in an elf file using this command:
//     objdump -s -j .boot2 <binary>.elf
//
// Original Source:
// https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/boot_stage2/boot2_w25q080.S
//


// ----------------------------------------------------------------------------
// Second stage boot code
// Copyright (c) 2019-2021 Raspberry Pi (Trading) Ltd.
// SPDX-License-Identifier: BSD-3-Clause
//
// Device:      Winbond W25Q080
//              Also supports W25Q16JV (which has some different SR instructions)
//              Also supports AT25SF081
//              Also supports S25FL132K0
//
// Description: Configures W25Q080 to run in Quad I/O continuous read XIP mode
//
// Details:     * Check status register 2 to determine if QSPI mode is enabled,
//                and perform an SR2 programming cycle if necessary.
//              * Use SSI to perform a dummy 0xEB read command, with the mode
//                continuation bits set, so that the flash will not require
//                0xEB instruction prefix on subsequent reads.
//              * Configure SSI to write address, mode bits, but no instruction.
//                SSI + flash are now jointly in a state where continuous reads
//                can take place.
//              * Jump to exit pointer passed in via lr. Bootrom passes null,
//                in which case this code uses a default 256 byte flash offset
//
// Building:    * This code must be position-independent, and use stack only
//              * The code will be padded to a size of 256 bytes, including a
//                4-byte checksum. Therefore code size cannot exceed 252 bytes.
// ----------------------------------------------------------------------------


//
// Expanded include files
//
#define CMD_WRITE_ENABLE 0x06
#define CMD_READ_STATUS 0x05
#define CMD_READ_STATUS2 0x35
#define CMD_WRITE_STATUS1 0x01
#define CMD_WRITE_STATUS2 0x31
#define SREG_DATA 0x02  // Enable quad-SPI mode

#define XIP_BASE       0x10000000
#define XIP_SSI_BASE   0x18000000
#define PADS_QSPI_BASE 0x40020000
#define PPB_BASE       0xe0000000

#define M0PLUS_VTOR_OFFSET 0x0000ed08

#define PADS_QSPI_GPIO_QSPI_SCLK_DRIVE_LSB      4
#define PADS_QSPI_GPIO_QSPI_SCLK_SLEWFAST_BITS  0x00000001
#define PADS_QSPI_GPIO_QSPI_SCLK_OFFSET         0x00000004
#define PADS_QSPI_GPIO_QSPI_SD0_OFFSET          0x00000008
#define PADS_QSPI_GPIO_QSPI_SD0_SCHMITT_BITS    0x00000002
#define PADS_QSPI_GPIO_QSPI_SD1_OFFSET          0x0000000c
#define PADS_QSPI_GPIO_QSPI_SD2_OFFSET          0x00000010
#define PADS_QSPI_GPIO_QSPI_SD3_OFFSET          0x00000014

#define SSI_CTRLR0_OFFSET        0x00000000
#define SSI_CTRLR1_OFFSET        0x00000004
#define SSI_SSIENR_OFFSET        0x00000008
#define SSI_BAUDR_OFFSET         0x00000014
#define SSI_SR_OFFSET            0x00000028
#define SSI_DR0_OFFSET           0x00000060
#define SSI_RX_SAMPLE_DLY_OFFSET 0x000000f0

#define SSI_CTRLR0_DFS_32_LSB 16

#define SSI_CTRLR0_SPI_FRF_VALUE_QUAD 0x2
#define SSI_CTRLR0_SPI_FRF_LSB        21

#define SSI_CTRLR0_TMOD_VALUE_TX_AND_RX   0x0
#define SSI_CTRLR0_TMOD_VALUE_EEPROM_READ 0x3
#define SSI_CTRLR0_TMOD_LSB               8

#define SSI_SPI_CTRLR0_TRANS_TYPE_VALUE_1C2A 0x1
#define SSI_SPI_CTRLR0_TRANS_TYPE_VALUE_2C2A 0x2

#define SSI_SPI_CTRLR0_OFFSET 0x000000f4

#define SSI_SPI_CTRLR0_INST_L_VALUE_NONE 0x0
#define SSI_SPI_CTRLR0_INST_L_VALUE_8B   0x2

#define SSI_SPI_CTRLR0_TRANS_TYPE_LSB  0
#define SSI_SPI_CTRLR0_ADDR_L_LSB      2
#define SSI_SPI_CTRLR0_INST_L_LSB      8
#define SSI_SPI_CTRLR0_WAIT_CYCLES_LSB 11
#define SSI_SPI_CTRLR0_XIP_CMD_LSB     24

#define SSI_SR_BUSY_BITS  0x00000001
#define SSI_SR_TFE_BITS   0x00000004


// ----------------------------------------------------------------------------
// Config section
// ----------------------------------------------------------------------------
// It should be possible to support most flash devices by modifying this section

// The serial flash interface will run at clk_sys/PICO_FLASH_SPI_CLKDIV.
// This must be a positive, even integer.
// The bootrom is very conservative with SPI frequency, but here we should be
// as aggressive as possible.

#define PICO_FLASH_SPI_CLKDIV BOARD_PICO_FLASH_SPI_CLKDIV
#if PICO_FLASH_SPI_CLKDIV & 1
#error PICO_FLASH_SPI_CLKDIV must be even
#endif

#if BOARD_QUAD_OK==1
// Define interface width: single/dual/quad IO
#define FRAME_FORMAT       SSI_CTRLR0_SPI_FRF_VALUE_QUAD
#define TRANSACTION_TYPE   SSI_SPI_CTRLR0_TRANS_TYPE_VALUE_2C2A
// Note that the INST_L field is used to select what XIP data gets pushed into
// the TX FIFO:
//      INST_L_0_BITS   {ADDR[23:0],XIP_CMD[7:0]}       Load "mode bits" into XIP_CMD
//      Anything else   {XIP_CMD[7:0],ADDR[23:0]}       Load SPI command into XIP_CMD
#define INSTRUCTION_LENGTH SSI_SPI_CTRLR0_INST_L_VALUE_NONE
#define READ_INSTRUCTION   MODE_CONTINUOUS_READ
#define ADDR_L             8 // 6 for address, 2 for mode
#else
#define FRAME_FORMAT       SSI_CTRLR0_SPI_FRF_VALUE_STD
#define TRANSACTION_TYPE   SSI_SPI_CTRLR0_TRANS_TYPE_VALUE_1C1A
#define INSTRUCTION_LENGTH SSI_SPI_CTRLR0_INST_L_VALUE_8B
#define READ_INSTRUCTION   BOARD_CMD_READ
#define ADDR_L             6 // * 4 = 24
#endif

// The flash-chip specific read isntruction
#define CMD_READ BOARD_CMD_READ

// "Mode bits" are 8 special bits sent immediately after
// the address bits in a "Read Data Fast Quad I/O" command sequence. 
// On W25Q080, the four LSBs are don't care, and if MSBs == 0xa, the
// next read does not require the 0xeb instruction prefix.
#define MODE_CONTINUOUS_READ 0xa0

// How many clocks of Hi-Z following the mode bits. For W25Q080, 4 dummy cycles
// are required.
#define WAIT_CYCLES BOARD_WAIT_CYCLES


// If defined, we will read status reg, compare to SREG_DATA, and overwrite
// with our value if the SR doesn't match.
// We do a two-byte write to SR1 (01h cmd) rather than a one-byte write to
// SR2 (31h cmd) as the latter command isn't supported by WX25Q080.
// This isn't great because it will remove block protections.
// A better solution is to use a volatile SR write if your device supports it.
#define PROGRAM_STATUS_REG

.syntax unified
.cpu cortex-m0plus
.thumb
.section .boot2, "ax"

// The exit point is passed in lr. If entered from bootrom, this will be the
// flash address immediately following this second stage (0x10000100).
// Otherwise it will be a return address -- second stage being called as a
// function by user code, after copying out of XIP region. r3 holds SSI base,
// r0...2 used as temporaries. Other GPRs not used.
.global _stage2_boot
.type _stage2_boot,%function
.thumb_func
_stage2_boot:
    push {lr}

    // Set pad configuration:
    // - SCLK 8mA drive, no slew limiting
    // - SDx disable input Schmitt to reduce delay

    ldr r3, =PADS_QSPI_BASE
    movs r0, #(2 << PADS_QSPI_GPIO_QSPI_SCLK_DRIVE_LSB | PADS_QSPI_GPIO_QSPI_SCLK_SLEWFAST_BITS)
    str r0, [r3, #PADS_QSPI_GPIO_QSPI_SCLK_OFFSET]
    ldr r0, [r3, #PADS_QSPI_GPIO_QSPI_SD0_OFFSET]
    movs r1, #PADS_QSPI_GPIO_QSPI_SD0_SCHMITT_BITS
    bics r0, r1
#if BOARD_QUAD_OK==1
    str r0, [r3, #PADS_QSPI_GPIO_QSPI_SD0_OFFSET]
#endif
    str r0, [r3, #PADS_QSPI_GPIO_QSPI_SD1_OFFSET]
#if BOARD_QUAD_OK==1
    str r0, [r3, #PADS_QSPI_GPIO_QSPI_SD2_OFFSET]
    str r0, [r3, #PADS_QSPI_GPIO_QSPI_SD3_OFFSET]
#endif

    ldr r3, =XIP_SSI_BASE

    // Disable SSI to allow further config
    movs r1, #0
    str r1, [r3, #SSI_SSIENR_OFFSET]

    // Set baud rate
    movs r1, #PICO_FLASH_SPI_CLKDIV
    str r1, [r3, #SSI_BAUDR_OFFSET]

    // Set 1-cycle sample delay. If PICO_FLASH_SPI_CLKDIV == 2 then this means,
    // if the flash launches data on SCLK posedge, we capture it at the time that
    // the next SCLK posedge is launched. This is shortly before that posedge
    // arrives at the flash, so data hold time should be ok. For
    // PICO_FLASH_SPI_CLKDIV > 2 this pretty much has no effect.

    movs r1, #1
    movs r2, #SSI_RX_SAMPLE_DLY_OFFSET  // == 0xf0 so need 8 bits of offset significance
    str r1, [r3, r2]

// On QSPI parts we usually need a 01h SR-write command to enable QSPI mode
// (i.e. turn WPn and HOLDn into IO2/IO3)
#ifdef PROGRAM_STATUS_REG
program_sregs:
#define CTRL0_SPI_TXRX \
    (7 << SSI_CTRLR0_DFS_32_LSB) | /* 8 bits per data frame */ \
    (SSI_CTRLR0_TMOD_VALUE_TX_AND_RX << SSI_CTRLR0_TMOD_LSB)

    ldr r1, =(CTRL0_SPI_TXRX)
    str r1, [r3, #SSI_CTRLR0_OFFSET]

     // Enable SSI and select slave 0
    movs r1, #1
    str r1, [r3, #SSI_SSIENR_OFFSET]

    // Check whether SR needs updating
#if BOARD_QUAD_OK==1
# if BOARD_QUAD_ENABLE_STATUS_BYTE==1
    movs r0, #CMD_READ_STATUS1
# elif BOARD_QUAD_ENABLE_STATUS_BYTE==2
    movs r0, #CMD_READ_STATUS2
# endif

    bl read_flash_sreg
    movs r2, #BOARD_QUAD_ENABLE_BIT_MASK
    cmp r0, r2
    beq skip_sreg_programming

    // Send write enable command
    movs r1, #CMD_WRITE_ENABLE
    str r1, [r3, #SSI_DR0_OFFSET]

    // Poll for completion and discard RX
    bl wait_ssi_ready
    ldr r1, [r3, #SSI_DR0_OFFSET]

    // Send status write command followed by data bytes
# if BOARD_SPLIT_STATUS_WRITE==1
#  if BOARD_QUAD_ENABLE_STATUS_BYTE==1
    movs r1, #CMD_WRITE_STATUS1
#  elif BOARD_QUAD_ENABLE_STATUS_BYTE==2
    movs r1, #CMD_WRITE_STATUS2
#  endif
    str r1, [r3, #SSI_DR0_OFFSET]
    str r2, [r3, #SSI_DR0_OFFSET]

    bl wait_ssi_ready
    //ldr r1, [r3, #SSI_DR0_OFFSET]
    ldr r1, [r3, #SSI_DR0_OFFSET]
    ldr r1, [r3, #SSI_DR0_OFFSET]

# else 
    movs r1, #CMD_WRITE_STATUS1
    str r1, [r3, #SSI_DR0_OFFSET]
#  if BOARD_QUAD_ENABLE_STATUS_BYTE==2
    movs r0, #0
    str r0, [r3, #SSI_DR0_OFFSET]
#  endif
    str r2, [r3, #SSI_DR0_OFFSET]

    bl wait_ssi_ready
    ldr r1, [r3, #SSI_DR0_OFFSET]
    ldr r1, [r3, #SSI_DR0_OFFSET]
#  if BOARD_QUAD_ENABLE_STATUS_BYTE==2
    ldr r1, [r3, #SSI_DR0_OFFSET]
#  endif

# endif
    // Poll status register for write completion
1:
    movs r0, #CMD_READ_STATUS
    bl read_flash_sreg
    movs r1, #1
    tst r0, r1
    bne 1b
#endif

skip_sreg_programming:

    // Disable SSI again so that it can be reconfigured
    movs r1, #0
    str r1, [r3, #SSI_SSIENR_OFFSET]
#endif

// Currently the flash expects an 8 bit serial command prefix on every
// transfer, which is a waste of cycles. Perform a dummy Fast Read Quad I/O
// command, with mode bits set such that the flash will not expect a serial
// command prefix on *subsequent* transfers. We don't care about the results
// of the read, the important part is the mode bits.

dummy_read:
#define CTRLR0_ENTER_XIP \
    (FRAME_FORMAT                          /* Quad I/O mode */                \
        << SSI_CTRLR0_SPI_FRF_LSB) |                                          \
    (31 << SSI_CTRLR0_DFS_32_LSB)  |       /* 32 data bits */                 \
    (SSI_CTRLR0_TMOD_VALUE_EEPROM_READ     /* Send INST/ADDR, Receive Data */ \
        << SSI_CTRLR0_TMOD_LSB)

    ldr r1, =(CTRLR0_ENTER_XIP)
    str r1, [r3, #SSI_CTRLR0_OFFSET]

    movs r1, #0x0                    // NDF=0 (single 32b read)
    str r1, [r3, #SSI_CTRLR1_OFFSET]

#if BOARD_QUAD_OK==1
#define SPI_CTRLR0_ENTER_XIP \
    (ADDR_L << SSI_SPI_CTRLR0_ADDR_L_LSB) |     /* Address + mode bits */ \
    (WAIT_CYCLES << SSI_SPI_CTRLR0_WAIT_CYCLES_LSB) | /* Hi-Z dummy clocks following address + mode */ \
    (SSI_SPI_CTRLR0_INST_L_VALUE_8B \
        << SSI_SPI_CTRLR0_INST_L_LSB) |        /* 8-bit instruction */ \
    (SSI_SPI_CTRLR0_TRANS_TYPE_VALUE_1C2A      /* Send Command in serial mode then address in Quad I/O mode */ \
        << SSI_SPI_CTRLR0_TRANS_TYPE_LSB)

    ldr r1, =(SPI_CTRLR0_ENTER_XIP)
    ldr r0, =(XIP_SSI_BASE + SSI_SPI_CTRLR0_OFFSET)  // SPI_CTRL0 Register
    str r1, [r0]

    movs r1, #1                      // Re-enable SSI
    str r1, [r3, #SSI_SSIENR_OFFSET]

    movs r1, #CMD_READ
    str r1, [r3, #SSI_DR0_OFFSET]   // Push SPI command into TX FIFO
    movs r1, #MODE_CONTINUOUS_READ   // 32-bit: 24 address bits (we don't care, so 0) and M[7:4]=1010
    str r1, [r3, #SSI_DR0_OFFSET]   // Push Address into TX FIFO - this will trigger the transaction

    // Poll for completion
    bl wait_ssi_ready

// The flash is in a state where we can blast addresses in parallel, and get
// parallel data back. Now configure the SSI to translate XIP bus accesses
// into QSPI transfers of this form.

    movs r1, #0
    str r1, [r3, #SSI_SSIENR_OFFSET]   // Disable SSI (and clear FIFO) to allow further config
#endif

// Note that the INST_L field is used to select what XIP data gets pushed into
// the TX FIFO:
//      INST_L_0_BITS   {ADDR[23:0],XIP_CMD[7:0]}       Load "mode bits" into XIP_CMD
//      Anything else   {XIP_CMD[7:0],ADDR[23:0]}       Load SPI command into XIP_CMD
configure_ssi:
#define SPI_CTRLR0_XIP \
    (READ_INSTRUCTION                          /* Mode bits to keep flash in continuous read mode */ \
        << SSI_SPI_CTRLR0_XIP_CMD_LSB) | \
    (ADDR_L << SSI_SPI_CTRLR0_ADDR_L_LSB) |    /* Total number of address + mode bits */ \
    (WAIT_CYCLES << SSI_SPI_CTRLR0_WAIT_CYCLES_LSB) |    /* Hi-Z dummy clocks following address + mode */ \
    (INSTRUCTION_LENGTH                        /* Do not send a command, instead send XIP_CMD as mode bits after address */ \
        << SSI_SPI_CTRLR0_INST_L_LSB) | \
    (TRANSACTION_TYPE                          /* Send Address in Quad I/O mode (and Command but that is zero bits long) */ \
        << SSI_SPI_CTRLR0_TRANS_TYPE_LSB)

    ldr r1, =(SPI_CTRLR0_XIP)

    ldr r0, =(XIP_SSI_BASE + SSI_SPI_CTRLR0_OFFSET)
    str r1, [r0]

    movs r1, #1
    str r1, [r3, #SSI_SSIENR_OFFSET]   // Re-enable SSI

// Bus accesses to the XIP window will now be transparently serviced by the
// external flash on cache miss. We are ready to run code from flash.


//
// Helper Includes
//

//
// #include "boot2_helpers/exit_from_boot2.S"
//

// If entered from the bootrom, lr (which we earlier pushed) will be 0,
// and we vector through the table at the start of the main flash image.
// Any regular function call will have a nonzero value for lr.
check_return:
    pop {r0}
    cmp r0, #0
    beq vector_into_flash
    bx r0
vector_into_flash:
    ldr r0, =(XIP_BASE + 0x100)
    ldr r1, =(PPB_BASE + M0PLUS_VTOR_OFFSET)
    str r0, [r1]
    ldmia r0, {r0, r1}
    msr msp, r0
    bx r1

//
// #include "boot2_helpers/wait_ssi_ready.S"
//
wait_ssi_ready:
    push {r0, r1, lr}

    // Command is complete when there is nothing left to send
    // (TX FIFO empty) and SSI is no longer busy (CSn deasserted)
1:
    ldr r1, [r3, #SSI_SR_OFFSET]
    movs r0, #SSI_SR_TFE_BITS
    tst r1, r0
    beq 1b
    movs r0, #SSI_SR_BUSY_BITS
    tst r1, r0
    bne 1b

    pop {r0, r1, pc}


#ifdef PROGRAM_STATUS_REG

//
// #include "boot2_helpers/read_flash_sreg.S"
//

// Pass status read cmd into r0.
// Returns status value in r0.
.global read_flash_sreg
.type read_flash_sreg,%function
.thumb_func
read_flash_sreg:
    push {r1, lr}
    str r0, [r3, #SSI_DR0_OFFSET]
    // Dummy byte:
    str r0, [r3, #SSI_DR0_OFFSET]
    
    bl wait_ssi_ready
    // Discard first byte and combine the next two
    ldr r0, [r3, #SSI_DR0_OFFSET]
    ldr r0, [r3, #SSI_DR0_OFFSET]

    pop {r1, pc}

#endif

.global literals
literals:
.ltorg

.end
