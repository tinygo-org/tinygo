// +build !avr,!nrf,!sam,!stm32

package machine

// Dummy machine package, filled with no-ops.

type GPIOMode uint8

const (
	GPIO_INPUT = iota
	GPIO_OUTPUT
)

// Fake LED numbers, for testing.
var (
	LED  uint8 = LED1
	LED1 uint8 = 0
	LED2 uint8 = 0
	LED3 uint8 = 0
	LED4 uint8 = 0
)

// Fake button numbers, for testing.
var (
	BUTTON  uint8 = BUTTON1
	BUTTON1 uint8 = 5
	BUTTON2 uint8 = 6
	BUTTON3 uint8 = 7
	BUTTON4 uint8 = 8
)

// Fake SPI interfaces, for testing.
var (
	SPI0 = SPI{0}
)

var (
	GPIOConfigure func(pin uint8, config GPIOConfig)
	GPIOSet       func(pin uint8, value bool)
	GPIOGet       func(pin uint8) bool
	SPIConfigure  func(bus uint8, sck uint8, mosi uint8, miso uint8)
	SPITransfer   func(bus uint8, w uint8) uint8
)

func (p GPIO) Configure(config GPIOConfig) {
	if GPIOConfigure != nil {
		GPIOConfigure(p.Pin, config)
	}
}

func (p GPIO) Set(value bool) {
	if GPIOSet != nil {
		GPIOSet(p.Pin, value)
	}
}

func (p GPIO) Get() bool {
	if GPIOGet != nil {
		return GPIOGet(p.Pin)
	}
	return false
}

type SPI struct {
	Bus uint8
}

type SPIConfig struct {
	Frequency uint32
	SCK       uint8
	MOSI      uint8
	MISO      uint8
	Mode      uint8
}

func (spi SPI) Configure(config SPIConfig) {
	if SPIConfigure != nil {
		SPIConfigure(spi.Bus, config.SCK, config.MOSI, config.MISO)
	}
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	if SPITransfer != nil {
		return SPITransfer(spi.Bus, w), nil
	}
	return 0, nil
}
