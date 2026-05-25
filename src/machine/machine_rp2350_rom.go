//go:build tinygo && rp2350

package machine

import (
	"runtime/interrupt"
	"unsafe"
)

/*
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned long uint32_t;
typedef unsigned long size_t;
typedef unsigned long uintptr_t;
typedef long int intptr_t;

typedef const volatile uint16_t io_ro_16;
typedef const volatile uint32_t io_ro_32;
typedef volatile uint16_t io_rw_16;
typedef volatile uint32_t io_rw_32;
typedef volatile uint32_t io_wo_32;

#define false 0
#define true 1
typedef int bool;

#define ram_func __attribute__((section(".ramfuncs"),noinline))

static inline __attribute__((always_inline)) void __compiler_memory_barrier(void) {
    __asm__ volatile ("" : : : "memory");
}

extern char __flash_size;

// https://datasheets.raspberrypi.com/rp2350/rp2350-datasheet.pdf
// 13.9. Predefined OTP Data Locations
// OTP_DATA: FLASH_DEVINFO Register

#define OTP_DATA_FLASH_DEVINFO_CS0_SIZE_BITS 0x0F00
#define OTP_DATA_FLASH_DEVINFO_CS0_SIZE_LSB  8
#define OTP_DATA_FLASH_DEVINFO_CS1_SIZE_BITS 0xF000
#define OTP_DATA_FLASH_DEVINFO_CS1_SIZE_LSB  12


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_regs/include/hardware/regs/addressmap.h

#define REG_ALIAS_RW_BITS  (0x0 << 12)
#define REG_ALIAS_XOR_BITS (0x1 << 12)
#define REG_ALIAS_SET_BITS (0x2 << 12)
#define REG_ALIAS_CLR_BITS (0x3 << 12)

#define XIP_BASE     0x10000000
#define XIP_QMI_BASE 0x400d0000
#define IO_QSPI_BASE 0x40030000
#define BOOTRAM_BASE 0x400e0000

// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/hardware_base/include/hardware/address_mapped.h

#define hw_alias_check_addr(addr) ((uintptr_t)(addr))
#define hw_set_alias_untyped(addr) ((void *)(REG_ALIAS_SET_BITS + hw_alias_check_addr(addr)))
#define hw_clear_alias_untyped(addr) ((void *)(REG_ALIAS_CLR_BITS + hw_alias_check_addr(addr)))
#define hw_xor_alias_untyped(addr) ((void *)(REG_ALIAS_XOR_BITS + hw_alias_check_addr(addr)))

__attribute__((always_inline))
static void hw_set_bits(io_rw_32 *addr, uint32_t mask) {
    *(io_rw_32 *) hw_set_alias_untyped((volatile void *) addr) = mask;
}

__attribute__((always_inline))
static void hw_clear_bits(io_rw_32 *addr, uint32_t mask) {
    *(io_rw_32 *) hw_clear_alias_untyped((volatile void *) addr) = mask;
}

__attribute__((always_inline))
static void hw_xor_bits(io_rw_32 *addr, uint32_t mask) {
    *(io_rw_32 *) hw_xor_alias_untyped((volatile void *) addr) = mask;
}

__attribute__((always_inline))
static void hw_write_masked(io_rw_32 *addr, uint32_t values, uint32_t write_mask) {
    hw_xor_bits(addr, (*addr ^ values) & write_mask);
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_platform_compiler/include/pico/platform/compiler.h

#define pico_default_asm_volatile(...) __asm volatile (".syntax unified\n" __VA_ARGS__)


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/pico_platform/include/pico/platform.h

static bool pico_processor_state_is_nonsecure(void) {
//    // todo add a define to disable NS checking at all?
//    // IDAU-Exempt addresses return S=1 when tested in the Secure state,
//    // whereas executing a tt in the NonSecure state will always return S=0.
//    uint32_t tt;
//    pico_default_asm_volatile (
//        "movs %0, #0\n"
//        "tt %0, %0\n"
//        : "=r" (tt) : : "cc"
//    );
//    return !(tt & (1u << 22));

    return false;
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/include/pico/bootrom_constants.h

// RP2040 & RP2350
#define ROM_DATA_SOFTWARE_GIT_REVISION          ROM_TABLE_CODE('G', 'R')
#define ROM_FUNC_FLASH_ENTER_CMD_XIP            ROM_TABLE_CODE('C', 'X')
#define ROM_FUNC_FLASH_EXIT_XIP                 ROM_TABLE_CODE('E', 'X')
#define ROM_FUNC_FLASH_FLUSH_CACHE              ROM_TABLE_CODE('F', 'C')
#define ROM_FUNC_CONNECT_INTERNAL_FLASH         ROM_TABLE_CODE('I', 'F')
#define ROM_FUNC_FLASH_RANGE_ERASE              ROM_TABLE_CODE('R', 'E')
#define ROM_FUNC_FLASH_RANGE_PROGRAM            ROM_TABLE_CODE('R', 'P')

// RP2350 only
#define ROM_FUNC_PICK_AB_PARTITION              ROM_TABLE_CODE('A', 'B')
#define ROM_FUNC_CHAIN_IMAGE                    ROM_TABLE_CODE('C', 'I')
#define ROM_FUNC_EXPLICIT_BUY                   ROM_TABLE_CODE('E', 'B')
#define ROM_FUNC_FLASH_RUNTIME_TO_STORAGE_ADDR  ROM_TABLE_CODE('F', 'A')
#define ROM_DATA_FLASH_DEVINFO16_PTR            ROM_TABLE_CODE('F', 'D')
#define ROM_FUNC_FLASH_OP                       ROM_TABLE_CODE('F', 'O')
#define ROM_FUNC_GET_B_PARTITION                ROM_TABLE_CODE('G', 'B')
#define ROM_FUNC_GET_PARTITION_TABLE_INFO       ROM_TABLE_CODE('G', 'P')
#define ROM_FUNC_GET_SYS_INFO                   ROM_TABLE_CODE('G', 'S')
#define ROM_FUNC_GET_UF2_TARGET_PARTITION       ROM_TABLE_CODE('G', 'U')
#define ROM_FUNC_LOAD_PARTITION_TABLE           ROM_TABLE_CODE('L', 'P')
#define ROM_FUNC_OTP_ACCESS                     ROM_TABLE_CODE('O', 'A')
#define ROM_DATA_PARTITION_TABLE_PTR            ROM_TABLE_CODE('P', 'T')
#define ROM_FUNC_FLASH_RESET_ADDRESS_TRANS      ROM_TABLE_CODE('R', 'A')
#define ROM_FUNC_REBOOT                         ROM_TABLE_CODE('R', 'B')
#define ROM_FUNC_SET_ROM_CALLBACK               ROM_TABLE_CODE('R', 'C')
#define ROM_FUNC_SECURE_CALL                    ROM_TABLE_CODE('S', 'C')
#define ROM_FUNC_SET_NS_API_PERMISSION          ROM_TABLE_CODE('S', 'P')
#define ROM_FUNC_BOOTROM_STATE_RESET            ROM_TABLE_CODE('S', 'R')
#define ROM_FUNC_SET_BOOTROM_STACK              ROM_TABLE_CODE('S', 'S')
#define ROM_DATA_SAVED_XIP_SETUP_FUNC_PTR       ROM_TABLE_CODE('X', 'F')
#define ROM_FUNC_FLASH_SELECT_XIP_READ_MODE     ROM_TABLE_CODE('X', 'M')
#define ROM_FUNC_VALIDATE_NS_BUFFER             ROM_TABLE_CODE('V', 'B')

#define BOOTSEL_FLAG_GPIO_PIN_SPECIFIED         0x20

#define BOOTROM_FUNC_TABLE_OFFSET 0x14

// todo remove this (or #ifdef it for A1/A2)
#define BOOTROM_IS_A2() ((*(volatile uint8_t *)0x13) == 2)
#define BOOTROM_WELL_KNOWN_PTR_SIZE (BOOTROM_IS_A2() ? 2 : 4)

#define BOOTROM_VTABLE_OFFSET 0x00
#define BOOTROM_TABLE_LOOKUP_OFFSET     (BOOTROM_FUNC_TABLE_OFFSET + BOOTROM_WELL_KNOWN_PTR_SIZE)


// https://github.com/raspberrypi/pico-sdk
// src/common/boot_picoboot_headers/include/boot/picoboot_constants.h

// values 0-7 are secure/non-secure
#define REBOOT2_FLAG_REBOOT_TYPE_NORMAL       0x0 // param0 = diagnostic partition
#define REBOOT2_FLAG_REBOOT_TYPE_BOOTSEL      0x2 // param0 = bootsel_flags, param1 = gpio_config
#define REBOOT2_FLAG_REBOOT_TYPE_RAM_IMAGE    0x3 // param0 = image_base, param1 = image_end
#define REBOOT2_FLAG_REBOOT_TYPE_FLASH_UPDATE 0x4 // param0 = update_base

#define REBOOT2_FLAG_NO_RETURN_ON_SUCCESS    0x100

#define RT_FLAG_FUNC_ARM_SEC    0x0004
#define RT_FLAG_FUNC_ARM_NONSEC 0x0010
#define RT_FLAG_DATA            0x0040


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/include/pico/bootrom.h

#define ROM_TABLE_CODE(c1, c2) ((c1) | ((c2) << 8))

typedef void *(*rom_table_lookup_fn)(uint32_t code, uint32_t mask);

__attribute__((always_inline))
static void *rom_func_lookup_inline(uint32_t code) {
    rom_table_lookup_fn rom_table_lookup = (rom_table_lookup_fn) (uintptr_t)*(uint16_t*)(BOOTROM_TABLE_LOOKUP_OFFSET);
    if (pico_processor_state_is_nonsecure()) {
        return rom_table_lookup(code, RT_FLAG_FUNC_ARM_NONSEC);
    } else {
        return rom_table_lookup(code, RT_FLAG_FUNC_ARM_SEC);
    }
}

__attribute__((always_inline))
static void *rom_data_lookup_inline(uint32_t code) {
    rom_table_lookup_fn rom_table_lookup = (rom_table_lookup_fn) (uintptr_t)*(uint16_t*)(BOOTROM_TABLE_LOOKUP_OFFSET);
    return rom_table_lookup(code, RT_FLAG_DATA);
}

typedef int (*rom_reboot_fn)(uint32_t flags, uint32_t delay_ms, uint32_t p0, uint32_t p1);

__attribute__((always_inline))
int rom_reboot(uint32_t flags, uint32_t delay_ms, uint32_t p0, uint32_t p1) {
    rom_reboot_fn func = (rom_reboot_fn) rom_func_lookup_inline(ROM_FUNC_REBOOT);
    return func(flags, delay_ms, p0, p1);
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/bootrom.c

void reset_usb_boot(uint32_t usb_activity_gpio_pin_mask, uint32_t disable_interface_mask) {
    uint32_t flags = disable_interface_mask;
    if (usb_activity_gpio_pin_mask) {
        flags |= BOOTSEL_FLAG_GPIO_PIN_SPECIFIED;
        // the parameter is actually the gpio number, but we only care if BOOTSEL_FLAG_GPIO_PIN_SPECIFIED
        usb_activity_gpio_pin_mask = (uint32_t)__builtin_ctz(usb_activity_gpio_pin_mask);
    }
    rom_reboot(REBOOT2_FLAG_REBOOT_TYPE_BOOTSEL | REBOOT2_FLAG_NO_RETURN_ON_SUCCESS, 10, flags, usb_activity_gpio_pin_mask);
    __builtin_unreachable();
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_regs/include/hardware/regs/qmi.h

#define QMI_DIRECT_CSR_EN_BITS          0x00000001
#define QMI_DIRECT_CSR_ASSERT_CS0N_BITS 0x00000004
#define QMI_DIRECT_CSR_RXEMPTY_BITS     0x00010000
#define QMI_DIRECT_CSR_TXFULL_BITS      0x00000400
#define QMI_M1_WFMT_RESET               0x00001000
#define QMI_M1_WCMD_RESET               0x0000a002


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_regs/include/hardware/regs/io_qspi.h

#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_BITS       0x00003000
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_LSB        12
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_VALUE_LOW  0x2
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_VALUE_HIGH 0x3


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_structs/include/hardware/structs/io_qspi.h

typedef struct {
    io_rw_32 inte; // IO_QSPI_PROC0_INTE
    io_rw_32 intf; // IO_QSPI_PROC0_INTF
    io_ro_32 ints; // IO_QSPI_PROC0_INTS
} io_qspi_irq_ctrl_hw_t;

typedef struct {
    io_ro_32 status; // IO_QSPI_GPIO_QSPI_SCLK_STATUS
    io_rw_32 ctrl;   // IO_QSPI_GPIO_QSPI_SCLK_CTRL
} io_qspi_status_ctrl_hw_t;

typedef struct {
    io_ro_32 usbphy_dp_status;                  // IO_QSPI_USBPHY_DP_STATUS
    io_rw_32 usbphy_dp_ctrl;                    // IO_QSPI_USBPHY_DP_CTRL
    io_ro_32 usbphy_dm_status;                  // IO_QSPI_USBPHY_DM_STATUS
    io_rw_32 usbphy_dm_ctrl;                    // IO_QSPI_USBPHY_DM_CTRL
    io_qspi_status_ctrl_hw_t io[6];
    uint32_t _pad0[112];
    io_ro_32 irqsummary_proc0_secure;           // IO_QSPI_IRQSUMMARY_PROC0_SECURE
    io_ro_32 irqsummary_proc0_nonsecure;        // IO_QSPI_IRQSUMMARY_PROC0_NONSECURE
    io_ro_32 irqsummary_proc1_secure;           // IO_QSPI_IRQSUMMARY_PROC1_SECURE
    io_ro_32 irqsummary_proc1_nonsecure;        // IO_QSPI_IRQSUMMARY_PROC1_NONSECURE
    io_ro_32 irqsummary_dormant_wake_secure;    // IO_QSPI_IRQSUMMARY_DORMANT_WAKE_SECURE
    io_ro_32 irqsummary_dormant_wake_nonsecure; // IO_QSPI_IRQSUMMARY_DORMANT_WAKE_NONSECURE
    io_rw_32 intr;                              // IO_QSPI_INTR

    union {
        struct {
            io_qspi_irq_ctrl_hw_t proc0_irq_ctrl;
            io_qspi_irq_ctrl_hw_t proc1_irq_ctrl;
            io_qspi_irq_ctrl_hw_t dormant_wake_irq_ctrl;
        };
        io_qspi_irq_ctrl_hw_t irq_ctrl[3];
    };
} io_qspi_hw_t;

#define io_qspi_hw ((io_qspi_hw_t *)IO_QSPI_BASE)


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_structs/include/hardware/structs/qmi.h

typedef struct {
    io_rw_32 timing; // QMI_M0_TIMING
    io_rw_32 rfmt;   // QMI_M0_RFMT
    io_rw_32 rcmd;   // QMI_M0_RCMD
    io_rw_32 wfmt;   // QMI_M0_WFMT
    io_rw_32 wcmd;   // QMI_M0_WCMD
} qmi_mem_hw_t;

typedef struct {
    io_rw_32 direct_csr; // QMI_DIRECT_CSR
    io_wo_32 direct_tx;  // QMI_DIRECT_TX
    io_ro_32 direct_rx;  // QMI_DIRECT_RX
    qmi_mem_hw_t m[2];
    io_rw_32 atrans[8];  // QMI_ATRANS0
} qmi_hw_t;

#define qmi_hw ((qmi_hw_t *)XIP_QMI_BASE)


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/hardware_flash/include/hardware/flash.h

#define FLASH_PAGE_SIZE (1u << 8)
#define FLASH_SECTOR_SIZE (1u << 12)
#define FLASH_BLOCK_SIZE (1u << 16)


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/hardware_flash/flash.c

#define FLASH_BLOCK_ERASE_CMD 0xd8

// Configure QMI M0 for standard single-SPI read (0x03 cmd, reset-value register config),
// then clear direct mode so XIP auto-mode resumes. Flash must already be in SPI mode.
static ram_func void flash_restore_xip_spi_mode(void) {
    qmi_hw->m[0].rfmt = 0x00001000u; // PREFIX_LEN=1 (8-bit cmd), all widths=single, no dummy
    qmi_hw->m[0].rcmd = 0x0000a003u; // command prefix = 0x03 (standard read)
    hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_EN_BITS);
    __compiler_memory_barrier();
}

// This is a static symbol because the layout of FLASH_DEVINFO is liable to change from device to
// device, so fields must have getters/setters.
static io_rw_16 * ram_func flash_devinfo_ptr(void) {
    // Note the lookup returns a pointer to a 32-bit pointer literal in the ROM
    io_rw_16 **p = (io_rw_16 **) rom_data_lookup_inline(ROM_DATA_FLASH_DEVINFO16_PTR);
    return *p;
}

// This is a RAM function because may be called during flash programming to enable save/restore of
// QMI window 1 registers on RP2350:
uint8_t ram_func __attribute__((used)) flash_devinfo_get_cs_size(uint8_t cs) {
    volatile uint8_t local_cs = cs;
    if (local_cs == 0u) {
        uint32_t flash_size = (uint32_t)&__flash_size;
        if (flash_size == 0) {
            return 0; // FLASH_DEVINFO_SIZE_NONE
        }
        // Size code = __builtin_ctz(flash_size / 8192) + 1
        // Where FLASH_DEVINFO_SIZE_8K is 1
        return (uint8_t)(__builtin_ctz(flash_size / 8192) + 1);
    } else {
        io_ro_16 *devinfo = (io_ro_16 *) flash_devinfo_ptr();
        return (uint8_t) (
            (*devinfo & OTP_DATA_FLASH_DEVINFO_CS1_SIZE_BITS) >> OTP_DATA_FLASH_DEVINFO_CS1_SIZE_LSB
        );
    }
}

// Saves/restores QMI window 1 registers (timing, rcmd, rfmt). Window 0 is
// reconfigured by flash_restore_xip_spi_mode(); window 1 (CS1) is preserved here.
typedef struct flash_rp2350_qmi_save_state {
    uint32_t timing;
    uint32_t rcmd;
    uint32_t rfmt;
} flash_rp2350_qmi_save_state_t;

static ram_func void flash_rp2350_save_qmi_cs1(flash_rp2350_qmi_save_state_t *state) {
    state->timing = qmi_hw->m[1].timing;
    state->rcmd = qmi_hw->m[1].rcmd;
    state->rfmt = qmi_hw->m[1].rfmt;
}

static ram_func void flash_rp2350_restore_qmi_cs1(const flash_rp2350_qmi_save_state_t *state) {
    if (flash_devinfo_get_cs_size(1) == 0) {
        // No CS1 device: restore saved config (still valid, nothing touched it).
        qmi_hw->m[1].timing = state->timing;
        qmi_hw->m[1].rcmd = state->rcmd;
        qmi_hw->m[1].rfmt = state->rfmt;
    } else {
        // CS1 has QSPI RAM: our direct-mode ops returned it to serial command state.
        // Old XIP write mode is no longer valid; restore default 02h serial write config.
        qmi_hw->m[1].wfmt = QMI_M1_WFMT_RESET;
        qmi_hw->m[1].wcmd = QMI_M1_WCMD_RESET;
    }
}

#define PADS_QSPI_BASE 0x40040000

typedef struct {
    io_rw_32 voltage_select;
    io_rw_32 io[6];
} pads_qspi_hw_t;

#define pads_qspi_hw ((pads_qspi_hw_t *)PADS_QSPI_BASE)

typedef struct flash_hardware_save_state {
    flash_rp2350_qmi_save_state_t qmi_save;
    uint32_t qspi_pads[6];
} flash_hardware_save_state_t;

static ram_func void flash_save_hardware_state(flash_hardware_save_state_t *state) {
    for (size_t i = 0; i < 6; ++i) {
        state->qspi_pads[i] = pads_qspi_hw->io[i];
    }
    flash_rp2350_save_qmi_cs1(&state->qmi_save);
}

static ram_func void flash_restore_hardware_state(flash_hardware_save_state_t *state) {
    for (size_t i = 0; i < 6; ++i) {
        pads_qspi_hw->io[i] = state->qspi_pads[i];
    }
    flash_rp2350_restore_qmi_cs1(&state->qmi_save);
}

// Invalidate the XIP cache via the maintenance window (0x18000000).
// ROM flush_cache (FC) hangs on this board; this replaces it.
static ram_func void flash_invalidate_xip_cache(void) {
    for (uint32_t i = 0; i < 16u * 1024u; i += 8u) {
        *(volatile uint32_t *)(0x18000000u + i) = 0u;
    }
    __asm volatile ("dsb" : : : "memory");
    __asm volatile ("isb" : : : "memory");
}

// Send a single-byte command with no address/data. EN_BITS must be set.
static ram_func void flash_spi_cmd(uint8_t cmd) {
    hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
    qmi_hw->direct_tx = cmd;
    while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
    hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
}

// Poll STATUS1 WIP bit until clear. EN_BITS must be set.
static ram_func void flash_spi_wait_ready(void) {
    uint8_t sr;
    do {
        hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
        qmi_hw->direct_tx = 0x05; // RDSR1
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        qmi_hw->direct_tx = 0x00;
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS);
        sr = (uint8_t)qmi_hw->direct_rx;
        hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
    } while (sr & 1u);
}

// Exit QPI/continuous-read mode on flash and switch to single-SPI direct mode.
// The RP2350 bootrom may leave flash in QPI or continuous-read mode. This must be
// called before any direct SPI command. Leaves EN_BITS set on return — caller must
// not clear it until flash_restore_xip_spi_mode() reconfigures QMI M0 for SPI XIP.
static ram_func void flash_exit_qpi_mode(void) {
    // Enable direct mode (suspends QMI XIP auto-mode).
    hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_EN_BITS);
    // OD=1/PUE=1 on WP# (io[3]=SD2) and HOLD# (io[4]=SD3): disable output driver,
    // pull-up keeps them HIGH. QMI may have left these driven LOW from a prior quad
    // XIP transfer; LOW on HOLD# while CS is asserted pauses the flash chip.
    pads_qspi_hw->io[3] = (pads_qspi_hw->io[3] & ~0x4) | 0x8 | 0x80;
    pads_qspi_hw->io[4] = (pads_qspi_hw->io[4] & ~0x4) | 0x8 | 0x80;
    for (volatile int i = 0; i < 200; i++);

    uint32_t quad_oe = (0x2u << 16) | (1u << 19); // IWIDTH=quad, OE=1

    // Transaction 1: CS → 0xFF on 4 lines → ~CS.
    // Exits QPI mode if flash is in QPI mode (Winbond Exit QPI = 0xFF on all 4 lines).
    // If flash is in continuous read mode, this is received as partial address; CS
    // deassert aborts the transaction but does NOT exit continuous read mode.
    hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
    qmi_hw->direct_tx = 0xFF | quad_oe;
    while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
    hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
    for (volatile int i = 0; i < 200; i++);

    // Transaction 2: CS → 3 addr bytes + mode byte on 4 lines → ~CS.
    // If flash is in continuous read mode (quad I/O / 0xEB performance enhance):
    //   flash expects [24-bit addr][8-bit mode] on 4 lines after CS assert.
    //   Mode byte M[7:4] = 0xF (≠ 0xA continuation pattern) → exits continuous read.
    // If flash was in QPI and exited via txn 1: now in SPI mode; this txn is harmless
    //   (D0 carries 0x00 command = NOP, WEL=0 so no accidental write).
    hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
    qmi_hw->direct_tx = 0x00 | quad_oe; // addr[23:16]
    while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
    qmi_hw->direct_tx = 0x00 | quad_oe; // addr[15:8]
    while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
    qmi_hw->direct_tx = 0x00 | quad_oe; // addr[7:0]
    while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
    qmi_hw->direct_tx = 0xF0 | quad_oe; // mode: M[7:4]=0xF ≠ 0xA → exit continuous read
    while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
    hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);

    for (volatile int i = 0; i < 500; i++); // let flash settle
    // EN_BITS stays set: do not clear until flash_restore_xip_spi_mode() reconfigures
    // QMI M0, or QMI will resume XIP in quad mode while flash is in SPI mode.
}

void ram_func flash_range_write(uint32_t offset, const uint8_t *data, size_t count) {
    flash_hardware_save_state_t state;
    flash_save_hardware_state(&state);
    uint32_t ints_saved;
    __asm volatile("mrs %0, PRIMASK" : "=r"(ints_saved) : : "memory");
    __asm volatile("cpsid i" : : : "memory");
    __compiler_memory_barrier();
    flash_exit_qpi_mode();
    // Direct SPI page program (0x02), 256 bytes per page.
    for (uint32_t pos = 0; pos < (uint32_t)count; pos += FLASH_PAGE_SIZE) {
        flash_spi_cmd(0x06); // WREN
        uint32_t addr = offset + pos;
        hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
        qmi_hw->direct_tx = 0x02; // Page Program
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        qmi_hw->direct_tx = (addr >> 16) & 0xFFu;
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        qmi_hw->direct_tx = (addr >> 8) & 0xFFu;
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        qmi_hw->direct_tx = addr & 0xFFu;
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        uint32_t page_count = FLASH_PAGE_SIZE;
        if (pos + FLASH_PAGE_SIZE > (uint32_t)count) page_count = (uint32_t)count - pos;
        for (uint32_t j = 0; j < page_count; j++) {
            while (qmi_hw->direct_csr & QMI_DIRECT_CSR_TXFULL_BITS);
            qmi_hw->direct_tx = data[pos + j];
            while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        }
        hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
        flash_spi_wait_ready();
    }
    flash_restore_xip_spi_mode();
    flash_invalidate_xip_cache();
    flash_restore_hardware_state(&state);
    __asm volatile("msr PRIMASK, %0" : : "r"(ints_saved) : "memory");
}

void ram_func flash_erase_blocks(uint32_t offset, size_t count) {
    flash_hardware_save_state_t state;
    flash_save_hardware_state(&state);
    uint32_t ints_saved;
    __asm volatile("mrs %0, PRIMASK" : "=r"(ints_saved) : : "memory");
    __asm volatile("cpsid i" : : : "memory");
    __compiler_memory_barrier();
    flash_exit_qpi_mode();
    // Direct SPI 64KB block erase (0xD8).
    for (uint32_t pos = 0; pos < (uint32_t)count; pos += FLASH_BLOCK_SIZE) {
        flash_spi_cmd(0x06); // WREN
        uint32_t addr = offset + pos;
        hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
        qmi_hw->direct_tx = FLASH_BLOCK_ERASE_CMD; // 0xD8
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        qmi_hw->direct_tx = (addr >> 16) & 0xFFu;
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        qmi_hw->direct_tx = (addr >> 8) & 0xFFu;
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        qmi_hw->direct_tx = addr & 0xFFu;
        while (qmi_hw->direct_csr & QMI_DIRECT_CSR_RXEMPTY_BITS); (void)qmi_hw->direct_rx;
        hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
        flash_spi_wait_ready(); // W25Q128JV 64KB erase: typ 150ms, max 2000ms
    }
    flash_restore_xip_spi_mode();
    flash_invalidate_xip_cache();
    flash_restore_hardware_state(&state);
    __asm volatile("msr PRIMASK, %0" : : "r"(ints_saved) : "memory");
}

void ram_func flash_do_cmd(const uint8_t *txbuf, uint8_t *rxbuf, size_t count) {
    flash_hardware_save_state_t state;
    flash_save_hardware_state(&state);
    uint32_t ints_saved;
    __asm volatile("mrs %0, PRIMASK" : "=r"(ints_saved) : : "memory");
    __asm volatile("cpsid i" : : : "memory");
    __compiler_memory_barrier();
    flash_exit_qpi_mode();

    // Assert CS and run the FIFO transfer in single-SPI direct mode.
    hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
    size_t tx_remaining = count;
    size_t rx_remaining = count;

    // QMI stalls on full DIRECT_RX, so no need to bound FIFO contents.
    while (tx_remaining || rx_remaining) {
        uint32_t flags = qmi_hw->direct_csr;
        bool can_put = !(flags & QMI_DIRECT_CSR_TXFULL_BITS);
        bool can_get = !(flags & QMI_DIRECT_CSR_RXEMPTY_BITS);
        if (can_put && tx_remaining) {
            qmi_hw->direct_tx = *txbuf++;
            --tx_remaining;
        }
        if (can_get && rx_remaining) {
            *rxbuf++ = (uint8_t)qmi_hw->direct_rx;
            --rx_remaining;
        }
    }
    hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_ASSERT_CS0N_BITS);
    flash_restore_xip_spi_mode();
    flash_invalidate_xip_cache();
    flash_restore_hardware_state(&state);
    __asm volatile("msr PRIMASK, %0" : : "r"(ints_saved) : "memory");
}

*/
import "C"

