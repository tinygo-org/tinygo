// This is a very minimal bootloader for the ESP32-C3. It only initializes the
// flash and then continues with the generic RISC-V initialization code, which
// in turn will call runtime.main.
// It is written in assembly (and not in a higher level language) to make sure
// it is entirely loaded into IRAM and doesn't accidentally call functions
// stored in IROM.
//
// For reference, here is a nice introduction into RISC-V assembly:
// https://www.imperialviolet.org/2016/12/31/riscv.html

.section .init
.global call_start_cpu0
.type call_start_cpu0,@function
call_start_cpu0:
    // At this point:
    // - The ROM bootloader is finished and has jumped to here.
    // - We're running from IRAM: both IRAM and DRAM segments have been loaded
    //   by the ROM bootloader.
    // - We have a usable stack (but not the one we would like to use).
    // - No flash mappings (MMU) are set up yet.

    // Reset MMU, see bootloader_reset_mmu in the ESP-IDF.
    call Cache_Suspend_ICache
    mv s0, a0 // autoload value
    call Cache_Invalidate_ICache_All
    call Cache_MMU_Init

    // Set up DROM from flash.
    // Somehow, this also sets up IROM from flash. Not sure why, but it avoids
    // the need for another such call.
    // C equivalent:
    //   Cache_Dbus_MMU_Set(MMU_ACCESS_FLASH, 0x3C00_0000, 0, 64, 128, 0)
    li a0, 0              // ext_ram: MMU_ACCESS_FLASH
    li a1, 0x3C000000     // vaddr: address in the data bus
    li a2, 0              // paddr: physical address in the flash chip
    li a3, 64             // psize: always 64 (kilobytes)
    li a4, 128            // num: pages to be set (8192K / 64K = 128)
    li a5, 0              // fixed
    call Cache_Dbus_MMU_Set

    // Enable the flash cache.
    mv a0, s0 // restore autoload value from Cache_Suspend_ICache call
    call Cache_Resume_ICache

    // Jump to generic RISC-V initialization, which initializes the stack
    // pointer and globals register. It should not return.
    // (It appears that the linker relaxes this jump and instead inserts the
    // _start function right after here).
    j _start

.section .text.exception_vectors
.global _vector_table
.type _vector_table,@function

_vector_table:

    .option push
    .option norvc

    .rept 32
    j handleInterruptASM            /* interrupt handler */
    .endr

    .option pop

.size _vector_table, .-_vector_table

