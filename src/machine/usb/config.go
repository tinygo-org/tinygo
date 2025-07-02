package usb

type EndpointConfig struct {
	Index          uint8
	IsIn           bool
	TxHandler      func()
	RxHandler      func([]byte)
	DelayRxHandler func([]byte) bool
	StallHandler   func(Setup) bool
	Type           uint8
}

type SetupConfig struct {
	Index   uint8
	Handler func(Setup) bool
}
