// This is a very minimal bootloader for the ESP32-C6. It only initializes the
// flash and then continues with the generic RISC-V initialization code, which
// in turn will call runtime.main.
// It is written in assembly (and not in a higher level language) to make sure
// it is entirely loaded into IRAM and doesn't accidentally call functions
// stored in IROM.
//
// The ESP32-C6 has a unified IRAM/DRAM address space at 0x40800000, and
// separate DROM (0x42800000) / IROM (0x42000000) flash-mapped regions.

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

    // Set up flash mapping (both IROM and DROM).
    // On ESP32-C6, Cache_Dbus_MMU_Set is replaced by Cache_MSPI_MMU_Set
    // which has an extra "sensitive" parameter.
    // C equivalent:
    //   Cache_MSPI_MMU_Set(0, 0, 0x42000000, 0, 64, 256, 0)
    // Maps 16MB starting at 0x42000000, covering both IROM and DROM.
    li a0, 0              // sensitive: no flash encryption
    li a1, 0              // ext_ram: MMU_ACCESS_FLASH
    li a2, 0x42000000     // vaddr: start of flash-mapped region
    li a3, 0              // paddr: physical address in the flash chip
    li a4, 64             // psize: always 64 (kilobytes)
    li a5, 256            // num: pages (16MB / 64K = 256, covers IROM+DROM)
    li a6, 0              // fixed
    call Cache_MSPI_MMU_Set

    // Enable the flash cache.
    mv a0, s0 // restore autoload value from Cache_Suspend_ICache call
    call Cache_Resume_ICache

    // Jump to generic RISC-V initialization, which initializes the stack
    // pointer and globals register. It should not return.
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
