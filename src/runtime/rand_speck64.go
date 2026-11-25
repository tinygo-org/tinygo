//go:build linux && !baremetal && !nintendoswitch && !wasm_unknown && !wasip2 && (386 || arm || mips || mipsle)

package runtime

import "math/bits"

// speckRand is Speck64/128 CTR (https://eprint.iacr.org/2013/404), faster than
// Speck128 on 32-bit CPUs. Like Speck128 it never rekeys.
type speckRand struct {
	k   [27]uint32
	ctr uint64
}

func speck64Round(x, y, k uint32) (uint32, uint32) {
	x = (bits.RotateLeft32(x, -8) + y) ^ k
	y = bits.RotateLeft32(y, 3) ^ x
	return x, y
}

func (s *speckRand) init(k0, k1 uint64) {
	a, b, c, d := uint32(k0), uint32(k0>>32), uint32(k1), uint32(k1>>32)
	for i := uint32(0); i < uint32(len(s.k)); {
		s.k[i] = a
		b, a = speck64Round(b, a, i)
		i++
		s.k[i] = a
		c, a = speck64Round(c, a, i)
		i++
		s.k[i] = a
		d, a = speck64Round(d, a, i)
		i++
	}
	s.ctr = 0
}

func (s *speckRand) next() uint64 {
	y, x := uint32(s.ctr), uint32(s.ctr>>32)
	s.ctr++
	for i := range s.k {
		x, y = speck64Round(x, y, s.k[i])
	}
	return uint64(x)<<32 | uint64(y)
}
