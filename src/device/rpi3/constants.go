// +build rpi3

package rpi3

// derived from the execellent tutorial by bzt
// https://github.com/bztsrc/raspi3-tutorial/

import "unsafe"

var MMIO_BASE = unsafe.Pointer(uintptr(0x3F000000))

/*
 * GPIO
 */
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

/*
 * MINI UART
 */
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


/*
 * PL011 UART registers
 */
var UART0_DR =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x00201000))
var UART0_FR =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x00201018))
var UART0_IBRD =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x00201024))
var UART0_FBRD =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x00201028))
var UART0_LCRH =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x0020102C))
var UART0_CR =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x00201030))
var UART0_IMSC =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x00201038))
var UART0_ICR =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x00201044))

/*
 * Mailboxes
 */

const MBOX_REQUEST   =  0

/* channels */
const MBOX_CH_POWER  = 0
const MBOX_CH_FB     = 1
const MBOX_CH_VUART  = 2
const MBOX_CH_VCHIQ  = 3
const MBOX_CH_LEDS   = 4
const MBOX_CH_BTNS   = 5
const MBOX_CH_TOUCH  = 6
const MBOX_CH_COUNT  = 7
const MBOX_CH_PROP   = 8

/* tags */
const MBOX_TAG_GETSERIAL =     0x10004
const MBOX_TAG_SETCLKRATE =     0x38002
const MBOX_TAG_LAST=           0


/*
 * mbox
 */

var VIDEOCORE_MBOX =       unsafe.Pointer(uintptr(MMIO_BASE)+uintptr(0x0000B880))
var MBOX_READ =       unsafe.Pointer(uintptr(VIDEOCORE_MBOX)+uintptr(0x0))
var MBOX_POLL =       unsafe.Pointer(uintptr(VIDEOCORE_MBOX)+uintptr(0x10))
var MBOX_SENDER =       unsafe.Pointer(uintptr(VIDEOCORE_MBOX)+uintptr(0x14))
var MBOX_STATUS =       unsafe.Pointer(uintptr(VIDEOCORE_MBOX)+uintptr(0x18))
var MBOX_CONFIG =       unsafe.Pointer(uintptr(VIDEOCORE_MBOX)+uintptr(0x1C))
var MBOX_WRITE =       unsafe.Pointer(uintptr(VIDEOCORE_MBOX)+uintptr(0x20))
const MBOX_FULL =     0x80000000
const MBOX_RESPONSE = 0x80000000
const MBOX_EMPTY =    0x40000000
