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
