
/* Memory map:
 * https://github.com/qemu/qemu/blob/master/hw/riscv/virt.c
 * RAM and flash are set to 1MB each. That should be enough for the foreseeable
 * future. QEMU does not seem to limit the flash/RAM size and in fact doesn't
 * seem to differentiate between it.
 */
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x20000000, LENGTH = 0x100000
    RAM (xrw)       : ORIGIN = 0x80000000, LENGTH = 0x100000
}

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
