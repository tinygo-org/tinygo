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
    .text . : { KEEP(*(.text.boot)) *(.init) *(.text .text.* .gnu.linkonce.t*) }
    .rodata : { *(.rodata .rodata.* .gnu.linkonce.r*) }
    PROVIDE(_data = .);
    .data :
    {
        . = ALIGN(16);
        _sdata = .;
        *(.data)
        *(.data*)
        . = ALIGN(16);
        _edata = .;
    }
    . = ALIGN(0x1000);
    .exc . : {
        . = ALIGN(0x1000);
        vector_start = . ;
        KEEP(*(.exc_vector));
    }
    .bss (NOLOAD) : {
        . = ALIGN(16);
        _sbss = .;
        *(.bss .bss.*)
        *(COMMON)
		. = ALIGN(16);
        _ebss = .;
    }
    _end = .;

        .stack_core0 : {
        . = ALIGN(16);                          /* Stack must always be aligned to 16 byte boundary AAPCS64 call standard */
        __stack_start_core0__ = .;
        . = . + 512;                            /* EL0 stack size */
        __EL0_stack_core0 = .;
        . = . + 16384;                          /* EL1 stack size */
        __EL1_stack_core0 = .;
        . = . + 512;                            /* EL2 stack size (start-up) */
        __EL2_stack_core0 = .;
        __stack_end_core0__ = .;
    }

        .stack_core1 : {
        . = ALIGN(16);                          /* Stack must always be aligned to 16 byte boundary AAPCS64 call standard */
        __stack_start_core1__ = .;
        . = . + 512;                            /* EL0 stack size */
        __EL0_stack_core1 = .;
                . = . + 1024;                           /* EL1 stack size */
        __EL1_stack_core1 = .;
        . = . + 512;                            /* EL2 stack size (start-up) */
        __EL2_stack_core1 = .;
        __stack_end_core1__ = .;
    }

        .stack_core2 :  {
        . = ALIGN(16);                           /* Stack must always be aligned to 16 byte boundary AAPCS call standard */
        __stack_start_core2__ = .;
        . = . + 512;                            /* EL0 stack size */
        __EL0_stack_core2 = .;
        . = . + 1024;                           /* EL1 stack size */
        __EL1_stack_core2 = .;
        . = . + 512;                            /* EL2 stack size (start-up) */
        __EL2_stack_core2 = .;
        __stack_end_core2__ = .;
    }

        .stack_core3 :  {
        . = ALIGN(16);                          /* Stack must always be aligned to 16 byte boundary AAPCS call standard */
        __stack_start_core3__ = .;
        . = . + 512;                            /* EL0 stack size */
        __EL0_stack_core3 = .;
        . = . + 1024;                           /* EL1 stack size */
        __EL1_stack_core3 = .;
        . = . + 512;                            /* EL2 stack size (start-up) */
        __EL2_stack_core3 = .;
        __stack_end_core3__ = .;
    }


   /DISCARD/ : { *(.comment) *(.gnu*) *(.note*) *(.eh_frame*)}
}

/* For the memory allocator. */
_globals_start = _sdata;
_globals_end = _ebss;
_stack_top = 0x80000;
_heap_start = _ebss;
_heap_end = _heap_start + 0x200000; /* 2Mb*/
