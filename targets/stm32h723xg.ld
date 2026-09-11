
MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x08000000, LENGTH = 1024K
  RAM (xrw)       : ORIGIN = 0x24000000, LENGTH = 128K
}

/*
 * D2 SRAM (0x30000000, 32K on H723) and D3 SRAM (0x38000000, 16K) are not
 * mapped here and unused by the runtime. If a future driver places DMA
 * buffers there, add MPU regions for them in initMPU()
 * (src/runtime/runtime_stm32h7_mpu.go) — they currently fall through to the
 * default privileged WBWA map, so cache maintenance via DCacheClean/
 * DCacheInvalidate/DCacheFlush would silently miss those addresses.
 */

_stack_size = 8K;

INCLUDE "targets/arm.ld"
