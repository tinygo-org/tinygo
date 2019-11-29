
/* memory map:
 * https://github.com/sifive/freedom-e-sdk/blob/v201908-branch/bsp/sifive-hifive1/metal.default.lds
 */
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x20400000, LENGTH = 0x1fc00000
    RAM (xrw)       : ORIGIN = 0x80000000, LENGTH = 0x4000
}

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
