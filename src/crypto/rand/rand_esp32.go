//go:build esp32

package rand

import (
	"crypto/internal/hmacdrbg"
	"errors"
	"machine"
	"sync"
)

// seedSize must match entropySeedSize in src/machine/machine_esp32_rng.go.
// The first 96 bytes are the entropy input, the last 32 bytes the nonce.
const seedSize = 128

var errNoSeed = errors.New("crypto/rand: no ESP32 startup entropy")

// takeEntropySeed is implemented in src/machine/machine_esp32_rng.go.
func takeEntropySeed(dst []byte) bool

func init() {
	var seed [seedSize]byte
	if !takeEntropySeed(seed[:]) {
		Reader = &reader{err: errNoSeed}
		return
	}
	Reader = &reader{drbg: hmacdrbg.New(seed[:96], seed[96:], []byte("TinyGo ESP32 crypto/rand"))}
	clear(seed[:])
}

// reader follows the ESP-IDF advice to seed a DRBG from the hardware RNG, see
// https://docs.espressif.com/projects/esp-idf/en/latest/esp32/api-reference/system/random.html
type reader struct {
	mu   sync.Mutex
	drbg *hmacdrbg.DRBG
	err  error
}

func (r *reader) Read(b []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return 0, r.err
	}
	for n < len(b) {
		chunk := min(len(b)-n, hmacdrbg.MaxRequest)
		// Fresh hardware output is extra input. It adds entropy while Wi-Fi or Bluetooth is on.
		var extra [8]byte
		for i := 0; i < len(extra); i += 4 {
			v, _ := machine.GetRNG()
			extra[i], extra[i+1], extra[i+2], extra[i+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
		}
		if err = r.drbg.Generate(b[n:n+chunk], extra[:]); err != nil {
			return n, err
		}
		n += chunk
	}
	return n, nil
}
