//go:build !baremetal

package machine

import (
	"crypto/rand"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRandomUint32(t *testing.T) {
	reader := rand.Reader
	defer func() {
		rand.Reader = reader
	}()

	for _, api := range []struct {
		name string
		read func() (uint32, error)
	}{
		{"RandomUint32", RandomUint32},
		{"GetRNG", GetRNG},
	} {
		t.Run(api.name, func(t *testing.T) {
			rand.Reader = strings.NewReader("\x78\x56\x34\x12\xef\xcd\xab\x90")
			for _, want := range []uint32{0x12345678, 0x90abcdef} {
				got, err := api.read()
				if got != want || err != nil {
					t.Fatalf("got (%#x, %v), want (%#x, nil)", got, err, want)
				}
			}

			got, err := api.read()
			if got != 0 || err != io.EOF {
				t.Fatalf("got (%#x, %v), want (0, EOF)", got, err)
			}

			rand.Reader = strings.NewReader("\x78\x56")
			got, err = api.read()
			if got != 0 || err != io.ErrUnexpectedEOF {
				t.Fatalf("got (%#x, %v), want (0, unexpected EOF)", got, err)
			}

			wantErr := errors.New("entropy source failed")
			rand.Reader = &randomErrorReader{err: wantErr}
			got, err = api.read()
			if got != 0 || err != wantErr {
				t.Fatalf("got (%#x, %v), want (0, %v)", got, err, wantErr)
			}
		})
	}
}

type randomErrorReader struct {
	err error
}

func (r *randomErrorReader) Read([]byte) (int, error) {
	return 0, r.err
}
