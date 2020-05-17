// +build !stm32f103xx,!atsamd51,!nrf52,!nrf52840,!stm32f407,!mk66f18

package arm

// GetFaultStatus is a stub because the Cortex-M0 does not have fault status
// registers.
func GetFaultStatus() FaultStatus { return 0 }

// Address is a stub because the Cortex-M0 does not have fault status registers.
func (fs MemFaultStatus) Address() (uintptr, bool) { return 0, false }

// Address is a stub because the Cortex-M0 does not have fault status registers.
func (fs BusFaultStatus) Address() (uintptr, bool) { return 0, false }
