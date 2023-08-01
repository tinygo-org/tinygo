package usb

type EndpointConfig struct {
	Index     uint8
	IsIn      bool
	TxHandler func()
	RxHandler func([]byte)
	Type      uint8
}

type SetupConfig struct {
	Index   uint8
	Handler func(Setup) bool
}
