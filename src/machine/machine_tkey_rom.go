//go:build tkey

package machine

/*
	#define TK1_MMIO_TK1_BLAKE2S 0xff000040

	typedef unsigned char uint8_t;
	typedef unsigned long uint32_t;
	typedef unsigned long size_t;

	// blake2s state context
	typedef struct {
		uint8_t b[64]; // input buffer
		uint32_t h[8]; // chained state
		uint32_t t[2]; // total number of bytes
		size_t c;      // pointer for b[]
		size_t outlen; // digest size
	} blake2s_ctx;

	typedef int (*fw_blake2s_p)(void *out, unsigned long outlen, const void *key,
		unsigned long keylen, const void *in,
		unsigned long inlen, blake2s_ctx *ctx);

	int blake2s(void *out, unsigned long outlen, const void *key, unsigned long keylen, const void *in, unsigned long inlen)
	{
		fw_blake2s_p const fw_blake2s =
	    	(fw_blake2s_p) * (volatile uint32_t *)TK1_MMIO_TK1_BLAKE2S;
		blake2s_ctx ctx;

		return fw_blake2s(out, outlen, key, keylen, in, inlen, &ctx);
	}
*/
import "C"
import (
	"errors"
	"unsafe"
)

var (
	ErrBLAKE2sInvalid = errors.New("invalid params for call to BLAKE2s")
	ErrBLAKE2sFailed  = errors.New("call to BLAKE2s failed")
)

func BLAKE2s(output []byte, key []byte, input []byte) error {
	if len(output) == 0 || len(input) == 0 {
		return ErrBLAKE2sInvalid
	}

	op := unsafe.Pointer(&output[0])
	kp := unsafe.Pointer(&key[0])
	ip := unsafe.Pointer(&input[0])

	if res := C.blake2s(op, C.size_t(len(output)), kp, C.size_t(len(key)), ip, C.size_t(len(input))); res != 0 {
		return ErrBLAKE2sFailed
	}

	return nil
}
