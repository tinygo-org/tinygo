//go:build stm32 && !stm32h7

package machine

// EnterBootloader resets the chip into the bootloader.
// This is currently a stub for STM32, required to satisfy machine.EnterBootloader
// called by machine/usb/cdc.
func EnterBootloader() {
}
