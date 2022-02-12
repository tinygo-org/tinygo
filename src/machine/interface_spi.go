//go:build !gameboyadvance && !mimxrt1062 && !mk66f18 && !stm32 && !attiny85 && !esp32c3 && !k210
// +build !gameboyadvance,!mimxrt1062,!mk66f18,!stm32,!attiny85,!esp32c3,!k210

package machine

func init() {
	type spier interface {
		SetBaudRate(br uint32) error
		Configure(config SPIConfig) error
		Transfer(b byte) (byte, error)
		Tx(w, r []byte) error
	}
	_ = SPIConfig{
		Frequency: 0,
	}
	var _ spier = &SPI{}
}
