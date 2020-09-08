SECTIONS
{
  . = 0;

  /* Code and file header */

  .text : {
    HIDDEN(__text_start = .);
    KEEP(*(.text.jmp))

    . = 0x80;

    *(.text .text.*)

    . = ALIGN(0x1000);
    HIDDEN(__text_end = .);
    HIDDEN(__text_size = . - __text_start);
  }

  /* Read-only sections */

  .rodata : {
    HIDDEN(__rodata_start = .);

    *(.rodata .rodata.*)

    *(.got)

    KEEP(crt0.nso.o(.data.mod0))
    KEEP(crt0.nro.o(.data.mod0))
    KEEP(crt0.lib.nro.o(.data.mod0))
    KEEP(*(.data.mod0))

    HIDDEN(__dynamic_start = .);
    *(.dynamic)

    . = ALIGN(0x1000);
    HIDDEN(__rodata_end = .);
    HIDDEN(__rodata_size = . - __rodata_start);
  }

  /* Read-write sections */

  .data : {
    HIDDEN(__data_start = .);

    *(.data .data.*)

    HIDDEN(__data_end = .);
    HIDDEN(__data_size = . - __data_start);
  }

  /* BSS section */

  .bss : {
    HIDDEN(__bss_start = .);

    *(.bss .bss.*)
    *(COMMON)

    HIDDEN(__bss_end = .);
    HIDDEN(__bss_size = . - __bss_start);
  }

  /DISCARD/ :
  {
    *(.eh_frame) /* This is probably unnecessary and bloats the binary. */
  }
}
