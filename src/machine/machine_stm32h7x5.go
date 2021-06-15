// +build stm32h7x5

package machine

import (
	"device/stm32"
	"runtime/volatile"
	"unsafe"
)

// EnableClock enables or disables the clock gate for the given peripheral.
func EnableClock(bus unsafe.Pointer, enable bool) bool {
	var reg *volatile.Register32
	var msk uint32
	switch bus {
	case unsafe.Pointer(stm32.HSEM):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_HSEMEN
	case unsafe.Pointer(stm32.ART):
		reg, msk = &stm32.RCC.AHB1ENR, stm32.RCC_AHB1ENR_ARTEN
	case unsafe.Pointer(stm32.SYSCFG):
		reg, msk = &stm32.RCC.APB4ENR, stm32.RCC_APB4ENR_SYSCFGEN
	case unsafe.Pointer(stm32.GPIOA):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOAEN
	case unsafe.Pointer(stm32.GPIOB):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOBEN
	case unsafe.Pointer(stm32.GPIOC):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOCEN
	case unsafe.Pointer(stm32.GPIOD):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIODEN
	case unsafe.Pointer(stm32.GPIOE):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOEEN
	case unsafe.Pointer(stm32.GPIOF):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOFEN
	case unsafe.Pointer(stm32.GPIOG):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOGEN
	case unsafe.Pointer(stm32.GPIOH):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOHEN
	case unsafe.Pointer(stm32.GPIOI):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOIEN
	case unsafe.Pointer(stm32.GPIOJ):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOJEN
	case unsafe.Pointer(stm32.GPIOK):
		reg, msk = &stm32.RCC.AHB4ENR, stm32.RCC_AHB4ENR_GPIOKEN
	}
	if nil != reg {
		for !SemRCC.Lock(CoreID) {
		} // wait until we have exclusive access to RCC
		if enable {
			reg.SetBits(msk)
		} else {
			reg.ClearBits(msk)
		}
		SemRCC.Unlock(CoreID)
		// msk must be 1-bit for this to be correct (always true, I think?)
		return enable == reg.HasBits(msk)
	}
	return false // unsupported peripheral
}
