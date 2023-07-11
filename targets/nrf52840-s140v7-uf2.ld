/*
  See also 
  https://github.com/Seeed-Studio/ArduinoCore-mbed/blob/master/variants/SEEED_XIAO_NRF52840_SENSE/linker_script.ld
*/
MEMORY
{
    FLASH_TEXT (rx) : ORIGIN = 0x27000, LENGTH = 0xED000 - 0x27000
    RAM (rwx) : ORIGIN = 0x20006000, LENGTH = 0x3A000
}

_stack_size = 4K + __softdevice_stack;

/* This value is needed by the Nordic SoftDevice. */
__app_ram_base = ORIGIN(RAM);
__softdevice_stack = DEFINED(__softdevice_stack) ? __softdevice_stack : 0;

INCLUDE "targets/arm.ld"
