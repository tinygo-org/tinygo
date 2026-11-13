// Package hmacdrbg implements HMAC_DRBG with SHA-256 from NIST SP 800-90A Rev. 1 section 10.1.2.
// It has no reseed function and no prediction resistance.
package hmacdrbg

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

// Limits from NIST SP 800-90A Rev. 1 section 10.1, table 2.
const (
	reseedInterval = 1 << 48
	MaxRequest     = 1 << 16
)

var (
	ErrReseed  = errors.New("hmacdrbg: reseed interval reached")
	ErrRequest = errors.New("hmacdrbg: request too large")
)

type DRBG struct {
	k, v          [sha256.Size]byte
	reseedCounter uint64
}

// New instantiates the DRBG. Entropy should hold at least 32 bytes of entropy and the nonce at least 16 bytes.
func New(entropy, nonce, personalization []byte) *DRBG {
	d := &DRBG{}
	for i := range d.k {
		d.k[i] = 0x00
		d.v[i] = 0x01
	}
	d.update(entropy, nonce, personalization)
	d.reseedCounter = 1
	return d
}

func (d *DRBG) update(provided ...[]byte) {
	d.updateRound(0x00, provided)
	n := 0
	for _, p := range provided {
		n += len(p)
	}
	if n == 0 {
		return
	}
	d.updateRound(0x01, provided)
}

func (d *DRBG) updateRound(round byte, provided [][]byte) {
	m := hmac.New(sha256.New, d.k[:])
	m.Write(d.v[:])
	m.Write([]byte{round})
	for _, p := range provided {
		m.Write(p)
	}
	m.Sum(d.k[:0])
	m = hmac.New(sha256.New, d.k[:])
	m.Write(d.v[:])
	m.Sum(d.v[:0])
}

// Generate fills out with random bytes. Additional input is optional.
func (d *DRBG) Generate(out, additional []byte) error {
	if len(out) > MaxRequest {
		return ErrRequest
	}
	if d.reseedCounter > reseedInterval {
		return ErrReseed
	}
	if len(additional) != 0 {
		d.update(additional)
	}
	m := hmac.New(sha256.New, d.k[:])
	for len(out) > 0 {
		m.Reset()
		m.Write(d.v[:])
		m.Sum(d.v[:0])
		out = out[copy(out, d.v[:]):]
	}
	d.update(additional)
	d.reseedCounter++
	return nil
}
