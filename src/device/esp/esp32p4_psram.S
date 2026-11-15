// ESP32-P4 ROM wrappers for the PSRAM user-command path.
//
// ESP-IDF's AP-Hex PSRAM driver performs the chip mode-register transactions
// through the mask-ROM functions esp_rom_spi_set_op_mode /
// esp_rom_spi_cmd_config / esp_rom_spi_cmd_start.  Re-using them keeps the
// DTR/OPI line-mode setup and the bit-length encodings identical to the
// working IDF flow instead of duplicating (and subtly mis-encoding) them.
//
// This file is a target extra-file (see targets/esp32p4.json), not a package
// source file: TinyGo's package loader does not assemble .S files found inside
// Go packages.
//
// Addresses from components/esp_rom/esp32p4/ld/esp32p4.rom.ld (IDF 5.5.5).

.section .text

// psram_rom_set_op_mode(spiNum, mode)
.global psram_rom_set_op_mode
psram_rom_set_op_mode:
    addi sp, sp, -16
    sw   ra, 12(sp)
    li   t0, 0x4fc00110        // esp_rom_spi_set_op_mode
    jalr ra, 0(t0)
    lw   ra, 12(sp)
    addi sp, sp, 16
    ret

// psram_rom_cmd_config(spiNum, *esp_rom_spi_cmd_t)
.global psram_rom_cmd_config
psram_rom_cmd_config:
    addi sp, sp, -16
    sw   ra, 12(sp)
    li   t0, 0x4fc00108        // esp_rom_spi_cmd_config
    jalr ra, 0(t0)
    lw   ra, 12(sp)
    addi sp, sp, 16
    ret

// psram_rom_cmd_start(spiNum, rxBuf, rxLen, csMask, isWriteErase)
.global psram_rom_cmd_start
psram_rom_cmd_start:
    addi sp, sp, -16
    sw   ra, 12(sp)
    li   t0, 0x4fc0010c        // esp_rom_spi_cmd_start
    jalr ra, 0(t0)
    lw   ra, 12(sp)
    addi sp, sp, 16
    ret
