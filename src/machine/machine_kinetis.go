// +build nxp,kinetis

package machine

import (
	"device/nxp"
	"runtime/volatile"
)

type PinMux struct {
	Pin Pin
	Mux uint32
}

type PeripheralConfig struct {
	SCGC     *volatile.Register32
	SCGCMask uint32
	Pins     []PinMux
}

//go:inline
func (c *PeripheralConfig) Enabled() bool {
	return c.SCGC.HasBits(c.SCGCMask)
}

//go:inline
func (c *PeripheralConfig) Enable() {
	c.SCGC.SetBits(c.SCGCMask)
}

//go:inline
func (c *PeripheralConfig) Disable() {
	c.SCGC.ClearBits(c.SCGCMask)
}

//go:inline
func (p *PinMux) EnableWith(flags uint32) {
	p.Pin.Control().Set(flags | (p.Mux << nxp.PORT_PCR0_MUX_Pos))
}
