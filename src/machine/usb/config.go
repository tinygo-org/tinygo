package usb

type EndpointConfig struct {
	No        uint8
	IsIn      bool
	TxHandler func()
	RxHandler func([]byte)
	Type      uint8
}

type SetupConfig struct {
	No      uint8
	Handler func(Setup) bool
}
