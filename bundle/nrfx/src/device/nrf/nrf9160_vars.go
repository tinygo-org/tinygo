//go:build nrf && nrf9160
// +build nrf,nrf9160

package nrf

// TODO: delete this file once drivers are refactored

const (
	IRQ_GPIOTE = IRQ_GPIOTE0
)

var (
	GPIOTE = GPIOTE0_S

	SAADC = SAADC_S
)

const (
	TWI_FREQUENCY_FREQUENCY_K100 = TWIM_FREQUENCY_FREQUENCY_K100
	TWI_FREQUENCY_FREQUENCY_K400 = TWIM_FREQUENCY_FREQUENCY_K400
	TWI_ENABLE_ENABLE_Disabled   = TWIM_ENABLE_ENABLE_Disabled
	TWI_ENABLE_ENABLE_Enabled    = TWIM_ENABLE_ENABLE_Enabled
)
