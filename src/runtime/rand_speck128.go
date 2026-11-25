//go:build darwin || (linux && !baremetal && !nintendoswitch && !wasm_unknown && !wasip2 && !(386 || arm || mips || mipsle))

package runtime

import "math/bits"

// speckRand is Speck128/128 CTR (https://eprint.iacr.org/2013/404). It never rekeys, so it
// has no forward secrecy. Rekeying every 16 KB or 4 KB would cost about 3% or 5%.
type speckRand struct {
	k    [32]uint64
	ctr  uint64
	hi   uint64
	have bool
}

func (s *speckRand) init(k0, l0 uint64) {
	s.k[0] = k0
	l := l0
	for i := uint64(0); i < uint64(len(s.k)-1); i++ {
		l = (s.k[i] + bits.RotateLeft64(l, -8)) ^ i
		s.k[i+1] = bits.RotateLeft64(s.k[i], 3) ^ l
	}
	s.ctr = 0
	s.have = false
}

func (s *speckRand) next() uint64 {
	if s.have {
		s.have = false
		return s.hi
	}
	y, x := s.ctr, uint64(0)
	s.ctr++
	for i := range s.k {
		x = (bits.RotateLeft64(x, -8) + y) ^ s.k[i]
		y = bits.RotateLeft64(y, 3) ^ x
	}
	s.hi = x
	s.have = true
	return y
}
