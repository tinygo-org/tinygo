package csw

type Status uint8

const (
	StatusPassed Status = iota
	StatusFailed
	StatusPhaseError
)

const (
	MsgLen    = 13
	Signature = 0x53425355 // "USBS" in little endian
)
