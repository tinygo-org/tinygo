package main

import (
	"slices"
	"strings"
)

func tweakDevice(d *Device, pkgName string) {
	if pkgName != "stm32" {
		// no-op for device types that do not need tweaks
		return
	}

	// Source file machine_stm32_iwdg.go relies on the presence of
	// a register IWDG. On some devices, though, like the h723,
	// there are two registers, IWDG1 and IWDG2. In this case we
	// define an alias IWDG for IWDG1.
	addUnnumberedAlias(d, "IWDG", "IWDG1")

	for _, p := range d.Peripherals {
		switch p.GroupName {
		case "TIM":
			// SVDs like stm32l4r5.svd define CCMR*_Input and _Output
			// alternate registers, with _Input sorted before _Output.
			// This would result in the _Output fields missing from the
			// TIM_type struct definition, hence compilation would fail.
			// Therefore we adjust the order of these alternate registers
			// accordingly.
			stm32EnsureCCMROrder(p.Registers)

		case "USART":
			isr := p.lookupRegister("ISR")
			if isr == nil {
				continue
			}

			// Some of the upstream SVD files, like the one for stm32wl5x_cm4,
			// lack FIFO enabled variants of the USART ISR register,
			// even if the register manual defines them. To make sure
			// that TXFNF is not missing from the generated .go files,
			// we add TXFNF here in case FIFOEN is present.
			if p.lookupRegister("CR1").hasBitfield("FIFOEN") {
				stm32EnsureBit(isr, "TXFNF", "TXE", "USART_ISR_")
			}

			// Svdtools handles the presence of alternate USART ISR registers,
			// like in case of the stm32l4r5, adjusting names like "ISR_enabled"
			// to "ISR", deleting "ISR_disabled" or "ISR_ALTERNATE" register definitions
			// from the SVD.
			// As this would result in USART_ISR_TXE definitions missing in the
			// generated .go file, a constant for TXE is added here
			// in case TXFNF is defined.
			stm32EnsureBit(isr, "TXE", "TXFNF", "USART_ISR_")
		}
	}
}

func addUnnumberedAlias(d *Device, dest, src string) {
	if _, ok := d.PeripheralDict[dest]; !ok {
		if p := d.PeripheralDict[src]; p != nil {
			p.Alias = dest
		}
	}
}

func stm32EnsureCCMROrder(registers []*PeripheralField) {
	for i, r := range registers {
		if i > 0 {
			prev := registers[i-1]
			if r.Address == prev.Address {
				// alternate field
				if strings.HasPrefix(prev.Name, "CCMR") && strings.HasPrefix(r.Name, "CCMR") && strings.HasSuffix(r.Name, "_Output") {
					// swap register pointers
					registers[i-1], registers[i] = r, prev
				}
			}
		}
	}
}

func stm32EnsureBit(reg *PeripheralField, want, have, prefix string) {
	iWant := -1
	iHave := -1
	wantConst := prefix + want
	haveConst := prefix + have
	for i := range reg.Constants {
		f := &reg.Constants[i]
		if f.Name == wantConst {
			iWant = i
			break
		}
		if f.Name == haveConst {
			iHave = i
			break
		}
	}
	if iHave != -1 && iWant == -1 {
		iWant = iHave + 1
		reg.Constants = slices.Insert(reg.Constants, iWant, reg.Constants[iHave])
		reg.Constants[iWant].Name = wantConst
		reg.Constants[iWant].Description = "Bit " + want + ". (added by gen-device-svd)"
	}
}

func (p *Peripheral) lookupRegister(name string) *PeripheralField {
	for _, r := range p.Registers {
		if r.Name == name {
			return r
		}
	}
	return nil
}

func (r *PeripheralField) hasBitfield(name string) bool {
	if r == nil {
		return false
	}
	for i := range r.Bitfields {
		if r.Bitfields[i].Name == name {
			return true
		}
	}
	return false
}
