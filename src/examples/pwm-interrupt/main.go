//go:build rp2040

package main

import (
	"device/rp"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"time"
)

var (
	wraps0 volatile.Register32
	wraps1 volatile.Register32
	ch0    uint8
	ch1    uint8
)

func handlePWM(interrupt.Interrupt) {
	if machine.PWM0.IsIRQSet() {
		machine.PWM0.ClearIRQ()
		n := wraps0.Get() + 1
		wraps0.Set(n)
		machine.GP4.Set(n&1 != 0)
		duty := machine.PWM0.Top() / 4
		if n&8 != 0 {
			duty *= 3
		}
		machine.PWM0.Set(ch0, duty)
	}
	if machine.PWM1.IsIRQSet() {
		machine.PWM1.ClearIRQ()
		n := wraps1.Get() + 1
		wraps1.Set(n)
		machine.GP5.Set(n&1 != 0)
		duty := machine.PWM1.Top() / 4
		if n&8 != 0 {
			duty *= 3
		}
		machine.PWM1.Set(ch1, duty)
	}
}

func main() {
	machine.GP4.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.GP5.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.GP4.Low()
	machine.GP5.Low()

	if err := machine.PWM0.Configure(machine.PWMConfig{Period: 1_000_000}); err != nil {
		panic(err)
	}
	var err error
	ch0, err = machine.PWM0.Channel(machine.GP0)
	if err != nil {
		panic(err)
	}
	if err = machine.PWM1.Configure(machine.PWMConfig{Period: 500_000}); err != nil {
		panic(err)
	}
	ch1, err = machine.PWM1.Channel(machine.GP2)
	if err != nil {
		panic(err)
	}
	machine.PWM0.Set(ch0, machine.PWM0.Top()/4)
	machine.PWM1.Set(ch1, machine.PWM1.Top()/4)

	irq := interrupt.New(rp.IRQ_PWM_IRQ_WRAP, handlePWM)
	machine.PWM0.ClearIRQ()
	machine.PWM1.ClearIRQ()
	machine.PWM0.SetIRQ(true)
	machine.PWM1.SetIRQ(true)
	irq.Enable()

	for {
		time.Sleep(time.Second)
		state := interrupt.Disable()
		n0, n1 := wraps0.Get(), wraps1.Get()
		interrupt.Restore(state)
		println("PWM wraps:", n0, n1)
	}
}
