// CMSIS abstraction functions.
//
// Original copyright:
//
//     Copyright (c) 2009 - 2015 ARM LIMITED
//
//     All rights reserved.
//     Redistribution and use in source and binary forms, with or without
//     modification, are permitted provided that the following conditions are met:
//     - Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     - Redistributions in binary form must reproduce the above copyright
//       notice, this list of conditions and the following disclaimer in the
//       documentation and/or other materials provided with the distribution.
//     - Neither the name of ARM nor the names of its contributors may be used
//       to endorse or promote products derived from this software without
//       specific prior written permission.
//
//     THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
//     AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
//     IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
//     ARE DISCLAIMED. IN NO EVENT SHALL COPYRIGHT HOLDERS AND CONTRIBUTORS BE
//     LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
//     CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
//     SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
//     INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
//     CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
//     ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
//     POSSIBILITY OF SUCH DAMAGE.
package arm

type __reg uint32
type RegValue = __reg

type __asm string

func Asm(s __asm)

const (
	SCS_BASE  = 0xE000E000
	NVIC_BASE = SCS_BASE + 0x0100
)

// Nested Vectored Interrupt Controller (NVIC).
var NVIC = struct {
	ISER [8]__reg
}{
	ISER: [8]__reg{
		NVIC_BASE + 0x000,
		NVIC_BASE + 0x004,
		NVIC_BASE + 0x008,
		NVIC_BASE + 0x00C,
		NVIC_BASE + 0x010,
		NVIC_BASE + 0x014,
		NVIC_BASE + 0x018,
		NVIC_BASE + 0x01C,
	},
}

// Enable the given interrupt number.
func EnableIRQ(irq uint32) {
	NVIC.ISER[irq >> 5] = 1 << (irq & 0x1F)
}
