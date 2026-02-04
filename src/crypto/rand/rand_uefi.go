//go:build uefi

package rand

import (
	"errors"
	"machine/uefi"
)

func init() {
	Reader = &reader{}
}

type reader struct {
}

func (r *reader) Read(b []byte) (n int, err error) {
	if !uefi.HasRNGSupport() {
		return 0, errors.New("no hardware rng available")
	} else if len(b) == 0 {
		return 0, nil
	}

	var randomByte uint64
	for i := range b {
		if i%8 == 0 {
			var ok bool
			randomByte, ok = uefi.ReadRandom()
			if !ok {
				return n, errors.New("no random seed available")
			}
		} else {
			randomByte >>= 8
		}
		b[i] = byte(randomByte)
	}

	return len(b), nil
}
