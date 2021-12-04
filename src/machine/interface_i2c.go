//go:build !gameboyadvance && !mimxrt1062 && !mk66f18 && !attiny85 && !stm32 && !esp32c3 && !k210
// +build !gameboyadvance,!mimxrt1062,!mk66f18,!attiny85,!stm32,!esp32c3,!k210

package machine

func init() {
	type i2cer interface {
		SetBaudRate(br uint32) error
		Configure(config I2CConfig) error
		ReadRegister(address uint8, register uint8, data []byte) error
		WriteRegister(address uint8, register uint8, data []byte) error

		Tx(addr uint16, w, r []byte) error
	}

	var _ i2cer = &I2C{}
}
