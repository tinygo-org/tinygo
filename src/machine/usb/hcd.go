package usb

type hcd interface {
	class() class
	init() status
	enable(enable bool) status
	critical(enter bool) status
	interrupt()
	udelay(micros uint32)
}
