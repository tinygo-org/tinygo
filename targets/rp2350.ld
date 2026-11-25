/* See Rust for a more complete reference: https://github.com/rp-rs/rp-hal/blob/main/rp235x-hal-examples/memory.x */
MEMORY
{
   /* 2MiB safe default.  */
    FLASH : ORIGIN = 0x10000000, LENGTH = __flash_size
    /* RAM consists of 8 banks, SRAM0..SRAM7 with striped mapping. */
    SRAM : ORIGIN = 0x20000000, LENGTH = 512k
    /* Banks 8 and 9 use direct mapping which can be
       specialized for applications where predictable access time is beneficial.
       i.e: Separate stacks for core0 and core1. */
    SRAM4 : ORIGIN = 0x20080000, LENGTH = 4k
    SRAM5 : ORIGIN = 0x20081000, LENGTH = 4k
    FLASH_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = __flash_size
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 512k
    /* Uncached alias of QMI window 1 (CS1), XIP_NOCACHE_NOALLOC_BASE + 16M.
       See RP2350 datasheet section 4.4.1. A board with PSRAM sets machine.psram and __psram_size. */
    PSRAM (rw)      : ORIGIN = 0x15000000, LENGTH = DEFINED(__psram_size) ? __psram_size : 0
}

_stack_size = 2K;

SECTIONS
{
    .psram (NOLOAD) : ALIGN(4)
    {
        _spsram = ABSOLUTE(.);
        *(.psram .psram.*)
        . = ALIGN(4);
        _epsram = ABSOLUTE(.);
    } > PSRAM
}

_psram_start = ORIGIN(PSRAM);
_psram_end = ORIGIN(PSRAM) + LENGTH(PSRAM);
ASSERT(LENGTH(PSRAM) <= 16M, "__psram_size is larger than the 16MB PSRAM window")

INCLUDE "targets/arm.ld"
