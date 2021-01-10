package main

import (
	"device/stm32"
	"machine"
	"runtime/interrupt"
	"time"
)

var (
	colorFlag uint8
)

// Interrupt handler for GPIO interrupts
func gpios_int(inter interrupt.Interrupt) {
	irqStatus := stm32.EXTI.PR.Get()
	stm32.EXTI.PR.Set(irqStatus)

	if (irqStatus & 0x4000) > 0 { // PB14 : Button
		colorFlag = colorFlag << 1
		if colorFlag == 8 {
			colorFlag = 1
		}

		println("ColorFlag change", colorFlag)

		machine.LED_RED.Set((colorFlag & 1) > 0)
		machine.LED_GREEN.Set((colorFlag & 2) > 0)
		machine.LED_BLUE.Set((colorFlag & 4) > 0)
	}

}

func hw_init() {

	// SYSCFGEN is NEEDED FOR IRQ HANDLERS (button + Dio) .. Do not remove
	stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SYSCFGEN)

	// BUTTON PB14
	machine.BUTTON.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	stm32.SYSCFG.EXTICR4.ReplaceBits(0b001, 0xf, stm32.SYSCFG_EXTICR4_EXTI14_Pos) // Enable PORTB On line 14
	//	stm32.EXTI.RTSR.SetBits(stm32.EXTI_RTSR_RT14)                                      // Detect Rising Edge of EXTI14 Line
	stm32.EXTI.FTSR.SetBits(stm32.EXTI_FTSR_FT14) // Detect Falling Edge of EXTI14 Line
	stm32.EXTI.IMR.SetBits(stm32.EXTI_IMR_IM14)   // Enable EXTI14 line

	// Led GPIO init
	machine.LED_RED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LED_GREEN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LED_BLUE.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Switch off LEDS
	machine.LED_RED.Low()
	machine.LED_GREEN.Low()
	machine.LED_BLUE.Low()

	// Enable UART
	uartConsole := &machine.UART0
	uartConsole.Configure(machine.UARTConfig{TX: machine.UART_TX_PIN, RX: machine.UART_TX_PIN, BaudRate: 9600})

	// Enable interrupts (For handling button interrupts)
	intr := interrupt.New(stm32.IRQ_EXTI4_15, gpios_int)
	intr.SetPriority(0x0)
	intr.Enable()

}

//----------------------------------------------------------------------------------------------//

func main() {

	println("*** LGT92 gpio interrupt demo 1  ***")
	println("Press button to change LED color")

	hw_init()

	colorFlag = 1
	for {
		time.Sleep(1 * time.Second)
	}

}
