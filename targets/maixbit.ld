
MEMORY
{
    /* This is a software workaround to avoid using the medany code model
       which causes the llvm9 build to fail. 0x80000000 is mirorred at
       0xffffffff80000000. This is not needed anymore on llvm10.
       https://github.com/rust-embedded/riscv-rt/issues/25 */
    RAM (xrw) : ORIGIN = 0xffffffff80000000, LENGTH = 6M
}

REGION_ALIAS("FLASH_TEXT", RAM);

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
