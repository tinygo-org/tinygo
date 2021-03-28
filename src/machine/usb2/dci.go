package usb2

type dci interface {
	init() status
	enable(enable bool) status
	critical(enter bool) status
	interrupt()
	udelay(micros uint32)
}
