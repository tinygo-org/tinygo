/* See Rust for a more complete reference: https://github.com/rp-rs/rp-hal/blob/main/rp235x-hal-examples/memory.x */
MEMORY
{
   /* 2MiB safe default.  */
    FLASH : ORIGIN = 0x10000000, LENGTH = __flash_size
    /* RAM consists of 8 banks, SRAM0..SRAM7 with striped mapping. */
    SRAM : ORIGIN = 0x20000000, LENGTH = 512k
    /* Banks 8 and 9 use direct mapping which can be
       specailized for applications where predictable access time is beneficial.
       i.e: Separate stacks for core0 and core1. */
    SRAM4 : ORIGIN = 0x20080000, LENGTH = 4k
    SRAM5 : ORIGIN = 0x20081000, LENGTH = 4k
    FLASH_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = __flash_size
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 512k
}

_stack_size = 2K;

SECTIONS
{
}

INCLUDE "targets/arm.ld"
