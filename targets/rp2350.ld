/* Reference: https://github.com/rp-rs/rp-hal/blob/main/rp235x-hal-examples/memory.x */
MEMORY {
    /* 2MiB safe default.  */
    FLASH : ORIGIN = 0x10000000, LENGTH = 2048K
    /* RAM consists of 8 banks, SRAM0..SRAM7 with striped mapping. */
    RAM : ORIGIN = 0x20000000, LENGTH = 512K
    /* Banks 8 and 9 use direct mapping which can be
       specailized for applications where predictable access time is beneficial.
       i.e: Separate stacks for core0 and core1. */
    SRAM4 : ORIGIN = 0x20080000, LENGTH = 4K
    SRAM5 : ORIGIN = 0x20081000, LENGTH = 4K
 }

 SECTIONS {
    /* Boot ROM info.
       Goes after .vector_table to keep it in first 4K of flash
       where Boot ROM and picotool can find it. */
    .start_block : ALIGN(4)
    {
        __start_block_addr = .;
        KEEP(*(.start_block));
    } > FLASH
 } INSERT AFTER .vector_table;

 /* move .text to start /after/ the boot info. */
 _stext = ADDR(.start_block) + SIZEOF(.start_block);

 SECTIONS {
    /* Picotool 'Binary Info' Entries. 
       Picotool looks through this block (as we have pointers to it in our
       header) to find interesting information. */
     .bi_entries : ALIGN(4)
     {
         __bi_entries_start = .;
         KEEP(*(.bi_entries));   /* Put entries. */
         . = ALIGN(4);           /* Keep a nice round size. */
         __bi_entries_end = .;
     } > FLASH
 } INSERT AFTER .text;

 SECTIONS {
    /* Boot ROM extra info. Goes after everything in program. */
    .end_block : ALIGN(4)
    {
        __end_block_addr = .;
        KEEP(*(.end_block));
    } > FLASH
 } INSERT AFTER .bss;

PROVIDE(start_to_end = __end_block_addr - __start_block_addr);
PROVIDE(end_to_start = __start_block_addr - __end_block_addr);
