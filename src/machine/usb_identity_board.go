//go:build (sam || nrf52840 || rp2040 || rp2350 || stm32f4 || stm32f7 || stm32h7 || mimxrt1062 || (esp32s3 && usbotg) || (ra4m1 && serial.usb)) && !boardpackage

package machine

func defaultUSBIdentity() (vid, pid uint16, manufacturer, product string) {
	return usb_VID, usb_PID, usb_STRING_MANUFACTURER, usb_STRING_PRODUCT
}
