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
    .text : {
                KEEP(*(.text.boot))
                KEEP(*(.text))
                KEEP(*(.text.*))
                *(.gnu.linkonce.t*)
    }

    .rodata : {
        *(.rodata .rodata.* .gnu.linkonce.r*)
    }

	/* Globals with initial value */
    .data : {
        *(.data)
        *(.data*)
    }

	/* zero initialized data is initialized by preinit */
    .bss (NOLOAD) : {
        _sbss = .;
        *(.bss .bss.*)
        *(COMMON)
        _ebss = .;
    }

    . = 0x96000;
   .exc  : {
        KEEP(*(.exc_vector))
     }
   /DISCARD/ : { *(.comment) *(.gnu*) *(.note*) *(.eh_frame*)  }
}

_heap_start = 0x400000;
_heap_end = 0x1000000;