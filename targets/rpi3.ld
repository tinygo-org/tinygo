/*
 * This file assumes that you are running with the Raspberry PI 3/4 firmware
 * that expects bare metal to be linked for a boot address of 0x80000.
 *
 * You may want to consult https://www.raspberrypi.org/documentation/configuration/config-txt/boot.md
 * if you want to see how to force this boot address in your config.txt.
 */

SECTIONS
{
    . = 0x80000;
    .text : { KEEP(*(.text.boot)) *(.text .text.* .gnu.linkonce.t*) }
    .rodata : { *(.rodata .rodata.* .gnu.linkonce.r*) }
    PROVIDE(_data = .);
    .data : { *(.data .data.* .gnu.linkonce.d*) }
    .bss (NOLOAD) : {
        . = ALIGN(16);
        __bss_start = .;
        *(.bss .bss.*)
        *(COMMON)
        __bss_end = .;
    }
    _end = .;

   /DISCARD/ : { *(.comment) *(.gnu*) *(.note*) *(.eh_frame*) }
}
__bss_size = (__bss_end - __bss_start)>>3;
