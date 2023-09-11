//go:build rp2040

package machine

import (
	"device/arm"
	"device/rp"
)

// https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/pico_fix/rp2040_usb_device_enumeration/rp2040_usb_device_enumeration.c
// According to errata RP2040-E5:
// "It is safe (and inexpensive) to enable the software workaround even when using versions of RP2040
// which include the fix in hardware."
// So let us always use the software fix.
func fixRP2040UsbDeviceEnumeration() {
	// After coming out of reset, the hardware expects 800us of LS_J (linestate J) time
	// before it will move to the connected state. However on a hub that broadcasts packets
	// for other devices this isn't the case. The plan here is to wait for the end of the bus
	// reset, force an LS_J for 1ms and then switch control back to the USB phy. Unfortunately
	// this requires us to use GPIO15 as there is no other way to force the input path.
	// We only need to force DP as DM can be left at zero. It will be gated off by GPIO
	// logic if it isn't func selected.

	// Wait SE0 phase will call force ls_j phase which will call finish phase
	hw_enumeration_fix_wait_se0()
}

func hw_enumeration_fix_wait_se0() {
	// Wait for SE0 to end (i.e. the host to stop resetting). This reset can last quite long.
	// 10-15ms so we are going to set a timer callback.

	// if timer pool disabled, or no timer available, have to busy wait.
	hw_enumeration_fix_busy_wait_se0()
}

const (
	LS_SE0 = 0b00
	LS_J   = 0b01
	LS_K   = 0b10
	LS_SE1 = 0b11
)

const (
	dp = 15
)

var (
	gpioCtrlPrev uint32
	padCtrlPrev  uint32
)

func hw_enumeration_fix_busy_wait_se0() {
	for ((rp.USBCTRL_REGS.SIE_STATUS.Get() & rp.USBCTRL_REGS_SIE_STATUS_LINE_STATE_Msk) >> rp.USBCTRL_REGS_SIE_STATUS_LINE_STATE_Pos) == LS_SE0 {
	}

	// Now force LS_J (next stage of fix)
	hw_enumeration_fix_force_ls_j()
}

func hw_enumeration_fix_force_ls_j() {
	// DM must be 0 for this to work. This is true if it is selected
	// to any other function. fn 8 on this pin is only for debug so shouldn't
	// be selected

	// Before changing any pin state, take a copy of the current gpio control register
	gpioCtrlPrev = ioBank0.io[dp].ctrl.Get()
	// Also take a copy of the pads register
	padCtrlPrev = padsBank0.io[dp].Get()

	// Enable bus keep and force pin to tristate, so USB DP muxing doesn't affect
	// pin state
	padsBank0.io[dp].SetBits(rp.PADS_BANK0_GPIO0_PUE | rp.PADS_BANK0_GPIO0_PDE)
	ioBank0.io[dp].ctrl.ReplaceBits(rp.IO_BANK0_GPIO0_CTRL_OEOVER_DISABLE, rp.IO_BANK0_GPIO0_CTRL_OEOVER_Msk>>rp.IO_BANK0_GPIO0_CTRL_OEOVER_Pos, rp.IO_BANK0_GPIO0_CTRL_OEOVER_Pos)

	// Select function 8 (USB debug muxing) without disturbing other controls
	ioBank0.io[dp].ctrl.ReplaceBits(8, rp.IO_BANK0_GPIO0_CTRL_FUNCSEL_Msk>>rp.IO_BANK0_GPIO0_CTRL_FUNCSEL_Pos, rp.IO_BANK0_GPIO0_CTRL_FUNCSEL_Pos)

	// J state is a differential 1 for a full speed device so
	// DP = 1 and DM = 0. Don't actually need to set DM low as it
	// is already gated assuming it isn't funcseld.
	ioBank0.io[dp].ctrl.ReplaceBits(rp.IO_BANK0_GPIO1_CTRL_INOVER_HIGH, rp.IO_BANK0_GPIO1_CTRL_INOVER_Msk>>rp.IO_BANK0_GPIO1_CTRL_INOVER_Pos, rp.IO_BANK0_GPIO1_CTRL_INOVER_Pos)

	// Force PHY pull up to stay before switching away from the phy
	rp.USBCTRL_REGS.USBPHY_DIRECT.SetBits(rp.USBCTRL_REGS_USBPHY_DIRECT_DP_PULLUP_EN)
	rp.USBCTRL_REGS.USBPHY_DIRECT_OVERRIDE.SetBits(rp.USBCTRL_REGS_USBPHY_DIRECT_OVERRIDE_DP_PULLUP_EN_OVERRIDE_EN)

	// Switch to GPIO phy with LS_J forced
	rp.USBCTRL_REGS.USB_MUXING.Set(rp.USBCTRL_REGS_USB_MUXING_TO_DIGITAL_PAD | rp.USBCTRL_REGS_USB_MUXING_SOFTCON)

	// LS_J is now forced but while loop to wait ~800us here just to check
	waitCycles(25000)

	// if timer pool disabled, or no timer available, have to busy wait.
	hw_enumeration_fix_finish()

}

func hw_enumeration_fix_finish() {
	// Should think we are connected now
	for (rp.USBCTRL_REGS.SIE_STATUS.Get() & rp.USBCTRL_REGS_SIE_STATUS_CONNECTED) != rp.USBCTRL_REGS_SIE_STATUS_CONNECTED {
	}

	// Switch back to USB phy
	rp.USBCTRL_REGS.USB_MUXING.Set(rp.USBCTRL_REGS_USB_MUXING_TO_PHY | rp.USBCTRL_REGS_USB_MUXING_SOFTCON)

	// Get rid of DP pullup override
	rp.USBCTRL_REGS.USBPHY_DIRECT_OVERRIDE.ClearBits(rp.USBCTRL_REGS_USBPHY_DIRECT_OVERRIDE_DP_PULLUP_EN_OVERRIDE_EN)

	// Finally, restore the gpio ctrl value back to GPIO15
	ioBank0.io[dp].ctrl.Set(gpioCtrlPrev)
	// Restore the pad ctrl value
	padsBank0.io[dp].Set(padCtrlPrev)
}

func waitCycles(n int) {
	for n > 0 {
		arm.Asm("nop")
		n--
	}
}
