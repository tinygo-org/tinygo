//go:build py32 && py32f002bxx

package machine

func (p Pin) SetAltFunc(af uint8) {
	port, pin := p.getPort()
	port.AFRL.ReplaceBits(uint32(af), 0xF, (pin%8)*4)
}
