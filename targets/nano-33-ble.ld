
/*
  See also 
  https://github.com/arduino/ArduinoCore-mbed/blob/master/variants/ARDUINO_NANO33BLE/linker_script.ld
*/
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x10000, LENGTH = 0xf0000
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 0x40000
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
