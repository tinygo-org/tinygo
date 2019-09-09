SECTIONS
{
. = 0x10000;
.text : { KEEP(*(.text.boot_bl)) *(.text .text.* .gnu.linkonce.t*) }
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

/DISCARD/ : { *(.comment) *(.gnu*) *(.note*) *(.eh_frame*) *(.text.boot)}
}
__bss_size = (__bss_end - __bss_start)>>3;
