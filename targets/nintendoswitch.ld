PHDRS
{
  text    PT_LOAD FLAGS(5) /* Read | Execute */;
  rodata  PT_LOAD FLAGS(4) /* Read */;
  data    PT_LOAD FLAGS(6) /* Read | Write */;
  bss     PT_LOAD FLAGS(6) /* Read | Write */;
  dynamic PT_DYNAMIC;
}

SECTIONS
{
  /* Code and file header */
  . = 0;

  .text : ALIGN(0x1000) {
    HIDDEN(__text_start = .);
    KEEP(*(.text.jmp))

    . = 0x80;
    KEEP(*(.text.start))

    *(.text .text.*)
    *(.plt .plt.*)

    HIDDEN(__text_end = .);
    HIDDEN(__text_size = . - __text_start);
  }

  /* Read-only sections */
  . = ALIGN(0x1000);

  HIDDEN(__rodata_start = .);
  .rodata : { *(.rodata .rodata.*) }  :rodata

  .mod0 : {
    KEEP(crt0.nso.o(.data.mod0))
    KEEP(crt0.nro.o(.data.mod0))
    KEEP(crt0.lib.nro.o(.data.mod0))
  }

  .dynsym            : { *(.dynsym) } :rodata
  .dynstr            : { *(.dynstr) } :rodata
  .rela.dyn          : { *(.rela.*) } :rodata

  HIDDEN(__rodata_end = .);
  HIDDEN(__rodata_size = . - __rodata_start);

  /* Read-write sections */
  . = ALIGN(0x1000);

  .data : {
    HIDDEN(__data_start = .);

    *(.data .data.*)
    *(.got .got.*)
    *(.got.plt .got.plt.*)

    HIDDEN(__data_end = .);
    HIDDEN(__data_size = . - __data_start);
  } :data

  .dynamic : {
    HIDDEN(__dynamic_start = .);
    *(.dynamic)
  }

  /* BSS section */
  . = ALIGN(0x1000);
  .bss : {
    HIDDEN(__bss_start = .);

    *(.bss .bss.*)
    *(COMMON)
    . = ALIGN(8);

    HIDDEN(__bss_end = .);
    HIDDEN(__bss_size = . - __bss_start);
  } : bss

  /DISCARD/ :
  {
    *(.eh_frame) /* This is probably unnecessary and bloats the binary. */
    *(.eh_frame_hdr)
  }
}
