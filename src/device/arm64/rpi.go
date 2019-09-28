// +build rpi

package arm64

import "unsafe"

//
// GPIO
//
var GPFSEL0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200000))
var GPFSEL1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200004))
var GPFSEL2 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200008))
var GPFSEL3 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0020000C))
var GPFSEL4 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200010))
var GPFSEL5 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200014))
var GPSET0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0020001C))
var GPSET1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200020))
var GPCLR0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200028))
var GPLEV0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200034))
var GPLEV1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200038))
var GPEDS0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200040))
var GPEDS1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200044))
var GPHEN0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200064))
var GPHEN1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200068))
var GPPUD = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200094))
var GPPUDCLK0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00200098))
var GPPUDCLK1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0020009C))

//
// MINI UART
//
var AUX_ENABLE = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215004))
var AUX_MU_IO = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215040))
var AUX_MU_IER = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215044))
var AUX_MU_IIR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215048))
var AUX_MU_LCR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0021504C))
var AUX_MU_MCR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215050))
var AUX_MU_LSR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215054))
var AUX_MU_MSR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215058))
var AUX_MU_SCRATCH = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0021505C))
var AUX_MU_CNTL = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215060))
var AUX_MU_STAT = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215064))
var AUX_MU_BAUD = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00215068))

//
// PL011 UART registers
//
var UART0_DR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00201000))
var UART0_FR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00201018))
var UART0_IBRD = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00201024))
var UART0_FBRD = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00201028))
var UART0_LCRH = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0020102C))
var UART0_CR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00201030))
var UART0_IMSC = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00201038))
var UART0_ICR = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00201044))

//
// Mailboxes
//

const MBOX_REQUEST = 0

//
// channels
//
const MBOX_CH_POWER = 0
const MBOX_CH_FB = 1
const MBOX_CH_VUART = 2
const MBOX_CH_VCHIQ = 3
const MBOX_CH_LEDS = 4
const MBOX_CH_BTNS = 5
const MBOX_CH_TOUCH = 6
const MBOX_CH_COUNT = 7
const MBOX_CH_PROP = 8

//
// MBOX tags
//
const MBOX_TAG_GETSERIAL = 0x10004
const MBOX_TAG_SETCLKRATE = 0x38002
const MBOX_TAG_SET_GPIO_STATE = 0x00038041
const MBOX_TAG_GET_MAX_CLKRATE = 0x00030004

const MBOX_TAG_LAST = 0

//
// MBOX commands
//
var VIDEOCORE_MBOX = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B880))
var MBOX_READ = unsafe.Pointer(uintptr(VIDEOCORE_MBOX) + uintptr(0x0))
var MBOX_POLL = unsafe.Pointer(uintptr(VIDEOCORE_MBOX) + uintptr(0x10))
var MBOX_SENDER = unsafe.Pointer(uintptr(VIDEOCORE_MBOX) + uintptr(0x14))
var MBOX_STATUS = unsafe.Pointer(uintptr(VIDEOCORE_MBOX) + uintptr(0x18))
var MBOX_CONFIG = unsafe.Pointer(uintptr(VIDEOCORE_MBOX) + uintptr(0x1C))
var MBOX_WRITE = unsafe.Pointer(uintptr(VIDEOCORE_MBOX) + uintptr(0x20))

//
// MBOX data
//
const MBOX_FULL = 0x80000000
const MBOX_RESPONSE = 0x80000000
const MBOX_EMPTY = 0x40000000

//
// system timer (64 bit clock,starts at boot)
//

var SYSTMR_LO = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003004))
var SYSTMR_HI = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003008))

//
// System Control Register
//

var SCTLR_RESERVED uint32 = (3 << 28) | (3 << 22) | (1 << 20) | (1 << 11)
var SCTLR_EE_LITTLE_ENDIAN uint32 = (0 << 25)
var SCTLR_EOE_LITTLE_ENDIAN uint32 = (0 << 24)
var SCTLR_I_CACHE_DISABLED uint32 = (0 << 12)
var SCTLR_D_CACHE_DISABLED uint32 = (0 << 2)
var SCTLR_MMU_DISABLED uint32 = (0 << 0)
var SCTLR_MMU_ENABLED uint32 = (1 << 0)
var SCTLR_VALUE_MMU_DISABLED uint32 = (SCTLR_RESERVED | SCTLR_EE_LITTLE_ENDIAN | SCTLR_I_CACHE_DISABLED | SCTLR_D_CACHE_DISABLED | SCTLR_MMU_DISABLED)

//
// size of all saved registers
//
const S_FRAME_SIZE = 256

//
// Exceptions (interrupts) for all 4 levels (4 types of exceptions, 4 levels of execution)
//
const SYNC_INVALID_EL1t = 0
const IRQ_INVALID_EL1t = 1
const FIQ_INVALID_EL1t = 2
const ERROR_INVALID_EL1t = 3

const SYNC_INVALID_EL1h = 4
const IRQ_INVALID_EL1h = 5
const FIQ_INVALID_EL1h = 6
const ERROR_INVALID_EL1h = 7

const SYNC_INVALID_EL0_64 = 8
const IRQ_INVALID_EL0_64 = 9
const FIQ_INVALID_EL0_64 = 10
const ERROR_INVALID_EL0_64 = 11

const SYNC_INVALID_EL0_32 = 12
const IRQ_INVALID_EL0_32 = 13
const FIQ_INVALID_EL0_32 = 14
const ERROR_INVALID_EL0_32 = 15

//
// Interrupt Request (IRQ)
//

var IRQ_BASIC_PENDING = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B200))
var IRQ_PENDING_1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B204))
var IRQ_PENDING_2 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B208))
var FIQ_CONTROL = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B20C))
var ENABLE_IRQS_1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B210))
var ENABLE_IRQS_2 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B214))
var ENABLE_BASIC_IRQS = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B218))
var DISABLE_IRQS_1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B21C))
var DISABLE_IRQS_2 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B220))
var DISABLE_BASIC_IRQS = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000B224))

const SYSTEM_TIMER_IRQ_0 = (1 << 0)
const SYSTEM_TIMER_IRQ_1 = (1 << 1)
const SYSTEM_TIMER_IRQ_2 = (1 << 2)
const SYSTEM_TIMER_IRQ_3 = (1 << 3)

//
// Timer
//

var TIMER_CS = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003000))
var TIMER_CLO = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003004))
var TIMER_CHI = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003008))
var TIMER_C0 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x0000300C))
var TIMER_C1 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003010))
var TIMER_C2 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003014))
var TIMER_C3 = unsafe.Pointer(uintptr(MMIO_BASE) + uintptr(0x00003018))

const TIMER_CS_M0 = (1 << 0)
const TIMER_CS_M1 = (1 << 1)
const TIMER_CS_M2 = (1 << 2)
const TIMER_CS_M3 = (1 << 3)

//
// These timer controls are probably not on the RPI3 hardware, just QEMU
// Note that these are at 0x40000000 not MMIO_BASE
//
var CORE0_TIMER_IRQCNTL = unsafe.Pointer(uintptr(0x40000000) + uintptr(0x40))
var CORE0_IRQ_SOURCE = unsafe.Pointer(uintptr(0x40000000) + uintptr(0x60))
