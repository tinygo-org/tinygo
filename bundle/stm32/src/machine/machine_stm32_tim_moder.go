//go:build stm32 && !stm32f1

package machine

// Configuration of a GPIO pin for PWM output for STM32 MCUs with MODER
// register (most MCUs except STM32F1 series).

func (t *TIM) configurePin(channel uint8, pf PinFunction) {
	pf.Pin.ConfigureAltFunc(PinConfig{Mode: PinModePWMOutput}, pf.AltFunc)
}
