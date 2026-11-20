//go:build (sam || nrf52840 || rp2040 || rp2350 || stm32f4 || stm32f7 || stm32h7 || mimxrt1062 || (esp32s3 && usbotg) || (ra4m1 && serial.usb)) && boardpackage

package machine

func boardUSBIdentity() (uint16, uint16, string, string)

func defaultUSBIdentity() (vid, pid uint16, manufacturer, product string) {
	return boardUSBIdentity()
}
