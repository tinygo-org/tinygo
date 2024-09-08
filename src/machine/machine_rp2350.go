//go:build rp2350

package machine

// GPIO pins
const (
	GPIO0  Pin = 0
	GPIO1  Pin = 1
	GPIO2  Pin = 2
	GPIO3  Pin = 3
	GPIO4  Pin = 4
	GPIO5  Pin = 5
	GPIO6  Pin = 6
	GPIO7  Pin = 7
	GPIO8  Pin = 8
	GPIO9  Pin = 9
	GPIO10 Pin = 10
	GPIO11 Pin = 11
	GPIO12 Pin = 12
	GPIO13 Pin = 13
	GPIO14 Pin = 14
	GPIO15 Pin = 15
	GPIO16 Pin = 16
	GPIO17 Pin = 17
	GPIO18 Pin = 18
	GPIO19 Pin = 19
	GPIO20 Pin = 20
	GPIO21 Pin = 21
	GPIO22 Pin = 22
	GPIO23 Pin = 23
	GPIO24 Pin = 24
	GPIO25 Pin = 25
	GPIO26 Pin = 26
	GPIO27 Pin = 27
	GPIO28 Pin = 28
	GPIO29 Pin = 29
)

/* memory.x
MEMORY {
    /*
     * The RP2350 has either external or internal flash.
     *
     * 2 MiB is a safe default here, although a Pico 2 has 4 MiB.
     * /
	 FLASH : ORIGIN = 0x10000000, LENGTH = 2048K
	 /*
	  * RAM consists of 8 banks, SRAM0-SRAM7, with a striped mapping.
	  * This is usually good for performance, as it distributes load on
	  * those banks evenly.
	  * /
	 RAM : ORIGIN = 0x20000000, LENGTH = 512K
	 /*
	  * RAM banks 8 and 9 use a direct mapping. They can be used to have
	  * memory areas dedicated for some specific job, improving predictability
	  * of access times.
	  * Example: Separate stacks for core0 and core1.
	  * /
	 SRAM4 : ORIGIN = 0x20080000, LENGTH = 4K
	 SRAM5 : ORIGIN = 0x20081000, LENGTH = 4K
 }

 SECTIONS {
	 /* ### Boot ROM info
	  *
	  * Goes after .vector_table, to keep it in the first 4K of flash
	  * where the Boot ROM (and picotool) can find it
	  * /
	 .start_block : ALIGN(4)
	 {
		 KEEP(*(.start_block));
	 } > FLASH

 } INSERT AFTER .vector_table;

 /* move .text to start /after/ the boot info * /
 _stext = ADDR(.start_block) + SIZEOF(.start_block);

 SECTIONS {
	 /* ### Picotool 'Binary Info' Entries
	  *
	  * Picotool looks through this block (as we have pointers to it in our
	  * header) to find interesting information.
	  * /
	 .bi_entries : ALIGN(4)
	 {
		 /* We put this in the header * /
		 __bi_entries_start = .;
		 /* Here are the entries * /
		 KEEP(*(.bi_entries));
		 /* Keep this block a nice round size * /
		 . = ALIGN(4);
		 /* We put this in the header * /
		 __bi_entries_end = .;
	 } > FLASH
 } INSERT AFTER .text;

 SECTIONS {
	 /* ### Boot ROM extra info
	  *
	  * Goes after everything in our program, so it can contain a signature.
	  * /
	 .end_block : ALIGN(4)
	 {
		 KEEP(*(.end_block));
	 } > FLASH

 } INSERT AFTER .bss;
*/
