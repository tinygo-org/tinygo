
MEMORY
{
    /* This SoftDevice requires 96K flash and 8K RAM according to the release
     * notes of version 8.0.0 */
    FLASH_TEXT (rw) : ORIGIN = 0x00000000 + 96K, LENGTH = 256K - 96K
    RAM (xrw)       : ORIGIN = 0x20000000 + 8K,  LENGTH = 16K  - 8K
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