func enterBootloader() {
	C.reset_usb_boot(0, 0)
}

func doFlashCommand(tx []byte, rx []byte) error {
	if len(tx) != len(rx) {
		return errFlashInvalidWriteLength
	}

	C.flash_do_cmd(
		(*C.uint8_t)(unsafe.Pointer(&tx[0])),
		(*C.uint8_t)(unsafe.Pointer(&rx[0])),
		C.ulong(len(tx)))

	return nil
}

// Flash related code
const memoryStart = C.XIP_BASE // memory start for purpose of erase

func (f flashBlockDevice) writeAt(p []byte, off int64) (n int, err error) {
	if writeAddress(off)+uintptr(C.XIP_BASE) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	state := interrupt.Disable()
	defer interrupt.Restore(state)

	// rp2350 writes to offset, not actual address
	// e.g. real address 0x10003000 is written to at
	// 0x00003000
	address := writeAddress(off)
	padded := flashPad(p, int(f.WriteBlockSize()))

	C.flash_range_write(C.uint32_t(address),
		(*C.uint8_t)(unsafe.Pointer(&padded[0])),
		C.ulong(len(padded)))

	return len(padded), nil
}

func (f flashBlockDevice) eraseBlocks(start, length int64) error {
	address := writeAddress(start * f.EraseBlockSize())
	if address+uintptr(C.XIP_BASE) > FlashDataEnd() {
		return errFlashCannotErasePastEOF
	}

	state := interrupt.Disable()
	defer interrupt.Restore(state)

	C.flash_erase_blocks(C.uint32_t(address), C.ulong(length*f.EraseBlockSize()))

	return nil
}
