
/* Memory map:
 * https://github.com/qemu/qemu/blob/master/hw/riscv/virt.c
 * Looks like we can use any address starting from 0x80000000 (so 2GB of space).
 * However, using a large space slows down tests.
 */
MEMORY
{
    RAM (rwx) : ORIGIN = 0x80000000, LENGTH = 100M
}

REGION_ALIAS("FLASH_TEXT", RAM)

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
