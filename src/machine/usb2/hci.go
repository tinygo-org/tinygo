package usb2

type hci interface {
	init() status
	enable(enable bool) status
	critical(enter bool) status
	interrupt()
	udelay(micros uint32)
}
