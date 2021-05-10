// +build stm32h7x7

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
		if enable {
			reg.SetBits(msk)
		} else {
			reg.ClearBits(msk)
		}
		return enable == reg.HasBits(msk)
	}
	return false // unsupported peripheral
}

// enableGPIO enables the bus clock necessary for peripheral read/write access
// for each GPIO port. If a clock fails to enable, returns false immediately and
// all remaining GPIO clocks are unaffected.
// func enableGPIO() bool {
// 	for _, g := range []*stm32.GPIO_Type{
// 		stm32.GPIOA, stm32.GPIOB, stm32.GPIOC, stm32.GPIOD,
// 		stm32.GPIOE, stm32.GPIOF, stm32.GPIOG, stm32.GPIOH,
// 		stm32.GPIOI, stm32.GPIOJ, stm32.GPIOK,
// 	} {
// 		if !enable(unsafe.Pointer(g), true) {
// 			return false
// 		}
// 	}
// 	return true
// }

// Allocate enables the receiver CPU core's access to the given peripheral.
// By default, some peripherals are only accessible from a single core, and the
// other core must explicitly request read/write access. Official ST reference
// manuals refer to this as "allocation".
func Allocate(bus unsafe.Pointer, core *stm32.RCC_CORE_Type) bool {
	var reg *volatile.Register32
	var msk uint32
	switch bus {
	case unsafe.Pointer(stm32.FLASH):
		switch core {
		case stm32.RCC_CORE1:
			// FLASH is implicitly allocated to core 1 (M7)
			return true
		case stm32.RCC_CORE2:
			reg, msk = &stm32.RCC.AHB3ENR, stm32.RCC_AHB3ENR_FLASHEN
		}
	}
	if nil != reg {
		reg.SetBits(msk)
		return reg.HasBits(msk)
	}
	return false // reripheral not supported for the given core
}
