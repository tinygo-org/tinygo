
/* memory map:
 * https://github.com/sifive/freedom-e-sdk/blob/v201908-branch/bsp/sifive-hifive1/metal.default.lds
 * With one change: this linkerscript uses 64K RAM instead of 16K as specified
 * in metal.default.lds. Not sure why this works, but it works, and it avoids
 * out-of-memory issues when running tests.
 */
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x20400000, LENGTH = 0x1fc00000
    RAM (xrw)       : ORIGIN = 0x80000000, LENGTH = 64K
}

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
