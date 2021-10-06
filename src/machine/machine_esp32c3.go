// +build esp32c3

package machine

import (
	"device/esp"
	"runtime/volatile"
)

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

type Module int

const (
	// Different modules of ESP32C3 have different clock, interrupt and configuration registers
	// Each module defined here and helper functions to provide uniform access to those registers.
	PERIPH_LEDC_MODULE Module = iota
	PERIPH_UART0_MODULE
	PERIPH_UART1_MODULE
	PERIPH_USB_DEVICE_MODULE
	PERIPH_I2C0_MODULE
	PERIPH_I2S1_MODULE
	PERIPH_TIMG0_MODULE
	PERIPH_TIMG1_MODULE
	PERIPH_UHCI0_MODULE
	PERIPH_RMT_MODULE
	PERIPH_SPI_MODULE  //SPI1
	PERIPH_SPI2_MODULE //SPI2
	PERIPH_TWAI_MODULE
	PERIPH_RNG_MODULE
	PERIPH_WIFI_MODULE
	PERIPH_BT_MODULE
	PERIPH_WIFI_BT_COMMON_MODULE
	PERIPH_BT_BASEBAND_MODULE
	PERIPH_BT_LC_MODULE
	PERIPH_RSA_MODULE
	PERIPH_AES_MODULE
	PERIPH_SHA_MODULE
	PERIPH_HMAC_MODULE
	PERIPH_DS_MODULE
	PERIPH_GDMA_MODULE
	PERIPH_SYSTIMER_MODULE
	PERIPH_SARADC_MODULE
	PERIPH_GPIO_MODULE
)

func (module Module) interruptForModule() uint32 {
	switch module {
	case PERIPH_GPIO_MODULE:
		return 6
	case PERIPH_UART0_MODULE, PERIPH_UART1_MODULE:
		return 7
	}
	return 0
}

func (module Module) mapRegisterForModule() *volatile.Register32 {
	switch module {
	case PERIPH_GPIO_MODULE:
		return &esp.INTERRUPT_CORE0.GPIO_INTERRUPT_PRO_MAP
	case PERIPH_UART0_MODULE:
		return &esp.INTERRUPT_CORE0.UART_INTR_MAP
	case PERIPH_UART1_MODULE:
		return &esp.INTERRUPT_CORE0.UART1_INTR_MAP
	}
	return nil
}

func (module Module) getClockEnableRegister() *volatile.Register32 {
	switch module {
	case PERIPH_RNG_MODULE, PERIPH_WIFI_MODULE, PERIPH_BT_MODULE, PERIPH_WIFI_BT_COMMON_MODULE, PERIPH_BT_BASEBAND_MODULE, PERIPH_BT_LC_MODULE:
		return &esp.APB_CTRL.WIFI_CLK_EN
	case PERIPH_HMAC_MODULE, PERIPH_DS_MODULE, PERIPH_AES_MODULE, PERIPH_RSA_MODULE, PERIPH_SHA_MODULE, PERIPH_GDMA_MODULE:
		return &esp.SYSTEM.PERIP_CLK_EN1
	default:
		return &esp.SYSTEM.PERIP_CLK_EN0
	}
}

func (module Module) getClockResetRegister() *volatile.Register32 {
	switch module {
	case PERIPH_RNG_MODULE, PERIPH_WIFI_MODULE, PERIPH_BT_MODULE, PERIPH_WIFI_BT_COMMON_MODULE, PERIPH_BT_BASEBAND_MODULE, PERIPH_BT_LC_MODULE:
		return &esp.APB_CTRL.WIFI_RST_EN
	case PERIPH_HMAC_MODULE, PERIPH_DS_MODULE, PERIPH_AES_MODULE, PERIPH_RSA_MODULE, PERIPH_SHA_MODULE, PERIPH_GDMA_MODULE:
		return &esp.SYSTEM.PERIP_RST_EN1
	default:
		return &esp.SYSTEM.PERIP_RST_EN0
	}
}

func (module Module) getClockBit() uint32 {
	switch module {
	case PERIPH_UART0_MODULE:
		return esp.SYSTEM_PERIP_CLK_EN0_UART_CLK_EN
	case PERIPH_UART1_MODULE:
		return esp.SYSTEM_PERIP_CLK_EN0_UART1_CLK_EN
	default:
		return 0
	}
}
